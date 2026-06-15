package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

var slugPattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()
	_ = v.RegisterValidation("alphanumdash", func(fl validator.FieldLevel) bool {
		return slugPattern.MatchString(fl.Field().String())
	})
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("form"), ",", 2)[0]
		if name == "-" || name == "" {
			name = strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		}
		if name == "-" || name == "" {
			return fld.Name
		}
		return name
	})
	return &Validator{validate: v}
}

func (v *Validator) Validate(i any) map[string]string {
	err := v.validate.Struct(i)
	if err == nil {
		return nil
	}

	errors := make(map[string]string)
	var validationErrors validator.ValidationErrors
	if ok := asValidationErrors(err, &validationErrors); ok {
		for _, fe := range validationErrors {
			field := fe.Field()
			errors[field] = formatValidationError(fe)
		}
		return errors
	}
	return map[string]string{"_error": err.Error()}
}

func asValidationErrors(err error, target *validator.ValidationErrors) bool {
	if ve, ok := err.(validator.ValidationErrors); ok {
		*target = ve
		return true
	}
	return false
}

func formatValidationError(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return fmt.Sprintf("Must be at least %s characters", fe.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s characters", fe.Param())
	case "eqfield":
		return "Fields do not match"
	case "gtfield":
		return "Must be after start date"
	case "url":
		return "Must be a valid URL"
	case "alphanumdash":
		return "Only letters, numbers, and dashes are allowed"
	default:
		return fmt.Sprintf("Failed validation on '%s'", fe.Tag())
	}
}

func FirstError(errors map[string]string) string {
	if len(errors) == 0 {
		return ""
	}
	for _, msg := range errors {
		return msg
	}
	return "Validation failed"
}
