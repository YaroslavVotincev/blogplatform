package blogs

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
	requestuser "posts-service/pkg/hidepost-requestuser"
	serverlogging "posts-service/pkg/serverlogging/gin"
	"strconv"
	"strings"
	"time"
)

type blogHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterBlogHandler(api *gin.RouterGroup, service *Service) {
	h := &blogHandler{
		service:  service,
		validate: NewValidator(),
	}

	userM := UserMiddleware()

	api.GET("/categories", h.getCategories)
	api.GET("/categories/my-preference", userM, h.getUserCategoriesPreference)
	api.PUT("/categories/my-preference", userM, h.setUserCategoriesPreference)

	api.GET("/all", h.all)
	api.GET("/id", h.byIdList)
	api.GET("/id/:id", h.byId)
	api.PUT("/id/:id", userM, h.update)
	//api.PUT("/id/:id/title", userM, h.updateTitle)
	//api.PUT("/id/:id/url", userM, h.updateUrl)
	api.GET("/id/:id/avatar", h.getAvatar)
	api.PUT("/id/:id/avatar", userM, h.updateAvatar)
	api.GET("/id/:id/cover", h.getCover)
	api.PUT("/id/:id/cover", userM, h.updateCover)
	//api.PUT("/id/:id/categories", userM, h.updateCategories)
	//api.PUT("/id/:id/status", userM, h.updateStatus)

	api.GET("/id/:id/stats", h.stats)

	api.GET("/id/:id/income", h.getIncome)

	api.GET("/id/:id/goals", h.getGoals)
	api.POST("/id/:id/goals/new", userM, h.createGoal)
	api.GET("/id/:id/goals/id/:goal_id", h.getGoalById)
	api.PUT("/id/:id/goals/id/:goal_id", userM, h.updateGoal)

	api.GET("/id/:id/follow", userM, h.amIFollowing)
	api.POST("/id/:id/follow", userM, h.follow)

	api.GET("/id/:id/subscriptions", h.getBlogSubscriptions)
	api.GET("/id/:id/subscriptions/my", userM, h.getMyBlogUserSubscriptions)
	api.POST("/id/:id/subscriptions/new/free", userM, h.createFreeSubscription)
	api.POST("/id/:id/subscriptions/new/paid", userM, h.createPaidSubscription)

	api.GET("/id/:id/donations", h.getBlogDonations)
	api.POST("/id/:id/donate/robokassa", userM, h.makeDonationRobokassa)
	api.POST("/id/:id/donate/toncoin", userM, h.makeDonationToncoin)
	api.GET("/donations/min-values", h.getMinDonationValues)
	api.GET("/donations/my/id/:id", userM, h.getMyDonationById)

	api.GET("/follows/my", userM, h.getMyFollowedBlogs)

	api.GET("/subscriptions/my", h.getMyUserSubscriptions)
	api.GET("/subscriptions/id", h.blogSubscriptionsByIdList)

	api.POST("/subscriptions/id/:id/subscribe", userM, h.subscribe)
	api.POST("/subscriptions/id/:id/subscribe/free", userM, h.subscribeFree)
	api.POST("/subscriptions/id/:id/subscribe/robokassa", userM, h.subscribeRobokassa)

	api.PUT("/subscriptions/id/:id/info", userM, h.updateSubscriptionInfo)
	api.GET("/subscriptions/id/:id/cover", h.getSubscriptionCover)
	api.PUT("/subscriptions/id/:id/cover", userM, h.updateSubscriptionCover)

	api.GET("/url/:url", h.byUrl)
	api.GET("/url/:url/avatar", h.getAvatar)
	api.GET("/url/:url/cover", h.getCover)

	api.POST("/new/personal", userM, h.newPersonal)
	api.POST("/new/thematic", userM, h.newThematic)
}

func (h *blogHandler) getCategories(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	categories, err := h.service.repository.AllCategories(ctx)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get all categories")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	loggingMap.None()
	ctx.JSON(http.StatusOK, categories)
}

