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
	"strings"
	"time"
)

type postHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterPostHandler(api *gin.RouterGroup, service *Service) {
	h := &postHandler{
		service:  service,
		validate: NewValidator(),
	}

	userM := UserMiddleware()

	api.GET("/all", h.all)
	api.GET("/id/:id", h.byId)
	api.PUT("/id/:id", userM, h.update)
	//api.PUT("/id/:id/title", userM, h.updateTitle)
	//api.PUT("/id/:id/url", userM, h.updateUrl)
	api.PUT("/id/:id/cover", userM, h.updateCover)
	//api.PUT("/id/:id/tags", userM, h.updateTags)
	//api.PUT("/id/:id/status", userM, h.updateStatus)
	//api.PUT("/id/:id/access-mode", userM, h.updateAccessMode)

	api.GET("/id/:id/content", h.getContent)
	api.GET("/id/:id/content/my-access", h.checkMyContentAccess)
	api.PUT("/id/:id/content", userM, h.updateContent)
	api.POST("/id/:id/content/upload", userM, h.uploadContentFile)
	api.DELETE("/id/:id/content/file/:file_id", h.deleteContentFile)

	api.POST("/id/:id/paid-access/robokassa", userM, h.buyPaidAccessRobokassa)

	api.GET("/id/:id/likes", h.getPostLikesInfo)
	api.POST("/id/:id/likes/like", userM, h.like)
	api.POST("/id/:id/likes/dislike", userM, h.dislike)
	api.POST("/id/:id/likes/unset", userM, h.unsetLike)

	api.GET("/id/:id/add-anon-view", h.addAnonView)

	api.GET("/content/file/:file_id", h.getContentFile)

	api.GET("/follows", userM, h.byFollowedBlogs)

	api.GET("/blog/id/:blog_id", h.byBlogID)
	api.POST("/blog/id/:blog_id/new", userM, h.create)
	api.GET("/blog/id/:blog_id/url/:post_url", h.byUrl)
	api.GET("/blog/url/:blog_url/url/:post_url", h.byUrl)
	api.GET("/blog/url/:blog_url/url/:post_url/cover", h.getCover)

	api.GET("/category", h.byCategories)
}

func (h *postHandler) all(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	posts, err := h.service.repository.AllPosts(ctx)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get all posts")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, posts)
}

