package middleware

import (
	"context"
	"github.com/go-playground/validator/v10"
	"gitlab.com/gookie/mvp/pkg/httpx"
	"net/http"
)

var validate *validator.Validate

type ValidationError struct {
	Field string
	Tag   string
	Value any
}

func validateStruct(s any) []ValidationError {
	var validationErrors []ValidationError
	validate = validator.New(validator.WithRequiredStructEnabled())

	errs := validate.Struct(s)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field: err.Field(),
				Tag:   err.Tag(),
				Value: err.Value(),
			})
		}
	}
	return validationErrors
}

type Middleware func(next http.Handler) http.Handler

func ValidationMiddleware(input any) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := httpx.ReadJSON(r, input)
			if err != nil {
				_, _ = httpx.ErrorJSON(w, &httpx.ErrorResponse{
					StatusCode: http.StatusBadRequest,
					Message:    err.Error(),
				})
			}

			errs := validateStruct(input)
			if len(errs) > 0 {
				w.WriteHeader(http.StatusBadRequest)
			}

			ctx := context.WithValue(r.Context(), "input", input)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
