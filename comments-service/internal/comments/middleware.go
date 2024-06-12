package comments

import (
	requestuser "comments-service/pkg/hidepost-requestuser"
	serverlogging "comments-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

func UserMiddleware() gin.HandlerFunc {
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
		if !requestuser.IsUser(ctx) {
			loggingMap.SetMessage("request user is not authenticated")
			loggingMap["user_id_header"] = ctx.GetHeader(requestuser.UserIdHeaderKey)
			loggingMap["user_role_header"] = ctx.GetHeader(requestuser.UserRoleHeaderKey)
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Next()
	}
}
