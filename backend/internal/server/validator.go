package server

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// https://github.com/go-playground/validator/issues/559#issuecomment-1871786235
type validationErrors []error

func (v validationErrors) Error() string {
	var message string

	for i, err := range v {
		if i > 0 {
			message += ","
		}
		message += err.Error()
	}

	return message
}

type Validator struct {
	validator *validator.Validate
}

func (v *Validator) Validate(p any) error {
	if err := v.validator.Struct(p); err != nil {
		var vErrs validationErrors

		for _, err := range err.(validator.ValidationErrors) {
			var e error
			switch err.Tag() {
			case "required":
				e = fmt.Errorf("'%s' is required", err.Field())
			case "number":
				e = fmt.Errorf("'%s' should have numeric value", err.Field())
			case "gte":
				e = fmt.Errorf("'%s' should be greater than or equal to %s", err.Field(), err.Param())
			case "gt":
				e = fmt.Errorf("'%s' should be greater than %s", err.Field(), err.Param())
			default:
				e = fmt.Errorf("'%s': '%v' must satisfy '%s' '%v' criteria", err.Field(), err.Value(), err.Tag(), err.Param())
			}
			vErrs = append(vErrs, e)
		}

		return echo.NewHTTPError(http.StatusBadRequest, vErrs.Error())
	}
	return nil
}

func NewValidator() *Validator {
	v := validator.New()

	// https://github.com/go-playground/validator/issues/861
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		// skip if tag key says it should be ignored
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validator: v,
	}
}
