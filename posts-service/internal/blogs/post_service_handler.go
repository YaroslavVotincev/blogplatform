package blogs

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
	serverlogging "posts-service/pkg/serverlogging/gin"
	"time"
)

type postServiceHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterPostServiceHandler(api *gin.RouterGroup, service *Service) {
	h := &postServiceHandler{
		service:  service,
		validate: NewValidator(),
	}
	serviceM := ServiceMiddleware()
	api.POST("/paid-access/grant", serviceM, h.grantPaidAccess)
}

func (h *postServiceHandler) grantPaidAccess(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	timeNow := time.Now().UTC()

	var req GrantItemServiceRequest
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

	post, err := h.service.PostById(ctx, req.ItemId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscription")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetError("post not found")
		loggingMap.SetMessage("failed to get post")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.repository.BlogById(ctx, post.BlogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetError("blog not found")
		loggingMap.SetMessage("blog not found")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blogIncome := BlogIncome{
		ID:               uuid.New(),
		BlogId:           post.BlogId,
		UserId:           req.UserId,
		Value:            req.Value,
		Currency:         req.Currency,
		ItemId:           post.ID,
		ItemType:         PaymentItemTypePost,
		SentToUserWallet: req.Currency == CurrencyTon,
		Created:          timeNow,
	}
	err = h.service.repository.CreateBlogIncome(ctx, &blogIncome)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create blog income")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	paidAccess, err := h.service.repository.PostPaidAccessByPostIdAndUserId(ctx, post.ID, req.UserId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post paid access")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	if paidAccess == nil {
		paidAccess = &PostPaidAccess{
			ID:      uuid.New(),
			PostId:  post.ID,
			UserId:  req.UserId,
			Created: timeNow,
		}
		if err := h.service.repository.CreatePostPaidAccess(ctx, paidAccess); err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to create post paid access")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	}

	go h.service.notifService.PostPaidAccessAuthor(
		blog.AuthorId.String(), blog.ID.String(), post.ID.String(),
		req.UserId.String(), req.Value, req.Currency)

	go h.service.notifService.PostPaidAccessUser(
		req.UserId.String(), blog.ID.String(), post.ID.String(),
		req.Value, req.Currency)

	loggingMap.SetMessage("post paid access granted")
	loggingMap.Info()
	ctx.JSON(http.StatusOK, nil)
	return
}
