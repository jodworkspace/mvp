package main

import (
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/handler/rest"
	v1 "gitlab.com/gookie/mvp/internal/handler/rest/v1"
	postgresrepo "gitlab.com/gookie/mvp/internal/repository/postgres"
	"gitlab.com/gookie/mvp/internal/usecase/oauth"
	taskuc "gitlab.com/gookie/mvp/internal/usecase/task"
	useruc "gitlab.com/gookie/mvp/internal/usecase/user"
	"gitlab.com/gookie/mvp/pkg/db"
	"gitlab.com/gookie/mvp/pkg/logger"
)

func main() {
	cfg := config.LoadConfig()
	zl := logger.MustNewZapLogger(cfg.Logger.Level)

	pgConn := db.MustNewPostgresConnection(
		cfg.Postgres.DSN(),
		db.WithMaxConns(10),
		db.WithMinConns(2))

	taskRepository := postgresrepo.NewTaskRepository(pgConn)
	taskUC := taskuc.NewTaskUsecase(taskRepository, zl)
	taskHandler := v1.NewTaskHandler(taskUC, zl)

	userRepository := postgresrepo.NewUserRepository(pgConn)
	federatedUserRepository := postgresrepo.NewFederatedUserRepository(pgConn)

	userUC := useruc.NewUserUsecase(userRepository, zl)
	oauthUC := oauth.NewManager(cfg.JWT, zl)
	googleUC := oauth.NewGoogleUseCase(cfg.GoogleOAuth, zl)

	oauthUC.RegisterOAuthProvider(googleUC)

	oauthHandler := v1.NewOAuthHandler(userUC, zl)

	srv := rest.NewServer(cfg, taskHandler, oauthHandler, zl)
	srv.Run()
}
