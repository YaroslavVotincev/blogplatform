package configservice

import (
	serverlogging "config-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

type serviceHandler struct {
	service *Service
}

func RegisterServiceHandler(api *gin.RouterGroup, service *Service) {
	h := &serviceHandler{service: service}
	api.GET("", h.serviceConfiguration)

}

func (h *serviceHandler) serviceConfiguration(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	serviceName := ctx.Request.Header.Get("SERVICE_NAME")
	if serviceName == "" {
		loggingMap.SetMessage("empty service name header")
		ctx.Status(http.StatusBadRequest)
		return
	}
	loggingMap["service_name"] = serviceName

	service, err := h.service.ServiceByName(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get service from db")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if service == nil {
		loggingMap.SetMessage("service by name not found")
		ctx.Status(http.StatusNotFound)
		return
	}

	settings, err := h.service.SettingsByService(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get settings from db")
		ctx.Status(http.StatusInternalServerError)
		return
	}
	loggingMap.None()
	ctx.JSON(http.StatusOK, settings)
}
