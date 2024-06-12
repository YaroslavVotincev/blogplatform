package users

import (
	"github.com/gin-gonic/gin"
	"net/http"
	serverlogging "registration-service/pkg/serverlogging/gin"
)

type confirmHandler struct {
	service *Service
}

func RegisterConfirmHandler(api *gin.RouterGroup, service *Service) {
	h := &confirmHandler{
		service: service,
	}

	api.GET("/:code", h.confirm)
}

func (h *confirmHandler) confirm(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	confirmCode := ctx.Param("code")
	loggingMap["confirm_code"] = confirmCode

	userId, err := h.service.UserIdByConfirmCode(ctx, confirmCode)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check the signup confirmation code")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if userId == nil {
		loggingMap.SetMessage("signup confirmation code not found")
		ctx.Status(http.StatusNotFound)
		return
	}

	user, err := h.service.repository.ByID(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check if user exists by id")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if user == nil {
		_ = h.service.EraseConfirmCode(ctx, confirmCode)
		loggingMap.SetMessage("user by confirmation code already deleted")
		ctx.Status(http.StatusNotFound)
		return
	}

	err = h.service.Enable(ctx, *userId, confirmCode)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to enable user")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	go h.service.notifService.Registration(userId.String(), user.Email, user.Login)

	loggingMap.SetMessage("user signup confirmed")
	ctx.Status(http.StatusOK)
}
