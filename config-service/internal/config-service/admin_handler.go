package configservice

import (
	serverlogging "config-service/pkg/serverlogging/gin"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

type adminHandler struct {
	service   *Service
	validator *validator.Validate
}

func RegisterAdminHandler(api *gin.RouterGroup, service *Service) {
	h := &adminHandler{service: service, validator: validator.New(validator.WithRequiredStructEnabled())}
	api.Use(AdminMiddleware())
	api.GET("/healthcheck", h.healthCheck)
	api.GET("/services", h.listServices)
	api.POST("/services", h.createService)
	api.GET("/services/:service", h.serviceByName)
	api.PUT("/services/:service", h.updateService)
	api.DELETE("/services/:service", h.deleteService)
	api.GET("/services/:service/settings", h.serviceSettings)
	api.POST("/services/:service/settings", h.newServiceSettings)
}

func (h *adminHandler) healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, HealthCheckResponse{Status: "ok", Up: true})
}

func (h *adminHandler) listServices(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	services, err := h.service.AllServices(ctx)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get all services")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, services)
}

func (h *adminHandler) serviceByName(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	serviceName := ctx.Param("service")
	loggingMap["service_name"] = serviceName
	service, err := h.service.ServiceByName(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get service by name")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if service == nil {
		loggingMap.SetMessage("service by name not found")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	ctx.JSON(http.StatusOK, service)
}

func (h *adminHandler) createService(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	var req CreateServiceRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("invalid request body, failed to unmarshal request")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["request_body"] = fmt.Sprintf("%+v", req)
	if err := h.validator.Struct(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("invalid request body, failed to validate")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	existingService, err := h.service.ServiceByName(ctx, req.Service)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get service by name")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if existingService != nil {
		loggingMap.SetMessage("service by name already exists")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	service, err := h.service.NewService(ctx, req.Service)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create new service")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap["new_service"] = fmt.Sprintf("%+v", *service)
	loggingMap.SetMessage("created new service")
	ctx.JSON(http.StatusCreated, service)
}

func (h *adminHandler) updateService(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	var req UpdateServiceRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("invalid request body, failed to unmarshal request")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["request_body"] = fmt.Sprintf("%+v", req)
	if err := h.validator.Struct(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("invalid request body, failed to validate request")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	serviceName := ctx.Param("service")
	loggingMap["service_name"] = serviceName

	service, err := h.service.ServiceByName(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get service by name")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if service == nil {
		loggingMap.SetMessage("service by name doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	loggingMap["service_old"] = fmt.Sprintf("%+v", *service)

	existingService, err := h.service.ServiceByName(ctx, req.Service)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get existing service by name")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if existingService != nil && existingService.Service != service.Service {
		loggingMap.SetMessage("service by request name already exists")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err = h.service.UpdateService(ctx, service, req.Service)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update service")
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	loggingMap.SetMessage("updated service name")
	loggingMap["service_new"] = fmt.Sprintf("%+v", *service)
	ctx.JSON(http.StatusAccepted, service)
}

func (h *adminHandler) deleteService(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	serviceName := ctx.Param("service")
	loggingMap["service_name"] = serviceName

	service, err := h.service.ServiceByName(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get service by name")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if service == nil {
		loggingMap.SetMessage("service by name doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	loggingMap["service_old"] = fmt.Sprintf("%+v", *service)

	err = h.service.DeleteService(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to delete service")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (h *adminHandler) serviceSettings(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	serviceName := ctx.Param("service")
	loggingMap["service_name"] = serviceName
	service, err := h.service.ServiceByName(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get service by name")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if service == nil {
		loggingMap.SetMessage("service by name doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	loggingMap["service_object"] = fmt.Sprintf("%+v", *service)

	settings, err := h.service.SettingsByService(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get settings of service")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, settings)
}

func (h *adminHandler) newServiceSettings(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	serviceName := ctx.Param("service")
	loggingMap["service_name"] = serviceName
	service, err := h.service.ServiceByName(ctx, serviceName)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get service by name")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if service == nil {
		loggingMap.SetMessage("service by name doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	loggingMap["service_object"] = fmt.Sprintf("%+v", *service)

	var req []CreateSettingRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("invalid request body, failed to unmarshal request")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["request_body"] = fmt.Sprintf("%+v", req)
	if err := h.validator.Var(req, "dive"); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("invalid request body, failed to validate request")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err = h.service.SetSettingsToService(ctx, serviceName, req)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set settings to service")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("set new service settings")
	ctx.JSON(http.StatusAccepted, nil)
}
