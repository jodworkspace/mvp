package rest

import (
	"encoding/json"
	"gitlab.com/tokpok/mvp/internal/handler/rest/v1"
	postgresrepo "gitlab.com/tokpok/mvp/internal/repository/postgres"
	user "gitlab.com/tokpok/mvp/internal/usecase/user"
	"net/http"
)

func (s *Server) RestHandlersRoute() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthcheck", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"status": "ok",
			"host":   r.Host,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	})

	userRepository := postgresrepo.NewUserRepository(s.db)
	userUsecase := user.NewUserUsecase(userRepository, s.logger)
	userHandler := v1.NewUserHandler(userUsecase, s.logger)
	mux.HandleFunc("GET /authorize", userHandler.GoogleAuthorize)
	mux.HandleFunc("POST /redirect", userHandler.GoogleRedirectEndpoint)

	return mux
}
