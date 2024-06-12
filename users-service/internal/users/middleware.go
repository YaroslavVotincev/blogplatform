package users

import (
	"github.com/gin-gonic/gin"
	"net/http"
	requestuser "users-service/pkg/hidepost-requestuser"
	serverlogging "users-service/pkg/serverlogging/gin"
)

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
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Next()
	}
}

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

func ServiceMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		loggingMap := serverlogging.GetLoggingMap(ctx)
		if !requestuser.IsService(ctx) {
			loggingMap.SetMessage("request user is not service")
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		ctx.Next()
	}
}
