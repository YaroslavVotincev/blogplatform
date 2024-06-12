package files

import (
	serverlogging "file-service/pkg/serverlogging/gin"
	"github.com/gin-gonic/gin"
	"net/http"
)

type fileHandler struct {
	service *Service
}

func RegisterFileHandler(api *gin.RouterGroup, service *Service) {
	h := &fileHandler{
		service: service,
	}
	api.GET("/:id", h.get)
	api.DELETE("/:id", h.delete)
	api.POST("/:id", h.upload)
}

func (h *fileHandler) get(ctx *gin.Context) {
	ctx.File(h.service.GetFilePath(ctx.Param("id")))
}

func (h *fileHandler) delete(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	id := ctx.Param("id")
	loggingMap["file_id"] = id
	err := h.service.DeleteFile(id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to delete file")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusNoContent, nil)
}

func (h *fileHandler) upload(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	id := ctx.Param("id")
	loggingMap["file_id"] = id
	data, err := ctx.GetRawData()
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get raw data")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	err = h.service.SetFile(id, data)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set file")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	loggingMap.SetMessage("file set")
	ctx.JSON(http.StatusAccepted, nil)
}
