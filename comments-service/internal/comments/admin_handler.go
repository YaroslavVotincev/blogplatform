package comments

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type adminHandler struct {
}

func RegisterAdminHandler(api *gin.RouterGroup) {
	h := &adminHandler{}

	api.GET("/healthcheck", h.healthCheck)
}

func (h *adminHandler) healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, HealthCheckResponse{Status: "ok", Up: true})
}
