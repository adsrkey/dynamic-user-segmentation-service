package validator

import (
	"errors"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

// Validator wraps the go playground validator for the echo framework interface.
type Validator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{validator: validator.New()}
}

// Validate implements the echo framework validator interface.
func (val *Validator) Validate(i interface{}) error {
	err := val.validator.Struct(i)
	if err == nil {
		return nil
	}
	str := err.Error()
	// ind := strings.Index(str, ":Field")
	// _ = ind
	indx := strings.Index(str, "Field validation")
	if indx > 0 {
		str = str[indx:]
	}

	err = errors.New(strings.Replace(strings.ToLower(str), "\n", ", ", -1))
	return err
}