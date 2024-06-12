package users

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

var loginRe = regexp.MustCompile("^[a-zA-Z0-9_]+$")

func LoginValidator(fl validator.FieldLevel) bool {
	return loginRe.MatchString(fl.Field().String())
}

func PasswordValidator(fl validator.FieldLevel) bool {
	return len([]byte(fl.Field().String())) <= 72
}

func NewUsersValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("login", LoginValidator)
	validate.RegisterValidation("password", PasswordValidator)
	return validate
}
