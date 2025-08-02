package main

import (
	"encoding/gob"
	"fmt"
	"github.com/gorilla/sessions"
	goredis "github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/internal/handler/rest"
	v1 "gitlab.com/jodworkspace/mvp/internal/handler/rest/v1"
	postgresrepo "gitlab.com/jodworkspace/mvp/internal/repository/postgres"
	"gitlab.com/jodworkspace/mvp/internal/usecase/oauth"
	taskuc "gitlab.com/jodworkspace/mvp/internal/usecase/task"
	useruc "gitlab.com/jodworkspace/mvp/internal/usecase/user"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
	"gitlab.com/jodworkspace/mvp/pkg/db/redis"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"
	"log"
	"os"
)

func main() {
	initGob()
	cfg := config.LoadConfig()
	zapLogger := logger.MustNewLogger(cfg.Logger.Level)
	aead := cipherx.MustNewAEAD([]byte(cfg.Server.AESKey))

	pgConn := postgres.MustNewPostgresConnection(
		cfg.Postgres.DSN(),
		postgres.WithMaxConns(10),
		postgres.WithMinConns(2),
	)
	defer pgConn.Pool().Close()

	redisClient, err := redis.NewClient(&goredis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Username: cfg.Redis.Username,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		panic(err)
	}
	defer redisClient.Close()

	sessionStore := redis.NewStore(redisClient, "session:", &sessions.Options{
		Path:     cfg.SessionConfig.Path,
		Domain:   cfg.SessionConfig.Domain,
		MaxAge:   cfg.SessionConfig.MaxAge,
		HttpOnly: cfg.SessionConfig.HTTPOnly,
		Secure:   cfg.SessionConfig.Secure,
	})

	app := &cli.App{
		Name:  "Jod",
		Usage: "MVP Server",
		Action: func(c *cli.Context) error {

			transactionManager := postgresrepo.NewTransactionManager(pgConn)

			taskRepository := postgresrepo.NewTaskRepository(pgConn)
			taskUC := taskuc.NewTaskUsecase(taskRepository, zapLogger)
			taskHandler := v1.NewTaskHandler(taskUC, zapLogger)

			userRepository := postgresrepo.NewUserRepository(pgConn)
			linkRepository := postgresrepo.NewLinkRepository(pgConn)
			userUC := useruc.NewUserUseCase(userRepository, linkRepository, transactionManager, aead, zapLogger)

			googleUC := oauthuc.NewGoogleUseCase(cfg.GoogleOAuth, zapLogger)
			oauthMng := oauthuc.NewManager(cfg.Token, zapLogger)
			oauthMng.RegisterOAuthProvider(googleUC)

			oauthHandler := v1.NewOAuthHandler(sessionStore, userUC, oauthMng, zapLogger)

			srv := rest.NewServer(cfg, sessionStore, taskHandler, oauthHandler, zapLogger)
			srv.Run()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func initGob() {
	gob.Register(&domain.User{})
	gob.Register(&domain.Document{})
}
