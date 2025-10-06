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
	"gitlab.com/jodworkspace/mvp/pkg/otel"
	otelhttp "gitlab.com/jodworkspace/mvp/pkg/otel/http"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"

	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/internal/handler/rest/middleware"
	v1 "gitlab.com/jodworkspace/mvp/internal/handler/rest/v1"
	"gitlab.com/jodworkspace/mvp/pkg/logger"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
)

type Server struct {
	cfg             *config.Config
	aead            *cipherx.AEAD
	sessionStore    sessions.Store
	taskHandler     *v1.TaskHandler
	oauthHandler    *v1.OAuthHandler
	documentHandler *v1.DocumentHandler
	wsHandler       *v1.WSHandler
	logger          *logger.ZapLogger
	monitorManager  *otel.Manager
}

func NewServer(
	cfg *config.Config,
	aead *cipherx.AEAD,
	sessionStore sessions.Store,
	taskHandler *v1.TaskHandler,
	oauthHandler *v1.OAuthHandler,
	documentHandler *v1.DocumentHandler,
	wsHandler *v1.WSHandler,
	logger *logger.ZapLogger,
	monitorManager *otel.Manager,
) *Server {
	return &Server{
		cfg:             cfg,
		aead:            aead,
		sessionStore:    sessionStore,
		taskHandler:     taskHandler,
		oauthHandler:    oauthHandler,
		documentHandler: documentHandler,
		wsHandler:       wsHandler,
		logger:          logger,
		monitorManager:  monitorManager,
	}
}

func (s *Server) Run() error {
	httpMonitor, err := otelhttp.NewMonitor()
	if err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%s", s.cfg.Server.Host, s.cfg.Server.Port)
	srv := &http.Server{
		Addr:           addr,
		Handler:        s.RestMux(httpMonitor),
		MaxHeaderBytes: 1 << 20,
	}

	errChan := make(chan error, 1)
	go func() {
		s.logger.Info("Server started on " + addr)
		err = srv.ListenAndServe() // blocking call
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

func (s *Server) registerOAuthRoutes(router chi.Router, m *otelhttp.Monitor) {
	router.Route("/api/v1/oauth", func(r chi.Router) {
		ir := s.instrumentedRouter(r, m)
		ir.Post("/token", s.oauthHandler.ExchangeToken)

		irWithAuth := ir.With(middleware.SessionAuth(s.sessionStore, domain.SessionCookieName))
		irWithAuth.Post("/logout", s.oauthHandler.Logout)
		irWithAuth.Get("/userinfo", s.oauthHandler.GetUserInfo)
	})
}

func (s *Server) registerTaskRoutes(router chi.Router, m *otelhttp.Monitor) {
	router.Route("/api/v1/tasks", func(r chi.Router) {
		ir := s.instrumentedRouter(r, m)
		ir.Use(middleware.SessionAuth(s.sessionStore, domain.SessionCookieName))
		ir.With(middleware.Pagination).Get("/", s.taskHandler.List)
		ir.Post("/", s.taskHandler.Create)
		ir.Get("/{id}", s.taskHandler.Get)
		ir.Put("/{id}", s.taskHandler.Update)
		ir.Delete("/{id}", s.taskHandler.Delete)
	})
}

func (s *Server) registerDocumentRoutes(router chi.Router, m *otelhttp.Monitor) {
	router.Route("/api/v1/documents", func(r chi.Router) {
		ir := s.instrumentedRouter(r, m)
		ir.Use(middleware.SessionAuth(s.sessionStore, domain.SessionCookieName))
		ir.Use(middleware.DecryptToken(s.aead))
		ir.With(middleware.Pagination).Get("/", s.documentHandler.List)
	})
}

func (s *Server) RestMux(m *otelhttp.Monitor) *chi.Mux {
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

	r.Get("/ws", s.wsHandler.Handle)
	r.Handle("/metrics", s.monitorManager.PrometheusHandler())

	ir := s.instrumentedRouter(r, m)
	ir.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		data, _ := json.Marshal(httpx.JSON{"status": "ok"})
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	})

	s.registerOAuthRoutes(r, m)
	s.registerTaskRoutes(r, m)
	s.registerDocumentRoutes(r, m)

	ir.NotFound(NotFoundRoute)
	ir.MethodNotAllowed(NotFoundRoute)
	return r
}

func NotFoundRoute(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

func (s *Server) instrumentedRouter(r chi.Router, m *otelhttp.Monitor) chi.Router {
	return r.With(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			routePattern := chi.RouteContext(r.Context()).RoutePattern()
			if routePattern == "" {
				routePattern = "unmatched"
			}

			m.Handle(routePattern, next).ServeHTTP(w, r)
		})
	})
}
