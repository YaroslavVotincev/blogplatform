package notifications

import (
	"github.com/gin-gonic/gin"
	"net/http"
	requestuser "notifications-service/pkg/hidepost-requestuser"
	serverlogging "notifications-service/pkg/serverlogging/gin"
)

type myHandler struct {
	service *Service
}

func RegisterMyNotificationsHandler(api *gin.RouterGroup, service *Service) {
	h := &myHandler{
		service: service,
	}
	api.Use(UserMiddleware())
	api.GET("", h.getMyNotifications)
	api.GET("/count-unseen", h.countUnseen)
	api.POST("/set-seen", h.updateToSeen)
}

func (h *myHandler) getMyNotifications(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	notifications, err := h.service.repository.AllByUser(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get notifications by user")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, notifications)
}

func (h *myHandler) countUnseen(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	count, err := h.service.repository.CountUnseenForUser(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to count unseen notifications")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

func (h *myHandler) updateToSeen(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	err := h.service.repository.UpdateAllToSeenForUser(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update notifications to seen")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusAccepted, nil)
}
