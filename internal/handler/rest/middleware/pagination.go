package middleware

import (
	"context"
	"encoding/json"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"net/http"
)

func Pagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var pagination domain.Pagination
		if err := json.NewDecoder(r.Body).Decode(&pagination); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if pagination.Page < 1 {
			pagination.Page = config.DefaultPage
		}
		if pagination.PageSize < 1 {
			pagination.PageSize = config.DefaultPageSize
		}

		ctx := context.WithValue(r.Context(), "pagination", pagination)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
