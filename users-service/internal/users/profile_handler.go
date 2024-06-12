package users

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"users-service/pkg/cryptservice"
	requestuser "users-service/pkg/hidepost-requestuser"
	serverlogging "users-service/pkg/serverlogging/gin"
)

type profileHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterProfileHandler(api *gin.RouterGroup, service *Service) {
	h := &profileHandler{service: service, validate: NewUsersValidator()}
	userM := UserMiddleware()

	api.GET("/my/fio", userM, h.getMyFio)
	api.PUT("/my/fio", userM, h.updateMyFio)
	api.GET("/my/wallet", userM, h.myWallet)
	api.PUT("/my/password", userM, h.myPassword)
	api.GET("/my/avatar", userM, h.getMyAvatar)
	api.POST("/my/avatar", userM, h.setMyAvatar)

	api.GET("/login/:login/avatar", h.byLoginGetAvatar)
}

func (h *profileHandler) getMyFio(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	profile, err := h.service.ProfileById(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get profile by user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if profile == nil {
		loggingMap.SetMessage("profile by user id doesn't exists")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, MyProfileFioRequest{
		FirstName:  profile.FirstName,
		LastName:   profile.LastName,
		MiddleName: profile.MiddleName,
	})
}

func (h *profileHandler) updateMyFio(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	var req MyProfileFioRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad update my profile fio request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["req_body"] = fmt.Sprintf("%+v", req)
	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad update my profile fio request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	profile, err := h.service.ProfileById(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get profile by user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if profile == nil {
		loggingMap.SetMessage("profile by user id doesn't exists")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	err = h.service.UpdateProfileFromFioRequest(ctx, profile, &req)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update profile from fio request")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *profileHandler) myWallet(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	walletObj, err := h.service.WalletByUserId(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get wallet by user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if walletObj == nil {
		walletObj, err = h.service.CreateWalletToUser(ctx, *userId)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to create wallet to user")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	}

	balance, err := h.service.GetWalletBalance(walletObj.Address)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get balance")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, MyProfileWalletResponse{
		BalanceTon: balance,
		BalanceRub: walletObj.BalanceRub,
		Address:    walletObj.Address,
	})
}

func (h *profileHandler) myPassword(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	var req ChangeMyPasswordRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad change my password request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad change my password request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	hashedPassword, err := h.service.ByIdHashedPassword(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get hashed password by user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if hashedPassword == nil {
		loggingMap.SetMessage("user by id authorized, but doesn't exists")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	oldMatched := cryptservice.ValueHashMatched(req.Old, *hashedPassword)
	if oldMatched == false {
		loggingMap.SetMessage("wrong old password")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err = h.service.UpdatePassword(ctx, *userId, req.New)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update password")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("user password updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *profileHandler) getMyAvatar(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	profile, err := h.service.ProfileById(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get profile by user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if profile == nil {
		loggingMap.SetMessage("profile by user id doesn't exists")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if profile.Avatar == nil {
		loggingMap.Debug()
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	redirectUrl, err := h.service.RedirectUrlToAvatar(*profile.Avatar)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to redirect url to avatar")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.None()
	ctx.Redirect(http.StatusSeeOther, redirectUrl)
}

func (h *profileHandler) setMyAvatar(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	profile, err := h.service.ProfileById(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get profile by user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if profile == nil {
		loggingMap.SetMessage("profile by user id doesn't exists")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	avatarBytes, err := ctx.GetRawData()
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get avatar from request body")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	contentType := http.DetectContentType(avatarBytes)
	if contentType != "image/jpeg" && contentType != "image/png" {
		loggingMap.SetMessage("wrong avatar file type")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err = h.service.SetProfileAvatar(ctx, profile, avatarBytes)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set profile avatar")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.Info()
	loggingMap.SetMessage("profile avatar set")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *profileHandler) byLoginGetAvatar(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	user, err := h.service.ByLogin(ctx, ctx.Param("login"))
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user by login")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if user == nil {
		loggingMap.SetMessage("user by login doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	profile, err := h.service.ProfileById(ctx, user.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get profile by user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if profile == nil {
		loggingMap.SetMessage("profile by user id doesn't exists")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if profile.Avatar == nil {
		loggingMap.Debug()
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	redirectUrl, err := h.service.RedirectUrlToAvatar(*profile.Avatar)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to redirect url to avatar")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.None()
	ctx.Redirect(http.StatusSeeOther, redirectUrl)
}
