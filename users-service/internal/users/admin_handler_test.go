package users

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"users-service/pkg/filelogger"
	"users-service/pkg/pgutils"
	"users-service/pkg/queuelogger"
	serverlogging "users-service/pkg/serverlogging/gin"
	"users-service/pkg/testhelpers"
)

type AdminHandlerTestSuite struct {
	suite.Suite
	ctx         context.Context
	pgContainer *testhelpers.PostgresContainer
	dbConn      *pgxpool.Pool
	repository  *Repository
	service     *Service
	router      *gin.Engine
}

func (suite *AdminHandlerTestSuite) performRequest(req *http.Request) *http.Response {
	req.Header.Set("USER-ROLE", "admin")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w.Result()
}

func (suite *AdminHandlerTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	pgContainer, err := testhelpers.CreatePostgresContainer(suite.ctx)
	if err != nil {
		log.Fatal(err)
	}
	suite.pgContainer = pgContainer

	dbConn, err := pgutils.DBPool(suite.ctx, suite.pgContainer.ConnectionString)
	if err != nil {
		log.Fatal(err)
	}
	suite.dbConn = dbConn

	err = pgutils.UpMigrations(suite.pgContainer.ConnectionString, "../../migrations")
	if err != nil {
		log.Fatal(err)
	}

	fileLogger := filelogger.NewFileLogger("../../app.log")
	fileLogger.EnableConsoleLog()
	queueLogger := queuelogger.Mock{}

	suite.repository = NewRepository(dbConn)
	suite.service = NewService(suite.ctx, suite.repository)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	suite.router = router
	router.Use(gin.CustomRecovery(serverlogging.NewPanicLogger(fileLogger, queueLogger)))
	router.Use(serverlogging.NewRequestLogger(fileLogger, queueLogger))
	RegisterAdminHandler(router.Group("/test"), suite.service)
}

func (suite *AdminHandlerTestSuite) TearDownSuite() {
	suite.dbConn.Close()
	_ = suite.pgContainer.Terminate(suite.ctx)
}

func TestAdminHandlerSuite(t *testing.T) {
	suite.Run(t, new(AdminHandlerTestSuite))
}

func (suite *AdminHandlerTestSuite) Test401() {
	t := suite.T()
	req, _ := http.NewRequest("GET", "/test/users", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	resp := w.Result()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func (suite *AdminHandlerTestSuite) TestHealthcheck() {
	t := suite.T()

	req, _ := http.NewRequest("GET", "/test/healthcheck", nil)
	resp := suite.performRequest(req)

	var respBody HealthCheckResponse
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, HealthCheckResponse{Status: "ok", Up: true}, respBody)
}

func (suite *AdminHandlerTestSuite) TestAll() {
	t := suite.T()

	req, _ := http.NewRequest("GET", "/test/users", nil)
	resp := suite.performRequest(req)

	var respBody []AdminGetUserRequest
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, len(respBody), 1)
}

func (suite *AdminHandlerTestSuite) TestGetById() {
	t := suite.T()

	testUser, err := suite.service.Create(suite.ctx,
		"TestGetById",
		"TestGetById@TestGetById.com",
		"TestGetById",
		"admin",
		true,
		nil)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/test/users/"+testUser.ID.String(), nil)
	resp := suite.performRequest(req)

	var respBody AdminGetUserRequest
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testUser.ID, respBody.ID)
	assert.Equal(t, testUser.Email, respBody.Email)
	assert.Equal(t, testUser.Role, respBody.Role)
	assert.Equal(t, testUser.Deleted, respBody.Deleted)
	assert.Equal(t, testUser.Enabled, respBody.Enabled)
	assert.Equal(t, testUser.Login, respBody.Login)
}

