package users

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
	serverlogging "users-service/pkg/serverlogging/gin"
)

type infoHandler struct {
	service *Service
}

func RegisterInfoHandler(api *gin.RouterGroup, service *Service) {
	h := &infoHandler{service: service}

	api.GET("/id", h.byIdList)
	api.GET("/id/:id", h.byId)
	api.GET("/id/:id/wallet-address", h.userWalletAddress)
}

func (h *infoHandler) byIdList(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idList := ctx.QueryArray("id")

	users, err := h.service.repository.ByIDList(ctx, idList)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get users by id list")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	usersInfo := make([]UserInfoResponse, len(users))
	for i, user := range users {
		var isBanned = false
		if user.BannedUntil != nil && time.Now().UTC().After(*user.BannedUntil) {
			isBanned = true
		}
		usersInfo[i] = UserInfoResponse{
			ID:           user.ID,
			Login:        user.Login,
			Email:        user.Email,
			Role:         user.Role,
			BannedReason: user.BannedReason,
			IsBanned:     isBanned,
		}
	}

	ctx.JSON(http.StatusOK, usersInfo)
}

func (h *infoHandler) byId(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	loggingMap["req_id_param"] = idParam
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to parse user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	user, err := h.service.ByID(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if user == nil {
		loggingMap.SetMessage("user by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	var isBanned = false
	if user.BannedUntil != nil && time.Now().UTC().After(*user.BannedUntil) {
		isBanned = true
	}

	ctx.JSON(http.StatusOK, UserInfoResponse{
		ID:           user.ID,
		Login:        user.Login,
		Email:        user.Email,
		Role:         user.Role,
		BannedReason: user.BannedReason,
		IsBanned:     isBanned,
	})
}

func (h *infoHandler) userWalletAddress(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	walletObj, err := h.service.WalletByUserId(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get wallet by user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if walletObj == nil {
		walletObj, err = h.service.CreateWalletToUser(ctx, id)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to create wallet to user")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"address": walletObj.Address,
	})
}
