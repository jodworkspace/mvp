package main

import (
	"github.com/urfave/cli/v2"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/handler/rest"
	v1 "gitlab.com/jodworkspace/mvp/internal/handler/rest/v1"
	postgresrepo "gitlab.com/jodworkspace/mvp/internal/repository/postgres"
	authuc "gitlab.com/jodworkspace/mvp/internal/usecase/auth"
	"gitlab.com/jodworkspace/mvp/internal/usecase/oauth"
	taskuc "gitlab.com/jodworkspace/mvp/internal/usecase/task"
	useruc "gitlab.com/jodworkspace/mvp/internal/usecase/user"
	"gitlab.com/jodworkspace/mvp/pkg/db/postgres"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"
	"log"
	"os"
)

func main() {
	cfg := config.LoadConfig()
	zapLogger := logger.MustNewLogger(cfg.Logger.Level)
	aead := cipherx.MustNewAEAD([]byte(cfg.Server.AESKey))

	pgConn := postgres.MustNewPostgresConnection(
		cfg.Postgres.DSN(),
		postgres.WithMaxConns(10),
		postgres.WithMinConns(2),
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
