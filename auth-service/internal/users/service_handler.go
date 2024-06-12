package users

import (
	requestuser "auth-service/pkg/hidepost-requestuser"
	serverlogging "auth-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type serviceHandler struct {
	service *Service
}

func RegisterServiceHandler(api *gin.RouterGroup, service *Service) {
	h := &serviceHandler{service: service}

	api.HEAD("/authorize", h.authorize)
}

func (h *serviceHandler) authorize(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	loggingMap["original_request_url"] = ctx.GetHeader("USER-REQUEST-URL")

	userId, err := h.service.ParseToken(ctx.GetHeader(requestuser.AuthorizationHeaderKey))
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("invalid jwt token")
		loggingMap.Debug()
		ctx.Header(requestuser.UserRoleHeaderKey, requestuser.UserRoleUnknown)
		ctx.Status(http.StatusUnauthorized)
		return
	}
	loggingMap.SetUserId(userId)

	user, err := h.service.ByID(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get user by id")
		ctx.Header(requestuser.UserRoleHeaderKey, requestuser.UserRoleUnknown)
		ctx.Status(http.StatusInternalServerError)
		return
	}
	if user == nil {
		loggingMap.SetMessage("user with id of successfully parsed jwt token not found")
		loggingMap.Error()
		ctx.Header(requestuser.UserRoleHeaderKey, requestuser.UserRoleUnknown)
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}
	if user.Enabled == false {
		loggingMap.SetMessage("user with id of successfully parsed jwt token is not enabled")
		loggingMap.Error()
		ctx.Header(requestuser.UserRoleHeaderKey, requestuser.UserRoleUnknown)
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}
	if user.Deleted {
		loggingMap.SetMessage("user with id of successfully parsed jwt token is deleted")
		loggingMap.Error()
		ctx.Header(requestuser.UserRoleHeaderKey, requestuser.UserRoleUnknown)
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}

	loggingMap.SetMessage("authorized request")
	loggingMap.Debug()

	if user.BannedUntil != nil {
		if time.Now().UTC().Before(*user.BannedUntil) {
			ctx.Header(requestuser.UserIsBannedHeaderKey, "true")
		}
	}
	ctx.Header(requestuser.UserRoleHeaderKey, user.Role)
	ctx.Header(requestuser.UserIdHeaderKey, user.ID.String())
	ctx.Status(http.StatusOK)
}