func (suite *AdminHandlerTestSuite) TestCreate() {
	t := suite.T()

	reqBody := AdminCreateUserRequest{
		Login:    "Test_Create123",
		Email:    "TestCreate@TestCreate.com",
		Role:     "moderator",
		Password: "132123",
		Enabled:  true,
	}

	req := httptest.NewRequest("POST", "/test/users", testhelpers.ReqBody(reqBody))
	resp := suite.performRequest(req)

	var respBody AdminGetUserRequest
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, reqBody.Email, respBody.Email)
	assert.Equal(t, reqBody.Role, respBody.Role)
	assert.Equal(t, reqBody.Enabled, respBody.Enabled)
	assert.Equal(t, reqBody.Login, respBody.Login)

	req, _ = http.NewRequest("GET", "/test/users/"+respBody.ID.String(), nil)
	resp = suite.performRequest(req)

	var respBody2 AdminGetUserRequest
	err = json.NewDecoder(resp.Body).Decode(&respBody2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, respBody.ID, respBody2.ID)
	assert.Equal(t, respBody.Email, respBody2.Email)
	assert.Equal(t, respBody.Role, respBody2.Role)
	assert.Equal(t, respBody.Deleted, respBody2.Deleted)
	assert.Equal(t, respBody.Enabled, respBody2.Enabled)
	assert.Equal(t, respBody.Login, respBody2.Login)
}

func (suite *AdminHandlerTestSuite) TestCreate400() {
	t := suite.T()

	_, err := suite.service.Create(suite.ctx,
		"TestCreate400",
		"TestCreate400@TestCreate400.com",
		"TestCreate400",
		"admin",
		true,
		nil)
	assert.NoError(t, err)

	reqBody := AdminCreateUserRequest{
		Login:    "testcreate400",
		Email:    "TestCreate@TestCreate.com",
		Role:     "moderator",
		Password: "132123",
		Enabled:  true,
	}

	req := httptest.NewRequest("POST", "/test/users", testhelpers.ReqBody(reqBody))
	resp := suite.performRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	reqBody = AdminCreateUserRequest{
		Login:    "TestCreate4001",
		Email:    "TestCreate400@TestCreate400.com",
		Role:     "moderator",
		Password: "132123",
		Enabled:  true,
	}

	req = httptest.NewRequest("POST", "/test/users", testhelpers.ReqBody(reqBody))
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	reqBody = AdminCreateUserRequest{
		Login:    "???",
		Email:    "TestCreate412300@TestCreate400.com",
		Role:     "moderator",
		Password: "132123",
		Enabled:  true,
	}

	req = httptest.NewRequest("POST", "/test/users", testhelpers.ReqBody(reqBody))
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	reqBody = AdminCreateUserRequest{
		Login:    "login123",
		Email:    "login123@login123.com",
		Role:     "role123",
		Password: "132123",
		Enabled:  true,
	}

	req = httptest.NewRequest("POST", "/test/users", testhelpers.ReqBody(reqBody))
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	reqBody = AdminCreateUserRequest{
		Login:    "long_long_long_long_long_long_long_",
		Email:    "long_long_long_long_long_long_long_@login123.com",
		Role:     "user",
		Password: "132123",
		Enabled:  true,
	}

	req = httptest.NewRequest("POST", "/test/users", testhelpers.ReqBody(reqBody))
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	req = httptest.NewRequest("POST", "/test/users", nil)
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

}

