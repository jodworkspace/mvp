package rest

import (
	"encoding/json"
	"fmt"

	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/sessions"
	"gitlab.com/jodworkspace/mvp/pkg/monitor"

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
	monitorManager *monitor.Manager
	httpMonitor    *monitor.HTTPMonitor
}

func NewServer(
	cfg *config.Config,
	sessionStore sessions.Store,
	taskHandler *v1.TaskHandler,
	oauthHandler *v1.OAuthHandler,
	logger *logger.ZapLogger,
	monitorManager *monitor.Manager,
	httpMetrics *monitor.HTTPMonitor,
) *Server {
	return &Server{
		cfg:            cfg,
		sessionStore:   sessionStore,
		taskHandler:    taskHandler,
		oauthHandler:   oauthHandler,
		logger:         logger,
		monitorManager: monitorManager,
		httpMonitor:    httpMetrics,
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

func (s *Server) registerOAuthRoutes(router chi.Router) {
	router.Route("/api/v1/oauth", func(r chi.Router) {
		ir := s.instrumentedRouter(r)
		ir.Post("/token", s.oauthHandler.ExchangeToken)

		irWithAuth := ir.With(middleware.SessionAuth(s.sessionStore, domain.SessionCookieName))
		irWithAuth.Post("/logout", s.oauthHandler.Logout)
		irWithAuth.Get("/userinfo", s.oauthHandler.GetUserInfo)
	})
}

func (s *Server) registerTaskRoutes(router chi.Router) {
	router.Route("/api/v1/tasks", func(r chi.Router) {
		ir := s.instrumentedRouter(r)
		ir.Use(middleware.SessionAuth(s.sessionStore, domain.SessionCookieName))
		ir.With(middleware.Filter).Get("/", s.taskHandler.List)
		ir.Post("/", s.taskHandler.Create)
		ir.Get("/{id}", s.taskHandler.Get)
		ir.Put("/{id}", s.taskHandler.Update)
		ir.Delete("/{id}", s.taskHandler.Delete)
	})
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

	ir.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(httpx.JSON{"status": "ok"})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	s.registerOAuthRoutes(r)
	s.registerTaskRoutes(r)

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

			s.httpMonitor.Handle(routePattern, next).ServeHTTP(w, r)
		})
	})
}
