package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/sessions"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
)

func SessionAuth(store sessions.Store, name string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.New(r, name)
			if err != nil || session.IsNew {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					Code:    http.StatusUnauthorized,
					Message: "invalid session",
				})
				return
			}

			userID, ok := session.Values[domain.SessionKeyUserID].(string)
			if !ok || userID == "" {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					Code:    http.StatusUnauthorized,
					Message: "invalid session",
				})
				return
			}

			ctx := context.WithValue(r.Context(), domain.SessionKeyUserID, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
