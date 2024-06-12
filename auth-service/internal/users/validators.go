package users

import (
	"github.com/go-playground/validator/v10"
)

func NewUsersValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	return validate
}
