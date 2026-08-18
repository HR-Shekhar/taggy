package validator

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode"

	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/go-playground/validator/v10"
)

var specialCharPattern = regexp.MustCompile(`[^a-zA-Z0-9]`)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})

	_ = v.RegisterValidation("strong_password", strongPassword)

	return &Validator{validate: v}
}

// strongPassword requires at least one uppercase, lowercase, digit, and special character.
func strongPassword(fl validator.FieldLevel) bool {
	password, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		case specialCharPattern.MatchString(string(r)):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func (v *Validator) Validate(i any) error {
	if err := v.validate.Struct(i); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			fields := make(map[string]string)

			for _, fieldErr := range validationErrors {
				fields[fieldErr.Field()] = validationMessage(fieldErr)
			}

			return apperrors.ValidationError{
				Fields: fields,
			}
		}

		return err
	}

	return nil
}

func validationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "is required"

	case "email":
		return "must be a valid email"

	case "min":
		return fmt.Sprintf("must be at least %s characters", err.Param())

	case "max":
		return fmt.Sprintf("must be at most %s characters", err.Param())

	case "oneof":
		return fmt.Sprintf("must be one of: %s", err.Param())

	case "len":
		return fmt.Sprintf("must be exactly %s characters", err.Param())

	case "numeric":
		return "must contain only digits"

	case "strong_password":
		return "must include uppercase, lowercase, number, and special character"

	default:
		return "is invalid"
	}
}
