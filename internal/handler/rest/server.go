package rest

import (
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"gitlab.com/gookie/mvp/config"
	"gitlab.com/gookie/mvp/internal/domain"
	mw "gitlab.com/gookie/mvp/internal/handler/rest/middleware"
	v1 "gitlab.com/gookie/mvp/internal/handler/rest/v1"
	"gitlab.com/gookie/mvp/pkg/logger"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
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

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.CORS.AllowedOrigins,
		AllowedMethods:   s.cfg.CORS.AllowedMethods,
		AllowedHeaders:   s.cfg.CORS.AllowedHeaders,
		AllowCredentials: s.cfg.CORS.AllowCredentials,
		MaxAge:           300,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(httpx.JSON{"status": "ok"})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	r.Post("/api/v1/login/google", s.oauthHandler.Login(domain.ProviderGoogle))
	r.Post("/api/v1/login/github", s.oauthHandler.Login(domain.ProviderGitHub))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(mw.ValidateToken([]byte(s.cfg.Token.Secret)))

		r.Get("/userinfo", s.oauthHandler.GetUserInfo)

		r.With(mw.Pagination).Get("/tasks", s.taskHandler.List)
		r.Post("/tasks", s.taskHandler.Create)
		r.Get("/tasks/{id}", s.taskHandler.Get)
		r.Put("/tasks/{id}", s.taskHandler.Update)
		r.Delete("/tasks/{id}", s.taskHandler.Delete)
	})

	r.NotFound(Return404)
	return r
}

func Return404(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}