func (suite *AdminHandlerTestSuite) TestUpdate() {
	t := suite.T()

	testUser, err := suite.service.Create(suite.ctx,
		"TestUpdate",
		"TestUpdate@TestUpdate.com",
		"TestUpdate",
		"user",
		true,
		nil)
	assert.NoError(t, err)

	reqBody := AdminUpdateUserRequest{
		Login:        "TestUpdate1",
		Email:        "TestUpdate1@TestUpdate1.com",
		Role:         "moderator",
		Deleted:      true,
		Enabled:      false,
		BannedUntil:  testhelpers.TimePtr(time.Now().UTC().Add(time.Hour)),
		BannedReason: testhelpers.StrPtr("TestUpdate"),
	}

	req := httptest.NewRequest("PUT", "/test/users/"+testUser.ID.String(), testhelpers.ReqBody(reqBody))
	resp := suite.performRequest(req)

	var respBody AdminGetUserRequest
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, testUser.ID, respBody.ID)
	assert.Equal(t, reqBody.Email, respBody.Email)
	assert.Equal(t, reqBody.Role, respBody.Role)
	assert.Equal(t, reqBody.Deleted, respBody.Deleted)
	assert.Equal(t, reqBody.Enabled, respBody.Enabled)
	assert.Equal(t, reqBody.Login, respBody.Login)
	assert.Equal(t, reqBody.BannedReason, respBody.BannedReason)

	req, _ = http.NewRequest("GET", "/test/users/"+testUser.ID.String(), nil)
	resp = suite.performRequest(req)
	respBody = AdminGetUserRequest{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testUser.ID, respBody.ID)
	assert.Equal(t, reqBody.Email, respBody.Email)
	assert.Equal(t, reqBody.Role, respBody.Role)
	assert.Equal(t, reqBody.Deleted, respBody.Deleted)
	assert.Equal(t, reqBody.Enabled, respBody.Enabled)
	assert.Equal(t, reqBody.Login, respBody.Login)
	assert.Equal(t, reqBody.BannedReason, respBody.BannedReason)

	reqBody = AdminUpdateUserRequest{
		Login:        "TestUpdate1",
		Email:        "TestUpdate@TestUpdate1.com",
		Role:         "moderator",
		Deleted:      true,
		Enabled:      false,
		BannedUntil:  testhelpers.TimePtr(time.Now().UTC().Add(time.Hour)),
		BannedReason: testhelpers.StrPtr("TestUpdate"),
	}
	req = httptest.NewRequest("PUT", "/test/users/"+testUser.ID.String(), testhelpers.ReqBody(reqBody))
	resp = suite.performRequest(req)
	respBody = AdminGetUserRequest{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, testUser.ID, respBody.ID)
	assert.Equal(t, reqBody.Email, respBody.Email)
	assert.Equal(t, reqBody.Role, respBody.Role)
	assert.Equal(t, reqBody.Deleted, respBody.Deleted)
	assert.Equal(t, reqBody.Enabled, respBody.Enabled)
	assert.Equal(t, reqBody.Login, respBody.Login)
	assert.Equal(t, reqBody.BannedReason, respBody.BannedReason)

	reqBody = AdminUpdateUserRequest{
		Login:        "TestUpdate",
		Email:        "TestUpdate@TestUpdate1.com",
		Role:         "moderator",
		Deleted:      true,
		Enabled:      false,
		BannedUntil:  testhelpers.TimePtr(time.Now().UTC().Add(time.Hour)),
		BannedReason: testhelpers.StrPtr("TestUpdate"),
	}
	req = httptest.NewRequest("PUT", "/test/users/"+testUser.ID.String(), testhelpers.ReqBody(reqBody))
	resp = suite.performRequest(req)
	respBody = AdminGetUserRequest{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, testUser.ID, respBody.ID)
	assert.Equal(t, reqBody.Email, respBody.Email)
	assert.Equal(t, reqBody.Role, respBody.Role)
	assert.Equal(t, reqBody.Deleted, respBody.Deleted)
	assert.Equal(t, reqBody.Enabled, respBody.Enabled)
	assert.Equal(t, reqBody.Login, respBody.Login)
	assert.Equal(t, reqBody.BannedReason, respBody.BannedReason)

	reqBody = AdminUpdateUserRequest{
		Login:        "TestUpdate",
		Email:        "TestUpdate@TestUpdate1.com",
		Role:         "admin",
		Deleted:      false,
		Enabled:      true,
		BannedUntil:  nil,
		BannedReason: nil,
	}
	req = httptest.NewRequest("PUT", "/test/users/"+testUser.ID.String(), testhelpers.ReqBody(reqBody))
	resp = suite.performRequest(req)
	respBody = AdminGetUserRequest{}
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.Equal(t, testUser.ID, respBody.ID)
	assert.Equal(t, reqBody.Email, respBody.Email)
	assert.Equal(t, reqBody.Role, respBody.Role)
	assert.Equal(t, reqBody.Deleted, respBody.Deleted)
	assert.Equal(t, reqBody.Enabled, respBody.Enabled)
	assert.Equal(t, reqBody.Login, respBody.Login)
	assert.Equal(t, reqBody.BannedReason, respBody.BannedReason)
}

func (suite *AdminHandlerTestSuite) TestErase() {
	t := suite.T()

	testUser, err := suite.service.Create(suite.ctx,
		"TestErase",
		"TestErase@TestErase.com",
		"TestErase",
		"user",
		true,
		nil)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/test/users/"+testUser.ID.String(), nil)
	resp := suite.performRequest(req)

	var respBody AdminGetUserRequest
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req, _ = http.NewRequest("DELETE", "/test/users/"+testUser.ID.String(), nil)
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	req, _ = http.NewRequest("GET", "/test/users/"+testUser.ID.String(), nil)
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
