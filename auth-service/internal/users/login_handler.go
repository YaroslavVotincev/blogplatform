package users

import (
	"auth-service/pkg/cryptservice"
	serverlogging "auth-service/pkg/serverlogging/gin"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type loginHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterLoginHandler(api *gin.RouterGroup, service *Service) {
	h := &loginHandler{
		service:  service,
		validate: NewUsersValidator(),
	}

	api.POST("", h.login)
}

func (h *loginHandler) login(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)

	var req LoginRequest

	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad login request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad login request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["request_login"] = req.Login
	loggingMap["request_remember_me"] = req.RememberMe

	user, err := h.service.ByValue(ctx, req.Login)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to fetch user by request login")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if user == nil {
		loggingMap.SetMessage("wrong login or password")
		loggingMap.Debug()
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}

	if !cryptservice.ValueHashMatched(req.Password, user.HashedPassword) {
		loggingMap.SetMessage("wrong login or password")
		loggingMap.Debug()
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}

	loggingMap.SetUserId(&user.ID)

	token, err := h.service.GenerateToken(user.ID, req.RememberMe)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("jwt token generation failed for user")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	go h.service.notifService.Authentication(user.ID.String())

	loggingMap.SetMessage("user successfully logged in")
	loggingMap.Info()
	ctx.JSON(http.StatusOK, LoginResponse{Token: token})
}
