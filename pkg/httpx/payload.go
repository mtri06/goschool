package httpx

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func init() {
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		jsonTag := fld.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			return fld.Name
		}
		return strings.SplitN(jsonTag, ",", 2)[0]
	})
}

func DecodeBody[T any](r *http.Request) (*T, error) {
	var payload T
	err := render.DecodeJSON(r.Body, &payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to decode request body")
		return nil, ErrInvalidBody
	}
	err = validate.Struct(payload)
	if err != nil {
		if ve, ok := err.(validator.ValidationErrors); ok {
			return nil, ErrValidationFailed.WithMsg(sanitizeValidationError(ve))
		}
		log.Error().Err(err).Msg("validation error")
		return nil, ErrValidationFailed
	}
	return &payload, nil
}

type validationMsgFn func(field, param string) string

var validationMessages = map[string]validationMsgFn{
	"required": func(field, _ string) string {
		return fmt.Sprintf("`%v` is required", field)
	},
	"email": func(field, _ string) string {
		return fmt.Sprintf("`%v` must be a valid email address", field)
	},
	"min": func(field, param string) string {
		return fmt.Sprintf("`%v` must be at least %s characters long", field, param)
	},
	"max": func(field, param string) string {
		return fmt.Sprintf("`%v` must be at most %s characters long", field, param)
	},
	"len": func(field, param string) string {
		return fmt.Sprintf("`%v` must be exactly %s characters long", field, param)
	},
	"gte": func(field, param string) string {
		return fmt.Sprintf("`%v` must be greater than or equal to %s", field, param)
	},
	"lte": func(field, param string) string {
		return fmt.Sprintf("`%v` must be less than or equal to %s", field, param)
	},
	"gt": func(field, param string) string {
		return fmt.Sprintf("`%v` must be greater than %s", field, param)
	},
	"lt": func(field, param string) string {
		return fmt.Sprintf("`%v` must be less than %s", field, param)
	},
	"oneof": func(field, param string) string {
		return fmt.Sprintf("`%v` must be one of: %s", field, strings.ReplaceAll(param, " ", ", "))
	},
	"url": func(field, _ string) string {
		return fmt.Sprintf("`%v` must be a valid URL", field)
	},
	"uuid": func(field, _ string) string {
		return fmt.Sprintf("`%v` must be a valid UUID", field)
	},
	"numeric": func(field, _ string) string {
		return fmt.Sprintf("`%v` must be a numeric value", field)
	},
	"alphanum": func(field, _ string) string {
		return fmt.Sprintf("`%v` must contain only alphanumeric characters", field)
	},
	"eqfield": func(field, param string) string {
		return fmt.Sprintf("`%v` must match %s", field, param)
	},
}

// sanitizeValidationError returns a human-readable message for all validation errors.
func sanitizeValidationError(errs validator.ValidationErrors) string {
	var sb strings.Builder
	for _, fe := range errs {
		if fn, ok := validationMessages[fe.Tag()]; ok {
			sb.WriteString(fn(fe.Field(), fe.Param()))
		} else {
			fmt.Fprintf(&sb, "%s is invalid (%s)", fe.Field(), fe.Tag())
		}
		sb.WriteString("; ")
	}
	return strings.TrimSuffix(sb.String(), "; ")
}
