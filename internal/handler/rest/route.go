package rest

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"gitlab.com/gookie/mvp/internal/handler/rest/v1"
	postgresrepo "gitlab.com/gookie/mvp/internal/repository/postgres"
	taskuc "gitlab.com/gookie/mvp/internal/usecase/task"
	user "gitlab.com/gookie/mvp/internal/usecase/user"
	"net/http"
)

func (s *Server) RestHandlersRoute() *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)

	r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		data, err := json.Marshal(map[string]any{
			"status": "ok",
			"host":   r.Host,
		})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	taskRepository := postgresrepo.NewTaskRepository(s.db)
	taskUsecase := taskuc.NewTaskUsecase(taskRepository, s.logger)
	taskHandler := v1.NewTaskHandler(taskUsecase, s.logger)
	r.Route("/tasks", func(r chi.Router) {
		r.Post("", taskHandler.CreateNewTask)
	})

	userRepository := postgresrepo.NewUserRepository(s.db)
	userUsecase := user.NewUserUsecase(userRepository, s.logger)
	oauthHandler := v1.NewOAuthHandler(userUsecase, s.logger, s.cfg)

	r.Route("/google", func(r chi.Router) {
		r.Post("/token", oauthHandler.GetToken)
	})
	r.Get("/userinfo", oauthHandler.GetUserInfo)
	return r
}
