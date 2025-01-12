package rest

import (
	"fmt"
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/pkg/db"
	"gitlab.com/gookie/mvp/pkg/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type Server struct {
	cfg    *config.Config
	db     db.PostgresConn
	logger *logger.ZapLogger
}

func NewServer(
	cfg *config.Config,
	db db.PostgresConn,
	logger *logger.ZapLogger,
) *Server {
	return &Server{
		cfg:    cfg,
		db:     db,
		logger: logger,
	}
}

func (s *Server) Run() {
	addr := fmt.Sprintf("%s:%s", s.cfg.Server.Host, s.cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        s.RestHandlersRoute(),
		MaxHeaderBytes: 1 << 20,
	}

	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("Server started on " + addr)
		err := srv.ListenAndServe() // blocking call
		if err != nil {
			errChan <- err
		}
	}()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt, os.Kill)
	for {
		select {
		case <-interruptChan:
			log.Fatal("Received shutdown signal!")
		case err := <-errChan:
			log.Fatal("Server error:", err)
		}
	}
}
