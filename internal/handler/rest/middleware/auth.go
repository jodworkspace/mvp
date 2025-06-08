package middleware

import (
	"context"
	"github.com/gorilla/sessions"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
	"gitlab.com/gookie/mvp/pkg/utils/jwtx"
	"net/http"
)

func SessionAuth(store sessions.Store, name string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, err := store.Get(r, name)
			if err != nil {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					StatusCode: http.StatusUnauthorized,
					Message:    "invalid session",
				})
				return
			}

			user := session.Values["user"]
			if user == nil {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					StatusCode: http.StatusUnauthorized,
					Message:    "invalid session",
				})
				return
			}

			ctx := context.WithValue(r.Context(), "user", user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func TokenAuth(secret []byte, issuer ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var accessToken string

			for _, cookie := range r.Cookies() {
				if cookie.Name == "access_token" {
					accessToken = cookie.Value
				}
			}

			if accessToken == "" {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					StatusCode: http.StatusUnauthorized,
					Message:    "invalid authorization header",
				})
				return
			}

			claims, err := jwtx.ParseToken(accessToken, secret, issuer...)
			if err != nil {
				_ = httpx.ErrorJSON(w, httpx.ErrorResponse{
					StatusCode: http.StatusUnauthorized,
					Message:    err.Error(),
				})
				return
			}

			ctx := context.WithValue(r.Context(), "claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
