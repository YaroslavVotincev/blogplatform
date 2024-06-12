package robokassa

import (
	serverlogging "billing-service/pkg/serverlogging/gin"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type confirmHandler struct {
	service *Service
}

func RegisterConfirmHandler(api *gin.RouterGroup, service *Service) {
	h := &confirmHandler{
		service: service,
	}

	api.GET("/confirm", h.confirm)
}

func (h *confirmHandler) confirm(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	OutSum := ctx.Query("OutSum")
	InvId := ctx.Query("InvId")
	SignatureValue := ctx.Query("SignatureValue")

	invId, err := strconv.Atoi(InvId)
	if err != nil {
		loggingMap.SetError("invalid invoice id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	if !h.service.ConfirmSignatureValueValid(OutSum, InvId, SignatureValue) {
		loggingMap.SetError("invalid confirm signature value")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	invoice, err := h.service.repository.InvoiceById(ctx, invId)
	if err != nil {
		loggingMap.SetMessage("fail to get invoice")
		loggingMap.SetError(err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if invoice == nil {
		loggingMap.SetMessage("invoice not found")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	err = h.service.GrantItemByType(ctx, invoice)
	if err != nil {
		loggingMap["invoice"] = fmt.Sprintf("%+v", *invoice)
		loggingMap.SetMessage("fail to grant item")
		loggingMap.SetError(err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	invoice.Status = InvoiceStatusConfirmed
	invoice.Updated = time.Now().UTC()
	err = h.service.repository.UpdateInvoice(ctx, invoice)
	if err != nil {
		loggingMap.SetMessage("fail to update invoice")
		loggingMap.SetError(err.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("invoice confirmed")
	loggingMap.Info()
	ctx.String(http.StatusOK, "OK%s", InvId)
}
