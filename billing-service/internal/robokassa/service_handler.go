package robokassa

import (
	serverlogging "billing-service/pkg/serverlogging/gin"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type serviceHandler struct {
	service *Service
}

func RegisterServiceHandler(api *gin.RouterGroup, service *Service) {
	h := &serviceHandler{
		service: service,
	}
	serviceM := ServiceMiddleware()
	api.POST("/payment-link", serviceM, h.paymentLink)
}

func (h *serviceHandler) paymentLink(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	var req PaymentLinkRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	invoices, err := h.service.repository.InvoicesByUserIdAndItemId(ctx, req.UserId, req.ItemId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get invoices list")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	if len(invoices) == 0 {
		link, err := h.service.CreateInvoice(ctx, req.ItemId, req.ItemType, req.Description, req.Sum, req.UserId)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("fail to create invoice")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		ctx.JSON(http.StatusOK, PaymentLinkResponse{Url: link})
	} else {
		latestInvoice := invoices[0]
		timeNow := time.Now().UTC()

		if latestInvoice.Status == InvoiceStatusNew {
			if timeNow.After(latestInvoice.ExpiresAt) {
				latestInvoice.Status = InvoiceStatusExpired
				latestInvoice.Updated = timeNow
				err = h.service.repository.UpdateInvoice(ctx, &latestInvoice)
				if err != nil {
					loggingMap.SetError(err.Error())
					loggingMap.SetMessage("fail to update invoice")
					ctx.JSON(http.StatusInternalServerError, nil)
					return
				}
				link, err := h.service.CreateInvoice(ctx, req.ItemId, req.ItemType, req.Description, req.Sum, req.UserId)
				if err != nil {
					loggingMap.SetError(err.Error())
					loggingMap.SetMessage("fail to create invoice")
					ctx.JSON(http.StatusInternalServerError, nil)
					return
				}
				ctx.JSON(http.StatusOK, PaymentLinkResponse{Url: link})
			} else {
				ctx.JSON(http.StatusOK, PaymentLinkResponse{Url: latestInvoice.PaymentLink})
			}
		} else {
			link, err := h.service.CreateInvoice(ctx, req.ItemId, req.ItemType, req.Description, req.Sum, req.UserId)
			if err != nil {
				loggingMap.SetError(err.Error())
				loggingMap.SetMessage("fail to create invoice")
				ctx.JSON(http.StatusInternalServerError, nil)
				return
			}
			ctx.JSON(http.StatusOK, PaymentLinkResponse{Url: link})
		}
	}
}
