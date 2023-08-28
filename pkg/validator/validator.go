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
	var (
		str string
	)
	if ok := err.(validator.ValidationErrors); ok != nil {
		for _, v := range ok {
			str = str + v.Field() + " "
		}
	}
	err = errors.New("err: " + strings.ToLower(str) + ": fields required (with: '_')")
	return err
}
