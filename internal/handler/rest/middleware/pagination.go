package middleware

import (
	"context"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"net/http"
	"net/url"
	"strconv"
)

func getInt(values url.Values, key string) int {
	v := values.Get(key)

	if v == "" {
		return 0
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		return 0
	}
	return i
}

func Pagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		pagination := &domain.Pagination{
			Page:     getInt(queryParams, "page"),
			PageSize: getInt(queryParams, "page_size"),
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
