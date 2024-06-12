package users

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	serverlogging "registration-service/pkg/serverlogging/gin"
)

type signupHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterSignupHandler(api *gin.RouterGroup, service *Service) {
	h := &signupHandler{
		service:  service,
		validate: NewUsersValidator(),
	}

	api.POST("", h.signup)
}

func (h *signupHandler) signup(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	var req SignupRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to decode signup request")
		ctx.Status(http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad update user admin request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	exists, err := h.service.ExistsByLogin(ctx, req.Login)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check if user exists by login")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if exists {
		loggingMap.SetMessage("user with requested login already exists")
		ctx.Status(http.StatusBadRequest)
		return
	}
	exists, err = h.service.ExistsByEmail(ctx, req.Email)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check if user exists by email")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if exists {
		loggingMap.SetMessage("user with requested email already exists")
		ctx.Status(http.StatusBadRequest)
		return
	}

	err = h.service.HandleSignup(ctx, req.Login, req.Email, req.Password)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to handle signup")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	loggingMap.SetMessage("user successfully signed up")
	ctx.Status(http.StatusCreated)
}
