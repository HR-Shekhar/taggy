package validator

import (
	"github.com/go-playground/validator/v10"
)

// Validator adapts go-playground/validator to Echo's Validator interface.
type Validator struct {
	validate *validator.Validate
}

// New creates a configured validator instance.
func New() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

// Validate satisfies Echo's Validator interface.
//
// Echo calls this whenever c.Validate(...) is invoked.
func (v *Validator) Validate(i any) error {
	return v.validate.Struct(i)
}