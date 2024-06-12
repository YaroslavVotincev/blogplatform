package users

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
	serverlogging "users-service/pkg/serverlogging/gin"
)

type serviceHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterServiceHandler(api *gin.RouterGroup, service *Service) {
	h := &serviceHandler{service: service, validate: NewUsersValidator()}

	serviceM := ServiceMiddleware()

	api.PUT("/id/:id/wallet/add-rub", serviceM, h.addRubToWallet)
}

func (h *serviceHandler) addRubToWallet(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	loggingMap["id_param"] = idParam
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to parse user id")
		ctx.JSON(http.StatusInternalServerError, nil)
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

	var req AddRubToBalanceServiceRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["req_body"] = fmt.Sprintf("%+v", req)
	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	walletObj.BalanceRub += req.Value

	err = h.service.repository.UpdateWalletRubBalance(ctx, walletObj)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update wallet")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}
