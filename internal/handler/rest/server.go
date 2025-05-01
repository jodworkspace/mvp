package rest

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/handler/rest/middleware"
	v1 "gitlab.com/gookie/mvp/internal/handler/rest/v1"
	"gitlab.com/gookie/mvp/pkg/logger"
	"log"
	"net/http"
	"os"
	"os/signal"
)

type Server struct {
	cfg          *config.Config
	taskHandler  *v1.TaskHandler
	oauthHandler *v1.OAuthHandler
	logger       *logger.ZapLogger
}

func NewServer(
	cfg *config.Config,
	taskHandler *v1.TaskHandler,
	oauthHandler *v1.OAuthHandler,
	logger *logger.ZapLogger,
) *Server {
	return &Server{
		cfg:          cfg,
		taskHandler:  taskHandler,
		oauthHandler: oauthHandler,
		logger:       logger,
	}
}

func (s *Server) Run() {
	addr := fmt.Sprintf("%s:%s", s.cfg.Server.Host, s.cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        s.RestMux(),
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

func (s *Server) RestMux() *chi.Mux {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Use(middleware.CORS)

	r.Get("/api/v1/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(map[string]any{
			"status": "ok",
		})

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	r.Get("/api/v1/userinfo", s.oauthHandler.GetUserInfo)
	r.Post("/api/v1/oauth/authorize", s.oauthHandler.Authorize)
	r.With(middleware.Validate(&v1.TokenRequest{})).Post("/api/v1/oauth/token", s.oauthHandler.ExchangeToken)

	r.Get("/api/v1/tasks", s.taskHandler.List)
	r.Post("/api/v1/tasks", s.taskHandler.Create)

	r.NotFound(Return404)
	return r
}

func Return404(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Route not found", http.StatusNotFound)
}
