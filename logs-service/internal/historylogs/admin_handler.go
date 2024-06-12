package historylogs

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	serverlogging "logs-service/pkg/serverlogging/gin"
	"strconv"
	"strings"
	"time"

	"net/http"
)

type adminHandler struct {
	service *Service
}

func RegisterAdminHandler(api *gin.RouterGroup, service *Service) {
	h := &adminHandler{service: service}
	api.Use(AdminMiddleware())
	api.GET("/healthcheck", h.healthCheck)
	api.GET("/logs", h.all)
	api.GET("/logs/:id", h.byId)
	api.GET("/levels", h.levelsList)
	api.GET("/services", h.servicesList)
}

func (h *adminHandler) healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, HealthCheckResponse{Status: "ok", Up: true})
}

func (h *adminHandler) all(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	levelQuery := ctx.Query("level")
	loggingMap["level_query"] = levelQuery
	var levels []string
	if levelQuery != "" {
		levels = strings.Split(levelQuery, ",")
	}

	serviceQuery := ctx.Query("service")
	loggingMap["service_query"] = serviceQuery
	var services []string
	if serviceQuery != "" {
		services = strings.Split(serviceQuery, ",")
	}

	userIdQuery := ctx.Query("user")
	loggingMap["user_id_query"] = userIdQuery
	var userIDs []*uuid.UUID
	if userIdQuery != "" {
		userIDs = make([]*uuid.UUID, 0)
		for _, value := range strings.Split(userIdQuery, ",") {
			id, err := uuid.Parse(value)
			if err != nil {
				switch value {
				case "none", "null", "nil":
					userIDs = append(userIDs, nil)
				}
			}
			userIDs = append(userIDs, &id)
		}
	}

	startTimeQuery := ctx.Query("start")
	loggingMap["start_time_query"] = startTimeQuery
	var startTime *time.Time
	if startTimeQuery != "" {
		parsedTime, err := time.Parse("2006-01-02T15:04", startTimeQuery)
		if err == nil {
			startTime = &parsedTime
		}
	}

	endTimeQuery := ctx.Query("end")
	loggingMap["end_time_query"] = endTimeQuery
	var endTime *time.Time
	if endTimeQuery != "" {
		parsedTime, err := time.Parse("2006-01-02T15:04", endTimeQuery)
		if err == nil {
			endTime = &parsedTime
		}
	}

	limitQuery := ctx.Query("limit")
	loggingMap["limit_query"] = limitQuery
	var limit = 50
	if limitQuery != "" {
		parseQueryNumber, err := strconv.Atoi(limitQuery)
		if err == nil {
			limit = parseQueryNumber
		}
	}

	skipQuery := ctx.Query("skip")
	loggingMap["skip_query"] = skipQuery
	var skip = 0
	if skipQuery != "" {
		parseQueryNumber, err := strconv.Atoi(skipQuery)
		if err == nil {
			skip = parseQueryNumber
		}
	}

	logs, err := h.service.All(ctx, levels, services, userIDs, startTime, endTime, &limit, &skip)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get logs from db")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, logs)
}

func (h *adminHandler) byId(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	loggingMap["id"] = idParam

	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to parse id url param")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	log, err := h.service.ByID(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get by id log from db")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, log)
}

func (h *adminHandler) levelsList(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	levels, err := h.service.Levels(ctx)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get levels from db")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, levels)
}

func (h *adminHandler) servicesList(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	services, err := h.service.Services(ctx)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get services from db")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, services)
}
