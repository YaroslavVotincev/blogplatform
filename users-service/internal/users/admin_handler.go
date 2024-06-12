package users

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"net/http"
	serverlogging "users-service/pkg/serverlogging/gin"
)

type adminHandler struct {
	service  *Service
	validate *validator.Validate
}

func RegisterAdminHandler(api *gin.RouterGroup, service *Service) {
	h := &adminHandler{service: service, validate: NewUsersValidator()}

	api.Use(AdminMiddleware())

	api.GET("/healthcheck", h.healthCheck)
	api.GET("/users", h.all)
	api.GET("/users/:id", h.byId)
	api.POST("/users", h.create)
	api.PUT("/users/:id", h.update)
	api.DELETE("/users/:id", h.erase)
}

func (h *adminHandler) healthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, HealthCheckResponse{Status: "ok", Up: true})
}

func (h *adminHandler) all(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	allUsers, err := h.service.All(ctx)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get all users")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	responseArray := make([]AdminGetUserRequest, len(allUsers))
	for i := range allUsers {
		responseArray[i] = AdminGetUserRequest{
			ID:               allUsers[i].ID,
			Login:            allUsers[i].Login,
			Email:            allUsers[i].Email,
			Role:             allUsers[i].Role,
			Deleted:          allUsers[i].Deleted,
			Enabled:          allUsers[i].Enabled,
			EmailConfirmedAt: allUsers[i].EmailConfirmedAt,
			EraseAt:          allUsers[i].EraseAt,
			BannedUntil:      allUsers[i].BannedUntil,
			BannedReason:     allUsers[i].BannedReason,
			Created:          allUsers[i].Created,
			Updated:          allUsers[i].Updated,
		}
	}
	ctx.JSON(http.StatusOK, responseArray)
}

func (h *adminHandler) byId(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	loggingMap["req_id_param"] = idParam
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to parse user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	user, err := h.service.ByID(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if user == nil {
		loggingMap.SetMessage("user by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	ctx.JSON(http.StatusOK, AdminGetUserRequest{
		ID:               user.ID,
		Login:            user.Login,
		Email:            user.Email,
		Role:             user.Role,
		Deleted:          user.Deleted,
		Enabled:          user.Enabled,
		EmailConfirmedAt: user.EmailConfirmedAt,
		EraseAt:          user.EraseAt,
		BannedUntil:      user.BannedUntil,
		BannedReason:     user.BannedReason,
		Created:          user.Created,
		Updated:          user.Updated,
	})
}

func (h *adminHandler) create(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	var req AdminCreateUserRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad create user admin request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["req_body"] = fmt.Sprintf("%+v", req)
	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad create user admin request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	userByLogin, err := h.service.ByLogin(ctx, req.Login)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check if user by request's login already exists")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if userByLogin != nil {
		loggingMap["existing_user"] = fmt.Sprintf("%+v", *userByLogin)
		loggingMap.SetMessage("user by request's login already exists")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	userByEmail, err := h.service.ByEmail(ctx, req.Email)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to check if user by request's email already exists")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if userByEmail != nil {
		loggingMap["existing_user"] = fmt.Sprintf("%+v", *userByEmail)
		loggingMap.SetMessage("user by request's email already exists")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	user, err := h.service.Create(ctx, req.Login, req.Email, req.Password, req.Role, req.Enabled, nil)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to create new user")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if user == nil {
		loggingMap.SetMessage("failed to create new user, service-created user is nil without error for some reason")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("admin created new user")
	user.HashedPassword = ""
	loggingMap["new_user"] = fmt.Sprintf("%+v", *user)

	ctx.JSON(http.StatusCreated, AdminGetUserRequest{
		ID:               user.ID,
		Login:            user.Login,
		Email:            user.Email,
		Role:             user.Role,
		Deleted:          user.Deleted,
		Enabled:          user.Enabled,
		EmailConfirmedAt: user.EmailConfirmedAt,
		EraseAt:          user.EraseAt,
		BannedUntil:      user.BannedUntil,
		BannedReason:     user.BannedReason,
		Created:          user.Created,
		Updated:          user.Updated,
	})
}

func (h *adminHandler) update(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	var req AdminUpdateUserRequest
	if err := json.NewDecoder(ctx.Request.Body).Decode(&req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad update user admin request, failed to unmarshal to struct")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}
	loggingMap["req_body"] = fmt.Sprintf("%+v", req)
	if err := h.validate.Struct(req); err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("bad update user admin request, failed to validate data")
		ctx.JSON(http.StatusBadRequest, nil)
		return
	}

	idParam := ctx.Param("id")
	loggingMap["req_id_param"] = idParam
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to parse user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	user, err := h.service.ByID(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if user == nil {
		loggingMap.SetMessage("user by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	loggingMap["user_old"] = fmt.Sprintf("%+v", *user)

	if user.Login != req.Login {
		userByLogin, err := h.service.ByLogin(ctx, req.Login)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to check if user by request's login already exists")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		if userByLogin != nil {
			loggingMap["existing_user"] = fmt.Sprintf("%+v", *userByLogin)
			loggingMap.SetMessage("user by request's login already exists")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}
	}

	if user.Email != req.Email {
		userByEmail, err := h.service.ByEmail(ctx, req.Email)
		if err != nil {
			loggingMap.SetError(err.Error())
			loggingMap.SetMessage("failed to check if user by request's email already exists")
			ctx.JSON(http.StatusInternalServerError, nil)
			return
		}
		if userByEmail != nil {
			loggingMap["existing_user"] = fmt.Sprintf("%+v", *userByEmail)
			loggingMap.SetMessage("user by request's email already exists")
			ctx.JSON(http.StatusBadRequest, nil)
			return
		}
	}

	err = h.service.Update(ctx, user, req.Login, req.Email, req.Role, req.Deleted, req.Enabled, req.BannedUntil, req.BannedReason)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to update user")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("admin updated user")
	loggingMap["user_new"] = fmt.Sprintf("%+v", *user)

	ctx.JSON(http.StatusAccepted, AdminGetUserRequest{
		ID:               user.ID,
		Login:            user.Login,
		Email:            user.Email,
		Role:             user.Role,
		Deleted:          user.Deleted,
		Enabled:          user.Enabled,
		EmailConfirmedAt: user.EmailConfirmedAt,
		EraseAt:          user.EraseAt,
		BannedUntil:      user.BannedUntil,
		BannedReason:     user.BannedReason,
		Created:          user.Created,
		Updated:          user.Updated,
	})
}

func (h *adminHandler) erase(ctx *gin.Context) {
	loggingMap := serverlogging.GetLoggingMap(ctx)

	idParam := ctx.Param("id")
	loggingMap["req_id_param"] = idParam
	id, err := uuid.Parse(idParam)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to parse user id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	user, err := h.service.ByID(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to get user by id")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}
	if user == nil {
		loggingMap.SetMessage("user by id doesn't exists")
		ctx.JSON(http.StatusNotFound, nil)
		return
	}

	loggingMap["deleted_user"] = fmt.Sprintf("%+v", *user)

	err = h.service.EraseById(ctx, id)
	if err != nil {
		loggingMap.SetError(err.Error())
		loggingMap.SetMessage("failed to erase user")
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	loggingMap.SetMessage("admin erased user")

	ctx.JSON(http.StatusNoContent, nil)
}
