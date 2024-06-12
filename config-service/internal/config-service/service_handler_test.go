package configservice

import (
	"config-service/pkg/filelogger"
	"config-service/pkg/pgutils"
	"config-service/pkg/queuelogger"
	serverlogging "config-service/pkg/serverlogging/gin"
	"config-service/pkg/testhelpers"
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
)

type ServiceHandlerTestSuit struct {
	suite.Suite
	ctx         context.Context
	pgContainer *testhelpers.PostgresContainer
	dbConn      *pgxpool.Pool
	repository  *Repository
	service     *Service
	router      *gin.Engine
}

func (suite *ServiceHandlerTestSuit) performRequest(req *http.Request) *http.Response {
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w.Result()
}

func (suite *ServiceHandlerTestSuit) SetupSuite() {
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
	RegisterServiceHandler(router.Group("/test"), suite.service)
}

func (suite *ServiceHandlerTestSuit) TearDownSuite() {
	suite.dbConn.Close()
	_ = suite.pgContainer.Terminate(suite.ctx)
}

func TestServiceHandlerSuite(t *testing.T) {
	suite.Run(t, new(ServiceHandlerTestSuit))
}

func (suite *ServiceHandlerTestSuit) TestGetSettings() {
	t := suite.T()

	testService, err := suite.service.NewService(suite.ctx, "test")
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

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("SERVICE_NAME", "test")
	resp := suite.performRequest(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respSettings []CreateSettingRequest
	err = json.NewDecoder(resp.Body).Decode(&respSettings)
	assert.NoError(t, err)

	assert.Equal(t, testSettings, respSettings)

}

func (suite *ServiceHandlerTestSuit) TestGet404() {
	t := suite.T()

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("SERVICE_NAME", "404")
	resp := suite.performRequest(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func (suite *ServiceHandlerTestSuit) TestGet401() {
	t := suite.T()

	req, _ := http.NewRequest("GET", "/test", nil)
	resp := suite.performRequest(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
