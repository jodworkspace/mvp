package middleware

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"gitlab.com/jodworkspace/mvp/config"
	"gitlab.com/jodworkspace/mvp/internal/domain"
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

func Pagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queryParams := r.URL.Query()

		p := &domain.Pagination{
			Page:      getUInt64(queryParams, "page"),
			PageSize:  getUInt64(queryParams, "pageSize"),
			PageToken: queryParams.Get("pageToken"),
		}

		if p.Page < 1 {
			p.Page = config.DefaultPage
		}
		if p.PageSize < 1 {
			p.PageSize = config.DefaultPageSize
		}

		ctx := context.WithValue(r.Context(), domain.KeyPagination, p)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
