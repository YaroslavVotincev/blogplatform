package comments

import (
	requestuser "comments-service/pkg/hidepost-requestuser"
	serverlogging "comments-service/pkg/serverlogging/gin"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
)

type commentsHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterCommentsHandler(api *gin.RouterGroup, service *Service) {
	h := &commentsHandler{
		service:  service,
		validate: NewValidator(),
	}

	userM := UserMiddleware()

	api.GET("/parent/:id", h.all)
	api.POST("/parent/:id", userM, h.create)
	api.GET("/parent/:id/count", h.count)
}

func (h *commentsHandler) all(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	parentIdParam := ctx.Param("id")
	loggingMap["parent_id"] = parentIdParam
	parentId, err := uuid.Parse(parentIdParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param parent_id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	comments, err := h.service.ByParentId(ctx, parentId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get comments by parent id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, comments)
}

func (h *commentsHandler) create(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	parentIdParam := ctx.Param("id")
	loggingMap["parent_id"] = parentIdParam
	parentId, err := uuid.Parse(parentIdParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param parent_id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	var req CommentCreateRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	_, err = h.service.CreateFromRequest(ctx, parentId, *userId, &req)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create comment")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusCreated, nil)
}

func (h *commentsHandler) count(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	parentIdParam := ctx.Param("id")
	loggingMap["parent_id"] = parentIdParam
	parentId, err := uuid.Parse(parentIdParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param parent_id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	amount, err := h.service.repository.ByParentId2LevelsCount(ctx, parentId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get comments by parent id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"amount": amount})
}
