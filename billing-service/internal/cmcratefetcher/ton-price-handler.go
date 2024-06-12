package cmcratefetcher

import (
	serverlogging "billing-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

type tonPriceHandler struct {
	service *Service
}

func RegisterTonPriceHandler(api *gin.RouterGroup, service *Service) {
	h := &tonPriceHandler{
		service: service,
	}

	api.GET("", h.tonPrice)
}

func (h *tonPriceHandler) tonPrice(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	price, err := h.service.TonPrice()
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("fail to get ton price")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"rub": price,
	})
}
