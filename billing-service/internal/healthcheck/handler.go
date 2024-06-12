package healthcheck

import (
	requestuser "billing-service/pkg/hidepost-requestuser"
	serverlogging "billing-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

type healthcheckHandler struct {
}

func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		loggingMap := serverlogging.GetLoggingMap(ctx)
		userId := requestuser.GetUserID(ctx)
		if userId == nil {
			loggingMap.SetMessage("request user is not authenticated")
			loggingMap["user_id_header"] = ctx.GetHeader(requestuser.UserIdHeaderKey)
			loggingMap["user_role_header"] = ctx.GetHeader(requestuser.UserRoleHeaderKey)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		loggingMap.SetUserId(userId)
		if !requestuser.IsAdmin(ctx) {
			loggingMap.SetMessage("request user is not admin")
			loggingMap["user_id_header"] = ctx.GetHeader(requestuser.UserIdHeaderKey)
			loggingMap["user_role_header"] = ctx.GetHeader(requestuser.UserRoleHeaderKey)
			ctx.AbortWithStatus(http.StatusForbidden)
			return
		}
		ctx.Next()
	}
}

func RegisterHealthcheckHandler(api *gin.RouterGroup) {
	h := &healthcheckHandler{}
	adminM := AdminMiddleware()
	api.GET("", adminM, h.healthCheck)
}

func (h *healthcheckHandler) healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"up":     true,
	})
}
