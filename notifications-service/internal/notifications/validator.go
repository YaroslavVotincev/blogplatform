package notifications

import (
	"github.com/go-playground/validator/v10"
)

//var loginRe = regexp.MustCompile("^[a-zA-Z0-9_]+$")
//var urlPathRe = regexp.MustCompile("^[a-zA-Z0-9_-]+$")
//
//func LoginValidator(fl validator.FieldLevel) bool {
//	return loginRe.MatchString(fl.Field().String())
//}
//
//func UrlPathValidator(fl validator.FieldLevel) bool {
//	return urlPathRe.MatchString(fl.Field().String())
//}

func NewValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())
	//validate.RegisterValidation("login", LoginValidator)
	//validate.RegisterValidation("urlpath", UrlPathValidator)
	return validate
}
