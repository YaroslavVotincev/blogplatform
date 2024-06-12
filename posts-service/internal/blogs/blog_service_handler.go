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

type blogServiceHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterBlogServiceHandler(api *gin.RouterGroup, service *Service) {
	h := &blogServiceHandler{
		service:  service,
		validate: NewValidator(),
	}
	serviceM := ServiceMiddleware()
	api.POST("/subscriptions/grant", serviceM, h.grantSubscription)
	api.POST("/donations/confirm", serviceM, h.confirmDonation)
}

func (h *blogServiceHandler) grantSubscription(ctx *gin.Context) {
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

	subscription, err := h.service.repository.SubscriptionById(ctx, req.ItemId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscription")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if subscription == nil {
		loggingMap.SetError("subscription not found")
		loggingMap.SetMessage("subscription not found")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.repository.BlogById(ctx, subscription.BlogId)
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
		BlogId:           subscription.BlogId,
		UserId:           req.UserId,
		Value:            req.Value,
		Currency:         req.Currency,
		ItemId:           subscription.ID,
		ItemType:         PaymentItemTypeSubscription,
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

	userSubscription, err := h.service.repository.UserSubscriptionByParams(ctx, req.UserId, subscription.BlogId, subscription.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user subscription")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	if userSubscription == nil {
		expiryDate := time.Now().Add(time.Hour * 720).UTC()
		userSubscription := UserSubscription{
			ID:             uuid.New(),
			UserId:         req.UserId,
			SubscriptionId: subscription.ID,
			BlogId:         subscription.BlogId,
			Status:         UserSubscriptionStatusCancelled,
			IsActive:       true,
			ExpiresAt:      &expiryDate,
			Created:        timeNow,
			Updated:        timeNow,
		}
		err = h.service.repository.CreateUserSubscription(ctx, &userSubscription)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to create user subscription")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	} else {
		expiryDate := time.Now().Add(time.Hour * 720).UTC()
		userSubscription.ExpiresAt = &expiryDate
		userSubscription.Status = UserSubscriptionStatusCancelled
		userSubscription.Updated = timeNow
		userSubscription.IsActive = true

		err = h.service.repository.UpdateUserSubscription(ctx, userSubscription)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to update user subscription")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	}

	follow, err := h.service.repository.UserFollowByUserIdAndBlogId(ctx, req.UserId, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get follow by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if follow == nil {
		follow = &UserFollow{
			ID:      uuid.New(),
			UserId:  req.UserId,
			BlogId:  blog.ID,
			Created: time.Now().UTC(),
			Updated: time.Now().UTC(),
		}
		err = h.service.repository.CreateUserFollow(ctx, follow)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to create user follow")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	}

	go h.service.notifService.SubscriptionAuthor(
		blog.AuthorId.String(), blog.ID.String(), subscription.ID.String(),
		req.UserId.String(), req.Value, req.Currency)

	go h.service.notifService.SubscriptionUser(
		req.UserId.String(), blog.ID.String(), subscription.ID.String(),
		req.Value, req.Currency)

	loggingMap.SetMessage("subscription granted")
	loggingMap.Info()
	ctx.JSON(http.StatusOK, nil)
	return
}

func (h *blogServiceHandler) confirmDonation(ctx *gin.Context) {
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

	donation, err := h.service.repository.DonationById(ctx, req.ItemId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get donation by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if donation == nil {
		loggingMap.SetError("donation not found")
		loggingMap.SetMessage("donation not found")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.repository.BlogById(ctx, donation.BlogId)
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
		BlogId:           donation.BlogId,
		UserId:           req.UserId,
		Value:            req.Value,
		Currency:         req.Currency,
		ItemId:           donation.ID,
		ItemType:         PaymentItemTypeDonation,
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

	donation.Status = DonationStatusConfirmed
	donation.PaymentConfirmed = true
	donation.Updated = timeNow

	err = h.service.repository.UpdateDonation(ctx, donation)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update donation")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	go h.service.notifService.DonationAuthor(
		blog.AuthorId.String(), blog.ID.String(), req.UserId.String(),
		donation.UserComment, req.Value, req.Currency)

	go h.service.notifService.DonationUser(
		req.UserId.String(), blog.ID.String(),
		req.Value, req.Currency)

	loggingMap.SetMessage("donation confirmed")
	loggingMap.Info()
	ctx.JSON(http.StatusOK, nil)
	return

}
