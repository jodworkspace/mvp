package main

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/gorilla/sessions"
	goredis "github.com/redis/go-redis/v9"
	"github.com/urfave/cli/v2"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/internal/handler/rest"
	v1 "gitlab.com/jodworkspace/mvp/internal/handler/rest/v1"
	"gitlab.com/jodworkspace/mvp/internal/instrument"
	postgresrepo "gitlab.com/jodworkspace/mvp/internal/repository/postgres"
	"gitlab.com/jodworkspace/mvp/internal/usecase/oauth"
	taskuc "gitlab.com/jodworkspace/mvp/internal/usecase/task"
	useruc "gitlab.com/jodworkspace/mvp/internal/usecase/user"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
	"gitlab.com/jodworkspace/mvp/pkg/db/redis"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/monitor/metrics"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	initGob()
	cfg := config.LoadConfig()
	ctx := context.Background()

	conn, err := grpc.NewClient(cfg.Monitor.CollectorEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	metricsManager := metrics.NewManager(cfg.Monitor.ServiceName, conn)
	shutdown, err := metricsManager.Init(ctx)
	defer shutdown(ctx)

	dbMetrics, err := metricsManager.NewPgxMetrics()
	if err != nil {
		panic(err)
	}

	httpMetrics, err := metricsManager.NewHTTPMetrics()
	if err != nil {
		panic(err)
	}

	zapLogger := logger.MustNewLogger(cfg.Logger.Level)
	aead := cipherx.MustNewAEAD([]byte(cfg.Server.AESKey))

	pgClient, err := postgres.NewPostgresConnection(
		cfg.Postgres.DSN(),
		postgres.WithMaxConns(10),
		postgres.WithMinConns(2),
	)
	if err != nil {
		panic(err)
	}
	defer pgClient.Pool().Close()

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
		Path:     cfg.Session.CookiePath,
		Domain:   cfg.Session.Domain,
		MaxAge:   cfg.Session.MaxAge,
		HttpOnly: cfg.Session.HTTPOnly,
		Secure:   cfg.Session.Secure,
	})

	app := &cli.App{
		Name:  "Jod",
		Usage: "MVP Server",
		Action: func(c *cli.Context) error {
			pgInstrumentedClient := instrument.NewInstrumentedClient(pgClient, dbMetrics)

			transactionManager := postgresrepo.NewTransactionManager(pgClient)

			taskRepository := postgresrepo.NewTaskRepository(pgInstrumentedClient)
			taskUC := taskuc.NewTaskUseCase(taskRepository, zapLogger)
			taskHandler := v1.NewTaskHandler(taskUC, zapLogger)

			userRepository := postgresrepo.NewUserRepository(pgClient)
			linkRepository := postgresrepo.NewLinkRepository(pgClient)
			userUC := useruc.NewUserUseCase(userRepository, linkRepository, transactionManager, aead, zapLogger)

			googleUC := oauthuc.NewGoogleUseCase(cfg.GoogleOAuth, zapLogger)
			oauthMng := oauthuc.NewManager(cfg.Token, zapLogger)
			oauthMng.RegisterOAuthProvider(googleUC)

			oauthHandler := v1.NewOAuthHandler(sessionStore, userUC, oauthMng, zapLogger)

			srv := rest.NewServer(cfg, sessionStore, taskHandler, oauthHandler, zapLogger, metricsManager, httpMetrics)
			srv.Run()

			return nil
		},
	}

	err = app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func initGob() {
	gob.Register(&domain.User{})
	gob.Register(&domain.Document{})
}
