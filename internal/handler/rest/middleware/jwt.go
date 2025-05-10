package middleware

import (
	"context"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
	"gitlab.com/gookie/mvp/pkg/utils/jwtx"
	"net/http"
)

type Middleware func(next http.Handler) http.Handler

func Auth(secret []byte, issuer ...string) Middleware {
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
					Message:    "",
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
