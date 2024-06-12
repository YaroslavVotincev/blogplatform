package configservice

import (
	"config-service/pkg/filelogger"
	"config-service/pkg/pgutils"
	"config-service/pkg/queuelogger"
	serverlogging "config-service/pkg/serverlogging/gin"
	"config-service/pkg/testhelpers"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

	suite.repository = NewRepository(dbConn)
	suite.service = NewService(suite.repository)

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	suite.router = router

	fileLogger := filelogger.NewFileLogger("../../app.log")
	fileLogger.EnableConsoleLog()
	queueLogger := queuelogger.Mock{}

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
	req, _ := http.NewRequest("GET", "/test/services", nil)
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

func (suite *AdminHandlerTestSuite) TestGetAllServices() {
	t := suite.T()

	_, err := suite.service.NewService(suite.ctx, "TestGetAllServices")
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/test/services", nil)
	resp := suite.performRequest(req)

	var respBody []ServiceModel
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.GreaterOrEqual(t, len(respBody), 1)
}

func (suite *AdminHandlerTestSuite) TestCreateService() {
	t := suite.T()

	req, _ := http.NewRequest("POST", "/test/services",
		testhelpers.ReqBody(CreateServiceRequest{"TestCreateService"}))
	resp := suite.performRequest(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody ServiceModel
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	req, _ = http.NewRequest("GET", "/test/services/TestCreateService", nil)
	resp = suite.performRequest(req)

	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "TestCreateService", respBody.Service)
}

func (suite *AdminHandlerTestSuite) TestCreateServiceWithSameNameError() {
	t := suite.T()

	req, _ := http.NewRequest("POST", "/test/services",
		testhelpers.ReqBody(CreateServiceRequest{"TestCreateServiceWithSameNameError"}))
	resp := suite.performRequest(req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var respBody ServiceModel
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	req, _ = http.NewRequest("POST", "/test/services",
		testhelpers.ReqBody(CreateServiceRequest{"TestCreateServiceWithSameNameError"}))
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func (suite *AdminHandlerTestSuite) TestGetServiceByName() {
	t := suite.T()

	testService, err := suite.service.NewService(suite.ctx, "TestGetServiceByName")
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/test/services/TestGetServiceByName", nil)
	resp := suite.performRequest(req)

	var respBody ServiceModel
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, respBody.Service, testService.Service)
}

func (suite *AdminHandlerTestSuite) TestUpdateService() {
	t := suite.T()

	testService, err := suite.service.NewService(suite.ctx, "TestUpdateService")
	assert.NoError(t, err)

	req, _ := http.NewRequest("PUT", "/test/services/TestUpdateService", testhelpers.ReqBody(
		UpdateServiceRequest{"TestUpdateService1"}))
	resp := suite.performRequest(req)

	var respBody ServiceModel
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	assert.NotEqual(t, respBody.Service, testService.Service)
	assert.Equal(t, respBody.Service, "TestUpdateService1")

	req, _ = http.NewRequest("GET", "/test/services/TestUpdateService", nil)
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	req, _ = http.NewRequest("GET", "/test/services/TestUpdateService1", nil)
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, respBody.Service, "TestUpdateService1")
}

func (suite *AdminHandlerTestSuite) TestUpdateServiceWithSameNameError() {
	t := suite.T()

	_, err := suite.service.NewService(suite.ctx, "TestUpdateServiceWithSameNameError1")
	assert.NoError(t, err)

	_, err = suite.service.NewService(suite.ctx, "TestUpdateServiceWithSameNameError2")
	assert.NoError(t, err)

	req, _ := http.NewRequest("PUT", "/test/services/TestUpdateServiceWithSameNameError1", testhelpers.ReqBody(
		UpdateServiceRequest{"TestUpdateServiceWithSameNameError2"}))
	resp := suite.performRequest(req)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func (suite *AdminHandlerTestSuite) TestDeleteService() {
	t := suite.T()

	_, err := suite.service.NewService(suite.ctx, "TestDeleteService")
	assert.NoError(t, err)

	req, _ := http.NewRequest("DELETE", "/test/services/TestDeleteService", nil)
	resp := suite.performRequest(req)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	req, _ = http.NewRequest("GET", "/test/services/TestDeleteService", nil)
	resp = suite.performRequest(req)
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func (suite *AdminHandlerTestSuite) TestGetSettings() {
	t := suite.T()

	testService, err := suite.service.NewService(suite.ctx, "TestGetSettings")
	assert.NoError(t, err)

	testSettings := []CreateSettingRequest{
		{
			Key:   "key1",
			Value: "value1",
		},
		{
			Key:   "key2",
			Value: "value2",
		},
	}

	err = suite.service.SetSettingsToService(suite.ctx, testService.Service, testSettings)
	assert.NoError(t, err)

	req, _ := http.NewRequest("GET", "/test/services/TestGetSettings/settings", nil)
	resp := suite.performRequest(req)

	var respSettings []CreateSettingRequest
	err = json.NewDecoder(resp.Body).Decode(&respSettings)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testSettings, respSettings)
}

func (suite *AdminHandlerTestSuite) TestSetSettings() {
	t := suite.T()

	testService, err := suite.service.NewService(suite.ctx, "TestSetSettings")
	assert.NoError(t, err)

	testInitialSettings := []CreateSettingRequest{
		{
			Key:   "key1",
			Value: "value1",
		},
		{
			Key:   "key2",
			Value: "value2",
		},
	}

	err = suite.service.SetSettingsToService(suite.ctx, testService.Service, testInitialSettings)
	assert.NoError(t, err)

	testSetSettings := []CreateSettingRequest{
		{
			Key:   "key1",
			Value: "value3",
		},
		{
			Key:   "key3",
			Value: "value4",
		},
		{
			Key:   "key4",
			Value: "value5",
		},
	}

	req, _ := http.NewRequest("POST", "/test/services/TestSetSettings/settings", testhelpers.ReqBody(testSetSettings))
	resp := suite.performRequest(req)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)

	req, _ = http.NewRequest("GET", "/test/services/TestSetSettings/settings", nil)
	resp = suite.performRequest(req)

	var respSettings []CreateSettingRequest
	err = json.NewDecoder(resp.Body).Decode(&respSettings)
	assert.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testSetSettings, respSettings)
}
