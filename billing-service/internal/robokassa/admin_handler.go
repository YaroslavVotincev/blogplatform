package robokassa

import (
	serverlogging "billing-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

type adminHandler struct {
	service *Service
}

func RegisterAdminHandler(api *gin.RouterGroup, service *Service) {
	h := &adminHandler{
		service: service,
	}
	adminM := AdminMiddleware()
	api.GET("/invoices", adminM, h.invoicesList)
}

func (h *adminHandler) invoicesList(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	invoices, err := h.service.repository.InvoicesByParams(ctx,
		nil, nil, nil, nil, nil, nil, nil, nil, nil)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get invoices list")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, invoices)
}
