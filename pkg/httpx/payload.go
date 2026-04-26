package httpx

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func DecodeBody[T any](r *http.Request) (*T, error) {
	var payload T
	err := render.DecodeJSON(r.Body, &payload)
	if err != nil {
		return nil, ErrInvalidBody.WithErr(err)
	}
	err = validate.Struct(payload)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			var sb strings.Builder
			for _, fe := range ve {
				sb.WriteString(ValidationError(fe))
				sb.WriteString("; ")
			}
			err = errors.New(strings.TrimSuffix(sb.String(), "; "))
		}
		return nil, ErrValidationFailed.WithErr(err)
	}
	return &payload, nil
}

type validationMsgFn func(field, param string) string

var validationMessages = map[string]validationMsgFn{
	"required": func(field, _ string) string {
		return fmt.Sprintf("%s is required", field)
	},
	"email": func(field, _ string) string {
		return fmt.Sprintf("%s must be a valid email address", field)
	},
	"min": func(field, param string) string {
		return fmt.Sprintf("%s must be at least %s characters long", field, param)
	},
	"max": func(field, param string) string {
		return fmt.Sprintf("%s must be at most %s characters long", field, param)
	},
	"len": func(field, param string) string {
		return fmt.Sprintf("%s must be exactly %s characters long", field, param)
	},
	"gte": func(field, param string) string {
		return fmt.Sprintf("%s must be greater than or equal to %s", field, param)
	},
	"lte": func(field, param string) string {
		return fmt.Sprintf("%s must be less than or equal to %s", field, param)
	},
	"gt": func(field, param string) string {
		return fmt.Sprintf("%s must be greater than %s", field, param)
	},
	"lt": func(field, param string) string {
		return fmt.Sprintf("%s must be less than %s", field, param)
	},
	"oneof": func(field, param string) string {
		return fmt.Sprintf("%s must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	},
	"url": func(field, _ string) string {
		return fmt.Sprintf("%s must be a valid URL", field)
	},
	"uuid": func(field, _ string) string {
		return fmt.Sprintf("%s must be a valid UUID", field)
	},
	"numeric": func(field, _ string) string {
		return fmt.Sprintf("%s must be a numeric value", field)
	},
	"alphanum": func(field, _ string) string {
		return fmt.Sprintf("%s must contain only alphanumeric characters", field)
	},
	"eqfield": func(field, param string) string {
		return fmt.Sprintf("%s must match %s", field, param)
	},
}

// ValidationError returns a human-readable message for a validator.FieldError
func ValidationError(err validator.FieldError) string {
	if fn, ok := validationMessages[err.Tag()]; ok {
		return fn(err.Field(), err.Param())
	}
	return fmt.Sprintf("%s is invalid (%s)", err.Field(), err.Tag())
}