func (h *postHandler) byId(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (h *postHandler) byBlogID(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	blogIdParam := ctx.Param("blog_id")
	blogId, err := uuid.Parse(blogIdParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param blog_id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	posts, err := h.service.repository.PostsByBlogId(ctx, blogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get posts by blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, posts)
}

func (h *postHandler) create(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	userId := requestuser.GetUserID(ctx)

	blogIdParam := ctx.Param("blog_id")
	blogId, err := uuid.Parse(blogIdParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param blog_id")
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

	post, err := h.service.CreatePostToBlog(ctx, blogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusCreated, post)
}

func (h *postHandler) update(ctx *gin.Context) {
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

	var req PostUpdateRequest
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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	postByUrl, err := h.service.repository.PostsByBlogIdAndUrl(ctx, blog.ID, req.Url)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by url")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if postByUrl != nil && postByUrl.ID != post.ID {
		loggingMap.SetMessage("post by url already exists")
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	post.AccessMode = req.AccessMode
	switch req.AccessMode {
	case "1":
		post.Price = nil
		post.SubscriptionId = nil
	case "2":
		post.Price = nil
		post.SubscriptionId = nil
	case "3":
		post.Price = nil
		if req.SubscriptionId == nil {
			post.SubscriptionId = nil
		} else {
			subscriptionsByBlog, err := h.service.repository.SubscriptionsByBlogId(ctx, blog.ID)
			if err != nil {
				loggingMap.SetError(err.Error())
				loggingMap.SetMessage("failed to get subscriptions by blog id")
				ctx.JSON(http.StatusInternalServerError, nil)
				return
			}
			subFound := false
			for i, sub := range subscriptionsByBlog {
				if sub.ID == *req.SubscriptionId {
					idTemp := subscriptionsByBlog[i].ID
					post.SubscriptionId = &idTemp
					subFound = true
					break
				}
			}
			if !subFound {
				loggingMap.SetMessage("subscription by id doesn't exists on the blog")
				ctx.JSON(http.StatusBadRequest, nil)
				return
			}
		}
	case "4":
		if req.Price == nil {
			loggingMap.SetMessage("price is required for access mode 4")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}
		post.Price = req.Price
		post.SubscriptionId = nil
	default:
		post.AccessMode = "1"
		post.Price = nil
		post.SubscriptionId = nil
	}

	post.Title = req.Title
	post.ShortDescription = req.ShortDescription
	post.Status = req.Status
	post.Url = req.Url
	post.TagsString = req.TagsString
	post.Updated = time.Now().UTC()

	err = h.service.repository.UpdatePost(ctx, post)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	err = h.service.SetTagsToPost(ctx, post.ID, req.TagsString)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set tags to post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) updateTitle(ctx *gin.Context) {
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

	var req PostUpdateTitleRequest
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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	post.Title = req.Title
	post.ShortDescription = req.ShortDescription
	post.Updated = time.Now().UTC()

	err = h.service.repository.UpdatePost(ctx, post)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) updateUrl(ctx *gin.Context) {
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

	var req PostUpdateUrlRequest
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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	postByUrl, err := h.service.repository.PostsByBlogIdAndUrl(ctx, blog.ID, req.Url)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by url")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if postByUrl != nil && postByUrl.ID != post.ID {
		loggingMap.SetMessage("post by url already exists")
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	post.Url = req.Url
	post.Updated = time.Now().UTC()

	err = h.service.repository.UpdatePost(ctx, post)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) getContent(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	userId := requestuser.GetUserID(ctx)
	if userId == nil {
		nilId := uuid.Nil
		userId = &nilId
	}

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	haveAccess, err := h.service.CheckUserContentAccess(ctx, post, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check user content access")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if !haveAccess {
		loggingMap.SetMessage("user doesn't have access to content")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	err = h.service.AddPostViewWithUser(ctx, post.ID, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to add post view")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	content, err := h.service.ContentById(ctx, post.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get content by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, content)
}

func (h *postHandler) updateContent(ctx *gin.Context) {
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

	var req PostUpdateContentRequest
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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	content, err := h.service.ContentById(ctx, post.ID)
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
		loggingMap.SetMessage("failed to update post content")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) uploadContentFile(ctx *gin.Context) {
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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	content, err := h.service.ContentById(ctx, post.ID)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get content by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	fileBytes, err := ctx.GetRawData()
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get cover from request body")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	contentType := http.DetectContentType(fileBytes)
	if contentType != "image/jpeg" && contentType != "image/png" {
		loggingMap.SetMessage("wrong cover file type")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	fileContent, err := h.service.CreateFileToContent(ctx, content, fileBytes, contentType)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create file to content")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusCreated, fileContent)
}

func (h *postHandler) getContentFile(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	fileIdParam := ctx.Param("file_id")
	if fileIdParam == "" {
		loggingMap.SetError("incorrect param file_id")
		loggingMap.SetMessage("failed to redirect url to content file")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	redirectUrl, err := h.service.RedirectUrlToFile(fileIdParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to redirect content file")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.None()
	ctx.Redirect(http.StatusSeeOther, redirectUrl)
}

func (h *postHandler) deleteContentFile(ctx *gin.Context) {
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
	fileIdParam := ctx.Param("file_id")
	fileId, err := uuid.Parse(fileIdParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param file_id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	contentFile, err := h.service.repository.ContentFileById(ctx, fileId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get content file by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if contentFile == nil {
		ctx.JSON(http.StatusNoContent, nil)
		return
	}

	err = h.service.DeleteContentFile(ctx, contentFile)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to delete content file")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

func (h *postHandler) updateTags(ctx *gin.Context) {
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

	var req PostUpdateTagsRequest
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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	post.TagsString = req.TagsString
	post.Updated = time.Now().UTC()

	err = h.service.repository.UpdatePost(ctx, post)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	err = h.service.SetTagsToPost(ctx, post.ID, req.TagsString)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set tags to post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) updateStatus(ctx *gin.Context) {
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

	var req PostUpdateStatusRequest
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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	post.Status = req.Status
	post.Updated = time.Now().UTC()

	err = h.service.repository.UpdatePost(ctx, post)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) updateAccessMode(ctx *gin.Context) {
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

	var req PostUpdateAccessModeRequest
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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	loggingMap["post_id"] = fmt.Sprintf("%s", post.ID.String())

	blog, err := h.service.BlogById(ctx, post.BlogId)
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
	loggingMap["blog_id"] = fmt.Sprintf("%s", blog.ID.String())
	if blog.AuthorId != *userId {
		loggingMap.SetMessage("request user is not the author of the blog")
		ctx.JSON(http.StatusForbidden, nil)
		return
	}

	post.AccessMode = req.AccessMode
	switch req.AccessMode {
	case "1":
		post.Price = nil
		post.SubscriptionId = nil
	case "2":
		post.Price = nil
		post.SubscriptionId = nil
	case "3":
		post.Price = nil
		if req.SubscriptionId == nil {
			post.SubscriptionId = nil
		} else {
			subscriptionsByBlog, err := h.service.repository.SubscriptionsByBlogId(ctx, blog.ID)
			if err != nil {
				loggingMap.SetError(err.Error())
				loggingMap.SetMessage("failed to get subscriptions by blog id")
				ctx.JSON(http.StatusInternalServerError, nil)
				return
			}
			subFound := false
			for i, sub := range subscriptionsByBlog {
				if sub.ID == *req.SubscriptionId {
					idTemp := subscriptionsByBlog[i].ID
					post.SubscriptionId = &idTemp
					subFound = true
					break
				}
			}
			if !subFound {
				loggingMap.SetMessage("subscription by id doesn't exists on the blog")
				ctx.JSON(http.StatusBadRequest, nil)
				return
			}
		}
	case "4":
		post.Price = req.Price
		post.SubscriptionId = nil
	default:
		post.AccessMode = "1"
		post.Price = nil
		post.SubscriptionId = nil
	}
	post.Updated = time.Now().UTC()

	err = h.service.repository.UpdatePost(ctx, post)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) updateCover(ctx *gin.Context) {
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
	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	blog, err := h.service.BlogById(ctx, post.BlogId)
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

	err = h.service.SetPostCover(ctx, post, coverBytes)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to set cover to post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) byUrl(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)
	var err error
	var blog *Blog
	blogUrlParam := ctx.Param("blog_url")
	blogIdParam := ctx.Param("blog_id")
	if blogIdParam != "" {
		blogId, err := uuid.Parse(blogIdParam)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("incorrect param blog_id")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}
		blog, err = h.service.BlogById(ctx, blogId)
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
	} else {
		blog, err = h.service.BlogByUrl(ctx, blogUrlParam)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get blog by url")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		if blog == nil {
			loggingMap.SetMessage("blog by url doesn't exists")
			ctx.JSON(http.StatusNotFound, nil)
			return
		}
	}
	postUrlParam := ctx.Param("post_url")
	post, err := h.service.repository.PostsByBlogIdAndUrl(ctx, blog.ID, postUrlParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by url")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by url doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	ctx.JSON(http.StatusOK, post)
}

func (h *postHandler) getCover(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	blogUrlParam := ctx.Param("blog_url")
	postUrlParam := ctx.Param("post_url")

	blog, err := h.service.BlogByUrl(ctx, blogUrlParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get blog by url")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if blog == nil {
		loggingMap.SetMessage("blog by url doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	post, err := h.service.repository.PostsByBlogIdAndUrl(ctx, blog.ID, postUrlParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by url")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by url doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	if post == nil {
		loggingMap.SetMessage("blog by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	if post.Cover == nil {
		loggingMap.Debug()
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	redirectUrl, err := h.service.RedirectUrlToFile(*post.Cover)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to redirect url to cover")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.None()
	ctx.Redirect(http.StatusSeeOther, redirectUrl)
}

func (h *postHandler) byFollowedBlogs(ctx *gin.Context) {
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
	posts, err := h.service.repository.PostsByBlogIdList(ctx, unique(blogIds))
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get posts")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, posts)

}

func (h *postHandler) getPostLikesInfo(ctx *gin.Context) {
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

	if userId == nil {
		nilId := uuid.Nil
		userId = &nilId
	}

	likesCount, err := h.service.PostLikesInfo(ctx, id, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post likes info")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, likesCount)
}

func (h *postHandler) like(ctx *gin.Context) {
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

	err = h.service.LikePost(ctx, id, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to like post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("user liked post")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) dislike(ctx *gin.Context) {
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

	err = h.service.DislikePost(ctx, id, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to dislike post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	loggingMap.SetMessage("user disliked post")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) unsetLike(ctx *gin.Context) {
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

	err = h.service.UnsetLike(ctx, id, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to unset like on a post")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	loggingMap.SetMessage("user unset like on post")
	ctx.JSON(http.StatusAccepted, nil)
}

func (h *postHandler) checkMyContentAccess(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	userId := requestuser.GetUserID(ctx)
	if userId == nil {
		nilId := uuid.Nil
		userId = &nilId
	}

	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	blog, err := h.service.BlogById(ctx, post.BlogId)
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
	if userId != nil {
		if blog.AuthorId == *userId {
			ctx.JSON(http.StatusOK, PostMyContentAccessResponse{
				HaveAccess:   true,
				UserId:       *userId,
				PostId:       post.ID,
				Price:        0,
				AccessMode:   post.AccessMode,
				Subscription: nil,
			})
			return
		}
	}

	blogSubscriptions, err := h.service.repository.SubscriptionsByBlogId(ctx, post.BlogId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get subscriptions by blog id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	switch post.AccessMode {
	case "1":
		ctx.JSON(http.StatusOK, PostMyContentAccessResponse{
			HaveAccess:   true,
			UserId:       *userId,
			PostId:       post.ID,
			Price:        0,
			AccessMode:   "1",
			Subscription: nil,
		})
		return

	case "2":
		haveFreeSubscription := false
		for _, sub := range blogSubscriptions {
			if sub.IsFree {
				haveFreeSubscription = true
				break
			}
		}

		if !haveFreeSubscription {
			ctx.JSON(http.StatusOK, PostMyContentAccessResponse{
				HaveAccess:   true,
				UserId:       *userId,
				PostId:       post.ID,
				Price:        0,
				AccessMode:   "2",
				Subscription: nil,
			})
			return
		}

		follow, err := h.service.repository.UserFollowByUserIdAndBlogId(ctx, *userId, post.BlogId)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get follow by user id and blog id")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}

		ctx.JSON(http.StatusOK, PostMyContentAccessResponse{
			HaveAccess:   follow != nil,
			UserId:       *userId,
			PostId:       post.ID,
			Price:        0,
			AccessMode:   "2",
			Subscription: nil,
		})
		return

	case "3":
		firstPaidSubscriptionIdx := 0
		havePaidSubscription := false
		for i, sub := range blogSubscriptions {
			if !sub.IsFree {
				firstPaidSubscriptionIdx = i
				havePaidSubscription = true
				break
			}
		}
		if !havePaidSubscription {
			ctx.JSON(http.StatusOK, PostMyContentAccessResponse{
				HaveAccess:   true,
				UserId:       *userId,
				PostId:       post.ID,
				Price:        0,
				AccessMode:   "3",
				Subscription: nil,
			})
			return
		}
		if post.SubscriptionId == nil {
			tempId := blogSubscriptions[firstPaidSubscriptionIdx].ID
			post.SubscriptionId = &tempId
			err = h.service.repository.UpdatePost(ctx, post)
			if err != nil {
				loggingMap.SetError(err.Error())
				loggingMap.SetMessage("failed to update post, changing nil subscription id to first paid subscription id")
				ctx.JSON(http.StatusInternalServerError, nil)
				return
			}
		}

		var reqSubId = 0
		for i, sub := range blogSubscriptions {
			if sub.ID == *post.SubscriptionId {
				reqSubId = i
				break
			}
		}
		var requiredSubscriptions = blogSubscriptions[reqSubId:]

		userSubs, err := h.service.repository.UserSubscriptionsByUserIdAndBlogId(ctx, *userId, post.BlogId)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get subscriptions by user id and blog id")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		haveRequiredSubscription := false

		for _, sub := range requiredSubscriptions {
			for _, userSub := range userSubs {
				if sub.ID == userSub.SubscriptionId && userSub.IsActive {
					haveRequiredSubscription = true
					ctx.JSON(http.StatusOK, PostMyContentAccessResponse{
						HaveAccess:   haveRequiredSubscription,
						UserId:       *userId,
						PostId:       post.ID,
						Price:        0,
						AccessMode:   "3",
						Subscription: &blogSubscriptions[reqSubId],
					})
					return
				}
			}
		}

		ctx.JSON(http.StatusOK, PostMyContentAccessResponse{
			HaveAccess:   haveRequiredSubscription,
			UserId:       *userId,
			PostId:       post.ID,
			Price:        0,
			AccessMode:   "3",
			Subscription: &blogSubscriptions[reqSubId],
		})
		return

	case "4":
		userPaidAccess, err := h.service.repository.PostPaidAccessByPostIdAndUserId(ctx, id, *userId)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to get paid access by post id and user id")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		ctx.JSON(http.StatusOK, PostMyContentAccessResponse{
			HaveAccess:   userPaidAccess != nil,
			UserId:       *userId,
			PostId:       post.ID,
			Price:        *post.Price,
			AccessMode:   "4",
			Subscription: nil,
		})
		return

	default:
		loggingMap.SetMessage("unknown access mode: " + post.AccessMode)
		ctx.JSON(http.StatusInternalServerError, nil)
	}
}

func (h *postHandler) buyPaidAccessRobokassa(ctx *gin.Context) {

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

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}
	if post.AccessMode != "4" {
		loggingMap.SetMessage("access mode is not paid")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	blog, err := h.service.repository.BlogById(ctx, post.BlogId)
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
		loggingMap.SetMessage("author can't buy paid access on his own blog")
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	userPaidAccess, err := h.service.repository.PostPaidAccessByPostIdAndUserId(ctx, id, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get paid access by post id and user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if userPaidAccess != nil {
		loggingMap.SetMessage("user already have paid access")
		ctx.JSON(http.StatusConflict, nil)
		return
	}

	link, err := h.service.GetPostRobokassaPaymentLink(ctx, post, *userId)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get payment link")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"payment_link": link})
}

func (h *postHandler) byCategories(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)

	codeListQuery := ctx.Query("list")
	codeList := strings.Split(codeListQuery, ",")

	posts, err := h.service.repository.PostsByCategories(ctx, codeList)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get posts by categories")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	ctx.JSON(http.StatusOK, posts)
}

func (h *postHandler) addAnonView(ctx *gin.Context) {

	loggingMap := serverlogging.GetLoggingMap(ctx)

	fingerprint := ctx.Query("fingerprint")
	if fingerprint == "" {
		loggingMap.SetMessage("fingerprint is empty")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	idParam := ctx.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("incorrect param id")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	post, err := h.service.PostById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get post by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if post == nil {
		loggingMap.SetMessage("post by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	err = h.service.AddPostViewWithFingerprint(ctx, post.ID, fingerprint)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to add post view with fingerprint")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.JSON(http.StatusOK, nil)
}
