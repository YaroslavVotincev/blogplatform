package users

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type adminHandler struct {
	service *Service
}

func RegisterAdminHandler(api *gin.RouterGroup, service *Service) {
	h := &adminHandler{service: service}

	api.GET("/healthcheck", h.healthCheck)
}

func (h *adminHandler) healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, HealthCheckResponse{Status: "ok", Up: true})
}
