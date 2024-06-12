package configservice

import (
	requestuser "config-service/pkg/hidepost-requestuser"
	serverlogging "config-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		loggingMap := serverlogging.GetLoggingMap(ctx)
		loggingMap.SetUserId(requestuser.GetUserID(ctx))
		if !requestuser.IsAdmin(ctx) {
			loggingMap.SetMessage("request user is not admin")
			loggingMap["user_role"] = ctx.GetHeader(requestuser.UserRoleHeaderKey)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Next()
	}
}
