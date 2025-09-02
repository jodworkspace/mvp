package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/utils/helper"
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

			ctx := helper.ContextWithValues(r.Context(), map[string]any{
				domain.SessionKeyUserID: userID,
				domain.SessionKeyIssuer: session.Values[domain.SessionKeyIssuer],
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
