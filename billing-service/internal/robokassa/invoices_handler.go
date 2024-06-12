package robokassa

import (
	requestuser "billing-service/pkg/hidepost-requestuser"
	serverlogging "billing-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type invoicesHandler struct {
	service *Service
}

func RegisterInvoicesHandler(api *gin.RouterGroup, service *Service) {
	h := &invoicesHandler{
		service: service,
	}
	userM := UserMiddleware()
	api.GET("/invoices/:id", userM, h.invoiceById)
}

func (h *invoicesHandler) invoiceById(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	invoiceIdStr := ctx.Param("id")
	if invoiceIdStr == "" {
		loggingMap.SetMessage("invoice id param is empty")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["invoice_id_param"] = invoiceIdStr

	invoiceId, err := strconv.Atoi(invoiceIdStr)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get invoice id from param")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	invoices, err := h.service.repository.InvoicesByParams(
		ctx,
		[]int{invoiceId}, nil, nil,
		[]string{userId.String()},
		nil, nil, nil, nil, nil)

	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get invoice")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	if len(invoices) == 0 {
		loggingMap.SetMessage("invoice not found")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	ctx.JSON(http.StatusOK, invoices[0])
}
