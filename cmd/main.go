package main

import (
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/handler/rest"
	"gitlab.com/gookie/mvp/pkg/db"
	"gitlab.com/gookie/mvp/pkg/logger"
)

func main() {
	cfg := config.NewConfig()
	pgConn := db.MustNewPostgresConnection(cfg.Postgres.DSN(), db.WithMaxConns(10))
	zl := logger.MustNewZapLogger(cfg.Logger.Level)

	srv := rest.NewServer(cfg, pgConn, zl)
	srv.Run()
}