func (h *blogHandler) getUserCategoriesPreference(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	categories, err := h.service.repository.CategoriesByUserPreferences(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user categories preference")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, categories)
}

func (h *blogHandler) setUserCategoriesPreference(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	var categories []string
	if err := json.NewDecoder(ctx.Request.Body).Decode(&categories); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err := h.service.repository.SetUserCategoriesPreferences(ctx, *userId, categories)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set user categories preference")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("user categories preference updated")
	ctx.JSON(http.StatusAccepted, nil)

}

func (h *blogHandler) all(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	userIdQuery := ctx.Query("user_id")
	userId, err := uuid.Parse(userIdQuery)
	loggingMap["user_query"] = userIdQuery
	if err != nil {
		blogs, err := h.service.repository.AllBlogs(ctx)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get all blogs")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		ctx.JSON(http.StatusOK, blogs)
	} else {
		blogs, err := h.service.repository.BlogsByUserId(ctx, userId)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get all blogs by user id")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		ctx.JSON(http.StatusOK, blogs)
	}
}

func (h *blogHandler) byIdList(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idListQuery := ctx.Query("list")
	idList := strings.Split(idListQuery, ",")

	blogs, err := h.service.repository.BlogsByIdList(ctx, idList)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id lis")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, blogs)
}

func (h *blogHandler) byId(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	ctx.JSON(http.StatusOK, blog)
}

func (h *blogHandler) byUrl(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	urlParam := ctx.Param("url")
	loggingMap["url_param"] = urlParam

	blog, err := h.service.BlogByUrl(ctx, urlParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by url")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	if blog == nil {
		loggingMap.Debug()
		loggingMap.SetMessage("blog by url doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	ctx.JSON(http.StatusOK, blog)
}

func (h *blogHandler) newPersonal(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	blog, err := h.service.NewPersonalBlog(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create new personal blog")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	loggingMap.SetMessage("new personal blog created")
	ctx.JSON(http.StatusCreated, blog)
}

func (h *blogHandler) newThematic(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	blog, err := h.service.NewThematicBlog(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create new thematic blog")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	loggingMap.SetMessage("new thematic blog created")
	ctx.JSON(http.StatusCreated, blog)
}

func (h *blogHandler) update(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	var req BlogUpdateRequest
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

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	blogByUrl, err := h.service.BlogByUrl(ctx, req.Url)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by url")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blogByUrl != nil && blogByUrl.ID != blog.ID {
		loggingMap.SetMessage("blog by url already exists")
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	err = h.service.repository.SetBlogCategories(ctx, id, req.Categories)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set blog categories")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	blog.Title = req.Title
	blog.ShortDescription = req.ShortDescription
	blog.Url = req.Url
	blog.Status = req.Status
	blog.AcceptDonations = req.AcceptDonations
	blog.Updated = time.Now().UTC()

	err = h.service.repository.UpdateBlog(ctx, blog)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update blog")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) updateTitle(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	var req BlogUpdateTitleRequest
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

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	blog.Title = req.Title
	blog.ShortDescription = req.ShortDescription
	blog.Updated = time.Now().UTC()

	err = h.service.repository.UpdateBlog(ctx, blog)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update blog")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) updateUrl(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	var req BlogUpdateUrlRequest
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

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	blogByUrl, err := h.service.BlogByUrl(ctx, req.Url)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by url")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blogByUrl != nil && blogByUrl.ID != blog.ID {
		loggingMap.SetMessage("blog by url already exists")
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	blog.Url = req.Url
	blog.Updated = time.Now().UTC()

	err = h.service.repository.UpdateBlog(ctx, blog)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update blog")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) getContent(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	content, err := h.service.ContentById(ctx, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get content by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog updated")
	ctx.JSON(http.StatusOK, content)
}

func (h *blogHandler) updateContent(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	var req BlogUpdateContentRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	//loggingMap["req_body"] = fmt.Sprintf("%+v", req)
	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	content, err := h.service.ContentById(ctx, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get content by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	content.Updated = time.Now().UTC()
	content.DataHtml = req.DataHtml
	content.DataJson = req.DataJson

	err = h.service.repository.UpdateContent(ctx, content)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update blog content")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) updateAcceptDonations(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	var req BlogUpdateAcceptDonationsRequest
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

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	blog.AcceptDonations = req.AcceptDonations
	blog.Updated = time.Now().UTC()

	err = h.service.repository.UpdateBlog(ctx, blog)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update blog")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) getAvatar(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	var blog *Blog
	var err error

	idParam := ctx.Param("id")
	urlParam := ctx.Param("url")
	if idParam != "" {
		loggingMap["param_id"] = idParam
		id, err := uuid.Parse(idParam)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("incorrect param id")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}
		blog, err = h.service.BlogById(ctx, id)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get blog by id")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	} else {
		loggingMap["param_url"] = urlParam
		blog, err = h.service.BlogByUrl(ctx, urlParam)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get blog by url")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	}

	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	if blog.Avatar == nil {
		loggingMap.Debug()
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	redirectUrl, err := h.service.RedirectUrlToFile(*blog.Avatar)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to redirect url to avatar")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.None()
	ctx.Redirect(http.StatusSeeOther, redirectUrl)
}

func (h *blogHandler) updateAvatar(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	avatarBytes, err := ctx.GetRawData()
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get avatar from request body")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	contentType := http.DetectContentType(avatarBytes)
	if contentType != "image/jpeg" && contentType != "image/png" {
		loggingMap.SetMessage("wrong avatar file type")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err = h.service.SetBlogAvatar(ctx, blog, avatarBytes)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set blog avatar")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog avatar updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) getCover(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	var blog *Blog
	var err error

	idParam := ctx.Param("id")
	urlParam := ctx.Param("url")
	if idParam != "" {
		loggingMap["param_id"] = idParam
		id, err := uuid.Parse(idParam)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("incorrect param id")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}
		blog, err = h.service.BlogById(ctx, id)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get blog by id")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	} else {
		loggingMap["param_url"] = urlParam
		blog, err = h.service.BlogByUrl(ctx, urlParam)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get blog by url")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	}

	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	if blog.Cover == nil {
		loggingMap.Debug()
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	redirectUrl, err := h.service.RedirectUrlToFile(*blog.Cover)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to redirect url to cover")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.None()
	ctx.Redirect(http.StatusSeeOther, redirectUrl)
}

func (h *blogHandler) updateCover(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	coverBytes, err := ctx.GetRawData()
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get cover from request body")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	contentType := http.DetectContentType(coverBytes)
	if contentType != "image/jpeg" && contentType != "image/png" {
		loggingMap.SetMessage("wrong cover file type")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err = h.service.SetBlogCover(ctx, blog, coverBytes)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set blog cover")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog cover updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) updateCategories(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	var categories []string
	if err := json.NewDecoder(ctx.Request.Body).Decode(&categories); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err = h.service.repository.SetBlogCategories(ctx, id, categories)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set blog categories")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog categories updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) updateStatus(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	var req BlogUpdateStatusRequest
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

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	blog.Status = req.Status
	blog.Updated = time.Now().UTC()

	err = h.service.repository.UpdateBlog(ctx, blog)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update blog")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("blog updated")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) getBlogSubscriptions(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	subscriptions, err := h.service.repository.SubscriptionsByBlogId(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscriptions by blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, subscriptions)
}

func (h *blogHandler) blogSubscriptionsByIdList(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idListQuery := ctx.Query("list")
	idList := strings.Split(idListQuery, ",")

	blogSubscriptions, err := h.service.repository.SubscriptionsByIdList(ctx, unique(idList))
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog subscriptions by id list")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, blogSubscriptions)
}

func (h *blogHandler) createFreeSubscription(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	subscriptions, err := h.service.repository.SubscriptionsByBlogId(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscriptions by blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	blogHaveFreeSubscription := false
	for _, subscription := range subscriptions {
		if subscription.IsFree {
			blogHaveFreeSubscription = true
			break
		}
	}
	if blogHaveFreeSubscription {
		loggingMap.SetMessage("free subscription already exists")
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	subscription := &Subscription{
		ID:               uuid.New(),
		BlogId:           id,
		Title:            "Бесплатная подписка",
		ShortDescription: "Бесплатная подписка",
		IsFree:           true,
		PriceRub:         0,
		Cover:            nil,
		IsActive:         true,
		Created:          time.Now().UTC(),
		Updated:          time.Now().UTC(),
	}

	err = h.service.repository.CreateSubscription(ctx, subscription)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create free subscription")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("free subscription created")
	ctx.JSON(http.StatusCreated, subscription)
}

func (h *blogHandler) createPaidSubscription(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	priceStr := ctx.Query("price")
	loggingMap["price_query_param"] = priceStr
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param price")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	subscriptions, err := h.service.repository.SubscriptionsByBlogId(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscriptions by blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	for _, subscription := range subscriptions {
		if !subscription.IsFree && subscription.PriceRub == price {
			loggingMap.SetMessage("paid subscription by this price already exists")
			ctx.JSON(http.StatusConflict, nil)
			return
		}
	}

	subscription := &Subscription{
		ID:               uuid.New(),
		BlogId:           id,
		Title:            "Платная подписка",
		ShortDescription: "Платная подписка",
		IsFree:           false,
		PriceRub:         price,
		Cover:            nil,
		IsActive:         false,
		Created:          time.Now().UTC(),
		Updated:          time.Now().UTC(),
	}

	err = h.service.repository.CreateSubscription(ctx, subscription)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create paid subscription")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("paid subscription created")
	ctx.JSON(http.StatusCreated, subscription)
}

func (h *blogHandler) updateSubscriptionInfo(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	var req SubscriptionUpdateInfoRequest
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

	subscription, err := h.service.repository.SubscriptionById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscriptions by blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if subscription == nil {
		loggingMap.SetMessage("subscription by blog id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	blog, err := h.service.BlogById(ctx, subscription.BlogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	subscription.Updated = time.Now().UTC()
	subscription.IsActive = req.IsActive
	subscription.ShortDescription = req.ShortDescription
	subscription.Title = req.Title
	subscription.PriceRub = req.PriceRub

	err = h.service.repository.UpdateSubscription(ctx, subscription)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update subscription")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, subscription)
}

func (h *blogHandler) updateSubscriptionCover(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	subscription, err := h.service.repository.SubscriptionById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscription by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if subscription == nil {
		loggingMap.SetMessage("subscription by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	blog, err := h.service.BlogById(ctx, subscription.BlogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	coverBytes, err := ctx.GetRawData()
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get cover from request body")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	contentType := http.DetectContentType(coverBytes)
	if contentType != "image/jpeg" && contentType != "image/png" {
		loggingMap.SetMessage("wrong cover file type")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	err = h.service.SetSubscriptionCover(ctx, subscription, coverBytes)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set subscription cover")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, subscription)
}

func (h *blogHandler) getSubscriptionCover(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	subscription, err := h.service.repository.SubscriptionById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscriptions by blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if subscription == nil {
		loggingMap.SetMessage("subscription by blog id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if subscription.Cover == nil {
		loggingMap.SetMessage("subscription cover doesn't exists")
		loggingMap.Debug()
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	redirectUrl, err := h.service.RedirectUrlToFile(*subscription.Cover)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to redirect url to cover")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.None()
	ctx.Redirect(http.StatusSeeOther, redirectUrl)
}

func (h *blogHandler) getMyBlogUserSubscriptions(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	if userId == nil {
		loggingMap.SetMessage("request user is not authorized")
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}

	idParam := ctx.Param("id")
	loggingMap["id_param"] = idParam
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	userSubscriptions, err := h.service.repository.UserSubscriptionsByUserIdAndBlogId(ctx, *userId, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user subscriptions by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, userSubscriptions)
}

func (h *blogHandler) getMyUserSubscriptions(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	if userId == nil {
		loggingMap.SetMessage("request user is not authorized")
		ctx.JSON(http.StatusUnauthorized, nil)
		return
	}

	userSubscriptions, err := h.service.repository.UserSubscriptionsByUserId(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user subscriptions by user id ")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, userSubscriptions)
}

func (h *blogHandler) subscribe(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	subscription, err := h.service.repository.SubscriptionById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscription by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if subscription == nil {
		loggingMap.SetMessage("subscription by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	blog, err := h.service.BlogById(ctx, subscription.BlogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId == *userId {
		loggingMap.SetMessage("request user is the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	follow, err := h.service.repository.UserFollowByUserIdAndBlogId(ctx, *userId, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get follow by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if follow == nil {
		follow = &UserFollow{
			ID:      uuid.New(),
			UserId:  *userId,
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

	userSubscription, err := h.service.repository.UserSubscriptionByParams(ctx, *userId, blog.ID, subscription.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user subscriptions by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if subscription.IsFree {
		if userSubscription == nil {
			newUserSubscription := UserSubscription{
				ID:             uuid.New(),
				UserId:         *userId,
				SubscriptionId: subscription.ID,
				BlogId:         blog.ID,
				Status:         UserSubscriptionStatusLifetime,
				IsActive:       true,
				ExpiresAt:      nil,
				Created:        time.Now().UTC(),
				Updated:        time.Now().UTC(),
			}
			err = h.service.repository.CreateUserSubscription(ctx, &newUserSubscription)
			if err != nil {
				loggingMap.SetError(err.Error())
				loggingMap.SetMessage("failed to create user subscription")
				ctx.JSON(http.StatusInternalServerError, nil)
				return
			}
			ctx.JSON(http.StatusOK, struct{}{})
		}
	} else {
		if userSubscription != nil {
			if userSubscription.IsActive {
				loggingMap.SetMessage("user is already subscribed")
				ctx.JSON(http.StatusOK, struct{}{})
			} else {
				link, err := h.service.GetSubscriptionRobokassaPaymentLink(ctx, subscription, *userId)
				if err != nil {
					loggingMap.SetError(err.Error())
					loggingMap.SetMessage("failed to get payment link")
					ctx.JSON(http.StatusInternalServerError, nil)
					return
				}
				ctx.JSON(http.StatusOK, gin.H{"payment_link": link})
			}
		} else {
			link, err := h.service.GetSubscriptionRobokassaPaymentLink(ctx, subscription, *userId)
			if err != nil {
				loggingMap.SetError(err.Error())
				loggingMap.SetMessage("failed to get payment link")
				ctx.JSON(http.StatusInternalServerError, nil)
				return
			}
			ctx.JSON(http.StatusOK, gin.H{"payment_link": link})
		}
	}
}

func (h *blogHandler) subscribeFree(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	subscription, err := h.service.repository.SubscriptionById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscription by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if subscription == nil {
		loggingMap.SetMessage("subscription by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if !subscription.IsFree {
		loggingMap.SetMessage("subscription is not free")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}
	blog, err := h.service.BlogById(ctx, subscription.BlogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId == *userId {
		loggingMap.SetMessage("request user is the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	follow, err := h.service.repository.UserFollowByUserIdAndBlogId(ctx, *userId, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get follow by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if follow == nil {
		follow = &UserFollow{
			ID:      uuid.New(),
			UserId:  *userId,
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

	userSubscription, err := h.service.repository.UserSubscriptionByParams(ctx, *userId, blog.ID, subscription.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user subscriptions by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	if userSubscription == nil {
		newUserSubscription := UserSubscription{
			ID:             uuid.New(),
			UserId:         *userId,
			SubscriptionId: subscription.ID,
			BlogId:         blog.ID,
			Status:         UserSubscriptionStatusLifetime,
			IsActive:       true,
			ExpiresAt:      nil,
			Created:        time.Now().UTC(),
			Updated:        time.Now().UTC(),
		}
		err = h.service.repository.CreateUserSubscription(ctx, &newUserSubscription)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to create user subscription")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
	}
	ctx.JSON(http.StatusOK, nil)
}

func (h *blogHandler) subscribeRobokassa(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	subscription, err := h.service.repository.SubscriptionById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscription by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if subscription == nil {
		loggingMap.SetMessage("subscription by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if subscription.IsFree {
		loggingMap.SetMessage("subscription is free")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, subscription.BlogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId == *userId {
		loggingMap.SetMessage("request user is the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	userSubscription, err := h.service.repository.UserSubscriptionByParams(ctx, *userId, blog.ID, subscription.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user subscriptions by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if userSubscription != nil {
		if userSubscription.IsActive {
			loggingMap.SetMessage("user is already subscribed")
			ctx.JSON(http.StatusConflict, nil)
			return
		}
	}
	link, err := h.service.GetSubscriptionRobokassaPaymentLink(ctx, subscription, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get payment link")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"payment_link": link})

}

func (h *blogHandler) amIFollowing(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId == *userId {
		loggingMap.SetMessage("request user is the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	follow, err := h.service.repository.UserFollowByUserIdAndBlogId(ctx, *userId, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get follow by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if follow == nil {
		ctx.JSON(http.StatusNoContent, nil)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}

func (h *blogHandler) follow(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId == *userId {
		loggingMap.SetMessage("request user is the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	follow, err := h.service.repository.UserFollowByUserIdAndBlogId(ctx, *userId, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get follow by user id and blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if follow == nil {
		follow = &UserFollow{
			ID:      uuid.New(),
			UserId:  *userId,
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
		ctx.JSON(http.StatusCreated, nil)
		return
	}

	ctx.JSON(http.StatusCreated, nil)
}

func (h *blogHandler) getMyFollowedBlogs(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	userFollows, err := h.service.repository.UserFollowsByUserId(ctx, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user follows")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	blogIds := make([]string, 0)
	for _, userFollow := range userFollows {
		blogIds = append(blogIds, userFollow.BlogId.String())
	}

	blogs, err := h.service.repository.BlogsByIdList(ctx, unique(blogIds))
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blogs")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, blogs)
}

func (h *blogHandler) getGoals(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	goals, err := h.service.repository.GoalsByBlogId(ctx, id)

	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog goals")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, goals)
}

func (h *blogHandler) createGoal(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	var req CreateGoalRequest
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

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	newGoal := &Goal{
		ID:          uuid.New(),
		BlogId:      id,
		Type:        req.Type,
		Description: req.Description,
		Target:      req.Target,
		Current:     0,
		Created:     time.Now().UTC(),
		Updated:     time.Now().UTC(),
	}

	err = h.service.repository.CreateGoal(ctx, newGoal)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create goal")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusCreated, newGoal)
}

func (h *blogHandler) getGoalById(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("goal_id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	goal, err := h.service.repository.GoalsById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog goals")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, goal)
}

func (h *blogHandler) updateGoal(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	blogIdParam := ctx.Param("id")
	blogId, err := uuid.Parse(blogIdParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	idParam := ctx.Param("goal_id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	var req UpdateGoalRequest
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

	blog, err := h.service.BlogById(ctx, blogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	goal, err := h.service.repository.GoalsById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog goals")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	goal.Updated = time.Now().UTC()
	goal.Target = req.Target
	goal.Description = req.Description

	err = h.service.repository.UpdateGoal(ctx, goal)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update goal")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, goal)
}

func (h *blogHandler) stats(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blogPosts, err := h.service.repository.PostsByBlogId(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog posts")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	commentsCount := 0
	postsCount := 0
	for _, post := range blogPosts {
		if post.Status == PostStatusPublic {
			commentsCount += post.CommentsCount
			postsCount++
		}
	}

	followsCount, err := h.service.repository.CountBlogUserFollows(ctx, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to count blog user follows")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	paidSubsCount, err := h.service.repository.CountBlogPaidUserSubscriptions(ctx, blog.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to count blog paid user subscriptions")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, BlogStatsResponse{
		CommentsCount:        commentsCount,
		PostsCount:           postsCount,
		FollowersCount:       followsCount,
		PaidSubscribersCount: paidSubsCount,
	})

}

func (h *blogHandler) makeDonationRobokassa(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	var req MakeDonationRequest
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

	if req.Value < h.service.DonationsRobokassaMinValue() {
		loggingMap.SetMessage("value is below minimum robokassa required value")
		loggingMap["min_value"] = h.service.donationsRobokassaMinValue
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId == *userId {
		loggingMap.SetMessage("cannot donate to your own blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	timeNow := time.Now().UTC()
	donation := Donation{
		ID:               uuid.New(),
		UserId:           *userId,
		BlogId:           id,
		Value:            req.Value,
		Currency:         CurrencyRub,
		UserComment:      req.Comment,
		Status:           DonationStatusNew,
		PaymentConfirmed: false,
		Created:          timeNow,
		Updated:          timeNow,
	}

	err = h.service.repository.CreateDonation(ctx, &donation)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create donation")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	paymentLink, err := h.service.GetDonationRobokassaPaymentLink(blog, &donation, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get robokassa payment link")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{
		"payment_link": paymentLink,
	})
}

func (h *blogHandler) makeDonationToncoin(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	var req MakeDonationRequest
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

	if req.Value < h.service.DonationsToncoinMinValue() {
		loggingMap.SetMessage("value is below minimum toncoin required value")
		loggingMap["min_value"] = h.service.donationsRobokassaMinValue
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId == *userId {
		loggingMap.SetMessage("cannot donate to your own blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	timeNow := time.Now().UTC()
	donation := Donation{
		ID:               uuid.New(),
		UserId:           *userId,
		BlogId:           id,
		Value:            req.Value,
		Currency:         CurrencyTon,
		UserComment:      req.Comment,
		Status:           DonationStatusNew,
		PaymentConfirmed: false,
		Created:          timeNow,
		Updated:          timeNow,
	}

	err = h.service.repository.CreateDonation(ctx, &donation)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create donation")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}

func (h *blogHandler) getMinDonationValues(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, gin.H{
		"rub":     h.service.donationsRobokassaMinValue,
		"toncoin": h.service.donationsToncoinMinValue,
	})
}

func (h *blogHandler) getMyDonationById(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	donation, err := h.service.repository.DonationById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get donation by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if donation == nil {
		loggingMap.SetMessage("donation by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	if donation.UserId != *userId {
		loggingMap.SetMessage("donation by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	ctx.JSON(http.StatusOK, donation)
}

func (h *blogHandler) getIncome(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	blogIncomes, err := h.service.repository.BlogIncomesByBlogId(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog incomes")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, blogIncomes)
}

func (h *blogHandler) getBlogDonations(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	donations, err := h.service.repository.DonationsByBlogIdConfirmed(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get donation by blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, donations)
}
