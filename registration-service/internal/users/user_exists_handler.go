package users

import (
	"github.com/gin-gonic/gin"
	"net/http"
	serverlogging "registration-service/pkg/serverlogging/gin"
)

type userExistsHandler struct {
	service *Service
}

func RegisterUserExistsHandler(api *gin.RouterGroup, service *Service) {
	h := &userExistsHandler{
		service: service,
	}

	api.GET("/login/:login", h.existsByLogin)
	api.GET("/email/:email", h.existsByEmail)
}

func (h *userExistsHandler) existsByLogin(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	login := ctx.Param("login")
	loggingMap["login"] = login

	exists, err := h.service.ExistsByLogin(ctx, login)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check if user exists by login")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if exists {
		ctx.Status(http.StatusOK)
	} else {
		ctx.Status(http.StatusNotFound)
	}
}

func (h *userExistsHandler) existsByEmail(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	email := ctx.Param("email")
	loggingMap["email"] = email

	exists, err := h.service.ExistsByEmail(ctx, email)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check if user exists by email")
		ctx.Status(http.StatusInternalServerError)
		return
	}

	if exists {
		ctx.Status(http.StatusOK)
	} else {
		ctx.Status(http.StatusNotFound)
	}
}
