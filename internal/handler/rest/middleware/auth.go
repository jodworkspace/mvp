package middleware

import (
	"net/http"

	"github.com/gorilla/sessions"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"gitlab.com/jodworkspace/mvp/pkg/utils/cipherx"
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
				domain.KeyUserID:       userID,
				domain.KeyIssuer:       session.Values[domain.KeyIssuer],
				domain.KeyAccessToken:  session.Values[domain.KeyAccessToken],
				domain.KeyRefreshToken: session.Values[domain.KeyRefreshToken],
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func DecryptToken(aead *cipherx.AEAD) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encryptedAccessToken, ok := r.Context().Value(domain.KeyAccessToken).(string)
			if !ok {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: "google token not found",
				})
				return
			}

			accessToken, err := aead.Decrypt([]byte(encryptedAccessToken))
			if err != nil {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					Code:    http.StatusInternalServerError,
					Message: err.Error(),
				})
				return
			}

			ctx := helper.ContextWithValues(r.Context(), map[string]any{
				domain.KeyAccessToken: string(accessToken),
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
