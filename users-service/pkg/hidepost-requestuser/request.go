package requestuser

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserRoleUnknown   = "unknown"
	UserRoleUser      = "user"
	UserRoleModerator = "moderator"
	UserRoleAdmin     = "admin"
	UserRoleService   = "service"

	UserRoleHeaderKey     = "USER-ROLE"
	UserIdHeaderKey       = "USER-ID"
	UserIsBannedHeaderKey = "USER-BANNED"
)

func IsUnknown(ctx *gin.Context) bool {
	switch ctx.GetHeader(UserRoleHeaderKey) {
	case UserRoleUser, UserRoleModerator, UserRoleAdmin, UserRoleService:
		return false
	default:
		return true
	}

}

func IsUser(ctx *gin.Context) bool {
	switch ctx.GetHeader(UserRoleHeaderKey) {
	case UserRoleUser, UserRoleModerator, UserRoleAdmin:
		return true
	default:
		return false
	}
}

func IsModerator(ctx *gin.Context) bool {
	switch ctx.GetHeader(UserRoleHeaderKey) {
	case UserRoleModerator, UserRoleAdmin:
		return true
	default:
		return false
	}
}

func IsAdmin(ctx *gin.Context) bool {
	return ctx.GetHeader(UserRoleHeaderKey) == UserRoleAdmin
}

func IsService(ctx *gin.Context) bool {
	return ctx.GetHeader(UserRoleHeaderKey) == UserRoleService
}

func GetUserID(ctx *gin.Context) *uuid.UUID {
	userID, err := uuid.Parse(ctx.GetHeader(UserIdHeaderKey))
	if err != nil {
		return nil
	}
	return &userID
}

func IsBanned(ctx *gin.Context) bool {
	return ctx.GetHeader(UserIsBannedHeaderKey) == "true"
}
