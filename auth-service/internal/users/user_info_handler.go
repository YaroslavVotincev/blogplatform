package users

import (
	requestuser "auth-service/pkg/hidepost-requestuser"
	serverlogging "auth-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type userInfoHandler struct {
	service *Service
}

func RegisterUserInfoHandler(api *gin.RouterGroup, service *Service) {
	h := &userInfoHandler{service: service}

	api.GET("", h.getUserInfo)
}

func (h *userInfoHandler) getUserInfo(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)

	userId, err := h.service.ParseToken(ctx.GetHeader(requestuser.AuthorizationHeaderKey))
	if err != nil {
		loggingMap["error"] = err.Error()
		loggingMap.SetMessage("invalid jwt token")
		loggingMap.Debug()
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}
	loggingMap.SetUserId(userId)

	user, err := h.service.ByID(ctx, *userId)
	if err != nil {
		loggingMap["error"] = err.Error()
		loggingMap.SetMessage("fail to get user by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if user == nil {
		loggingMap.SetMessage("user with id of successfully parsed jwt token not found")
		loggingMap.Error()
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}
	if user.Enabled == false {
		loggingMap.SetMessage("user with id of successfully parsed jwt token is not enabled")
		loggingMap.Error()
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}
	if user.Deleted {
		loggingMap.SetMessage("user with id of successfully parsed jwt token is deleted")
		loggingMap.Error()
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}

	loggingMap["message"] = "got user info by jwt token"
	loggingMap.Debug()

	banned := false
	if user.BannedUntil != nil {
		banned = time.Now().UTC().Before(*user.BannedUntil)
	}

	ctx.JSON(http.StatusOK, UserInfoResponse{
		ID:           user.ID,
		Role:         user.Role,
		Login:        user.Login,
		Email:        user.Email,
		IsBanned:     banned,
		BannedReason: user.BannedReason,
	})
}
