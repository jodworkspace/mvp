package v1

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"gitlab.com/jodworkspace/mvp/pkg/utils/httpx"
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

func BindWithValidation(r *http.Request, input any) (err error, details []string) {
	err = httpx.ReadJSON(r, input)
	if err != nil {
		return
	}

	errs := validateStruct(input)
	if len(errs) > 0 {
		err = fmt.Errorf(
			"invalid value for %s: expected %s, got %v",
			errs[0].Field,
			errs[0].Tag,
			errs[0].Value,
		)
	}

	for _, e := range errs {
		details = append(details, fmt.Sprintf(
			"invalid value for %s: expected %s, got %v",
			e.Field,
			e.Tag,
			e.Value,
		))
	}

	return
}
