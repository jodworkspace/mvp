package middleware

import (
	"context"
	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
	"net/http"
	"net/url"
	"strconv"
)

func getUInt64(values url.Values, key string) uint64 {
	v := values.Get(key)

	if v == "" {
		return 0
	}

	i, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func Filter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		filter := &domain.Filter{
			Page:       getUInt64(queryParams, "page"),
			PageSize:   getUInt64(queryParams, "pageSize"),
			Conditions: map[string]any{},
		}

		if filter.Page < 1 {
			filter.Page = config.DefaultPage
		}
		if filter.PageSize < 1 {
			filter.PageSize = config.DefaultPageSize
		}

		ctx := context.WithValue(r.Context(), "filter", filter)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
