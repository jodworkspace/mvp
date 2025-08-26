package rest

import (
	"encoding/json"
	"fmt"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"gitlab.com/jodworkspace/mvp/pkg/monitor/metrics"
	"gitlab.com/jodworkspace/mvp/pkg/monitor/metrics/request"

	"log"
	"net/http"
	"os"
	"os/signal"

	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/internal/handler/rest/middleware"
	v1 "gitlab.com/jodworkspace/mvp/internal/handler/rest/v1"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
)

type Server struct {
	cfg            *config.Config
	sessionStore   sessions.Store
	taskHandler    *v1.TaskHandler
	oauthHandler   *v1.OAuthHandler
	logger         *logger.ZapLogger
	metricsManager *metrics.Manager
	httpMetrics    *request.HTTPMetrics
}

func NewServer(
	cfg *config.Config,
	sessionStore sessions.Store,
	taskHandler *v1.TaskHandler,
	oauthHandler *v1.OAuthHandler,
	logger *logger.ZapLogger,
	metricsManager *metrics.Manager,
	httpMetrics *request.HTTPMetrics,
) *Server {

	return &Server{
		cfg:            cfg,
		sessionStore:   sessionStore,
		taskHandler:    taskHandler,
		oauthHandler:   oauthHandler,
		logger:         logger,
		metricsManager: metricsManager,
		httpMetrics:    httpMetrics,
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

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   s.cfg.CORS.AllowedOrigins,
		AllowedMethods:   s.cfg.CORS.AllowedMethods,
		AllowedHeaders:   s.cfg.CORS.AllowedHeaders,
		AllowCredentials: s.cfg.CORS.AllowCredentials,
		MaxAge:           300,
	}))

	ir := s.instrumentedRouter(r)
	r.Handle("/metrics", s.metricsManager.PrometheusHandler())

	ir.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(httpx.JSON{"status": "ok"})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	ir.Post("/api/v1/login/google", s.oauthHandler.Login(domain.ProviderGoogle))
	ir.Post("/api/v1/login/github", s.oauthHandler.Login(domain.ProviderGitHub))

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.SessionAuth(s.sessionStore, domain.SessionCookieName))
		r.Get("/userinfo", s.oauthHandler.GetUserInfo)

		r.With(middleware.Filter).Get("/tasks", s.taskHandler.List)
		r.Post("/tasks", s.taskHandler.Create)
		r.Get("/tasks/{id}", s.taskHandler.Get)
		r.Put("/tasks/{id}", s.taskHandler.Update)
		r.Delete("/tasks/{id}", s.taskHandler.Delete)
	})

	ir.NotFound(NotFoundRoute)
	ir.MethodNotAllowed(NotFoundRoute)
	return r
}

func NotFoundRoute(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) instrumentedRouter(r chi.Router) chi.Router {
	return r.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			if routePattern == "" {
				routePattern = "unmatched"
			}

			s.httpMetrics.Handle(routePattern, next).ServeHTTP(w, r)
		})
	})
}
