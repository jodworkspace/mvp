package main

import (
	"gitlab.com/tokpok/mvp/config"
	"gitlab.com/tokpok/mvp/internal/handler/rest"
	"gitlab.com/tokpok/mvp/pkg/db"
	"gitlab.com/tokpok/mvp/pkg/logger"
	"log"
)

func main() {
	cfg := config.NewConfig()
	pgdb, err := db.NewPostgresConnection(cfg, db.PostgresOptions{})
	if err != nil {
		log.Fatal(err)
	}
	zl, err := logger.NewZapLogger(cfg.Logger.Level)
	if err != nil {
		log.Fatal(err)
	}

	srv := rest.NewServer(cfg, pgdb, zl)
	srv.Run()
}
