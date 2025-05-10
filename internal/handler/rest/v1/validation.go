package v1

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"gitlab.com/gookie/mvp/pkg/utils/httpx"
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

func Bind(r *http.Request, input any) error {
	err := httpx.ReadJSON(r, input)
	if err != nil {
		return err
	}

	errs := validateStruct(input)
	if len(errs) > 0 {
		return fmt.Errorf("invalid value for %s: expected %s, got %v", errs[0].Field, errs[0].Tag, errs[0].Value)
	}

	return nil
}
