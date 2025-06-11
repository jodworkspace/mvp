package main

import (
	"github.com/urfave/cli/v2"
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/handler/rest"
	v1 "gitlab.com/gookie/mvp/internal/handler/rest/v1"
	postgresrepo "gitlab.com/gookie/mvp/internal/repository/postgres"
	authuc "gitlab.com/gookie/mvp/internal/usecase/auth"
	"gitlab.com/gookie/mvp/internal/usecase/oauth"
	taskuc "gitlab.com/gookie/mvp/internal/usecase/task"
	useruc "gitlab.com/gookie/mvp/internal/usecase/user"
	"gitlab.com/gookie/mvp/pkg/db"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/cipherx"
	"log"
	"os"
)

func main() {
	cfg := config.LoadConfig()
	zapLogger := logger.MustNewZapLogger(cfg.Logger.Level)
	aead := cipherx.MustNewAEAD([]byte(cfg.Server.AESKey))

	pgConn := db.MustNewPostgresConnection(
		cfg.Postgres.DSN(),
		db.WithMaxConns(10),
		db.WithMinConns(2),
	)
	defer pgConn.Pool().Close()

	app := &cli.App{
		Name:  "Gookie",
		Usage: "MVP server",
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

			authUC := authuc.NewUseCase(cfg.Token, zapLogger)

			oauthHandler := v1.NewOAuthHandler(userUC, oauthMng, authUC, zapLogger)

			srv := rest.NewServer(cfg, taskHandler, oauthHandler, zapLogger)
			srv.Run()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
