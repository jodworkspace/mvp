package main

import (
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
)

func main() {
	cfg := config.LoadConfig()
	zapLogger := logger.MustNewZapLogger(cfg.Logger.Level)

	pgConn := db.MustNewPostgresConnection(
		cfg.Postgres.DSN(),
		db.WithMaxConns(10),
		db.WithMinConns(2))

	transactionManager := postgresrepo.NewTransactionManager(pgConn)

	taskRepository := postgresrepo.NewTaskRepository(pgConn)
	taskUC := taskuc.NewTaskUsecase(taskRepository, zapLogger)
	taskHandler := v1.NewTaskHandler(taskUC, zapLogger)

	userRepository := postgresrepo.NewUserRepository(pgConn)
	federatedUserRepository := postgresrepo.NewFederatedUserRepository(pgConn)
	userUC := useruc.NewUserUseCase(userRepository, federatedUserRepository, transactionManager, zapLogger)

	authUC := authuc.NewUseCase(cfg.JWT, zapLogger)
	oauthUC := oauthuc.NewManager(cfg.JWT, zapLogger)
	googleUC := oauthuc.NewGoogleUseCase(cfg.GoogleOAuth, zapLogger)
	oauthUC.RegisterOAuthProvider(googleUC)

	oauthHandler := v1.NewOAuthHandler(userUC, oauthUC, authUC, zapLogger)

	srv := rest.NewServer(cfg, taskHandler, oauthHandler, zapLogger)
	srv.Run()
}
