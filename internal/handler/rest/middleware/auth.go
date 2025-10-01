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

			userID, ok := session.Values[domain.KeyUserID].(string)
			if !ok || userID == "" {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					Code:    http.StatusUnauthorized,
					Message: "invalid session",
				})
				return
			}

			ctx := helper.ContextWithValues(r.Context(), map[string]any{
				domain.KeyUserID:      userID,
				domain.KeyIssuer:      session.Values[domain.KeyIssuer],
				domain.KeyAccessToken: session.Values[domain.KeyAccessToken],
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
