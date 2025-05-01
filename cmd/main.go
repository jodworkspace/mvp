package main

import (
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/handler/rest"
	v1 "gitlab.com/gookie/mvp/internal/handler/rest/v1"
	postgresrepo "gitlab.com/gookie/mvp/internal/repository/postgres"
	taskuc "gitlab.com/gookie/mvp/internal/usecase/task"
	useruc "gitlab.com/gookie/mvp/internal/usecase/user"
	"gitlab.com/gookie/mvp/pkg/db"
	"gitlab.com/gookie/mvp/pkg/httpclient"
	"gitlab.com/gookie/mvp/pkg/logger"
	"net/http"
)

func main() {
	cfg := config.NewConfig()
	zl := logger.MustNewZapLogger(cfg.Logger.Level)

	pgConn := db.MustNewPostgresConnection(cfg.Postgres.DSN(), db.WithMaxConns(10))
	httpClient := httpclient.NewHTTPClient(&http.Client{})

	taskRepository := postgresrepo.NewTaskRepository(pgConn)
	taskUsecase := taskuc.NewTaskUsecase(taskRepository, zl)
	taskHandler := v1.NewTaskHandler(taskUsecase, zl)

	userRepository := postgresrepo.NewUserRepository(pgConn)
	userUsecase := useruc.NewUserUsecase(userRepository, zl)
	oauthHandler := v1.NewOAuthHandler(userUsecase, httpClient, cfg, zl)

	srv := rest.NewServer(cfg, taskHandler, oauthHandler, zl)
	srv.Run()
}
