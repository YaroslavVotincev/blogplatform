package comments

import "github.com/go-playground/validator/v10"

func NewValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate
}
