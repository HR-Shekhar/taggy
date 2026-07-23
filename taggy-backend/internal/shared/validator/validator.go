package validator

import (
	"fmt"

	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
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

	default:
		return "is invalid"
	}
}