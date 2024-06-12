package historylogs

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"log"
	"logs-service/pkg/filelogger"
	"logs-service/pkg/pgutils"
	"logs-service/pkg/queuelogger"
	serverlogging "logs-service/pkg/serverlogging/gin"
	"logs-service/pkg/testhelpers"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	req, _ := http.NewRequest("GET", "/test/logs", nil)
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

func (suite *AdminHandlerTestSuite) TestLogsByLevel() {
	t := suite.T()
	testLogs := []HistoryLog{
		{
			Level:   "TestLogsByLevel_level1",
			Service: "TestLogsByLevel_service1",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByLevel_level2",
			Service: "TestLogsByLevel_service2",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByLevel_level1",
			Service: "TestLogsByLevel_service2",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByLevel_level2",
			Service: "TestLogsByLevel_service3",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByLevel_level1",
			Service: "TestLogsByLevel_service3",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByLevel_level3",
			Service: "TestLogsByLevel_service4",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key2": "value2",
			},
		},
	}

	for i := range testLogs {
		createdLog, err := suite.service.Create(suite.ctx, testLogs[i].Level, testLogs[i].Service, testLogs[i].UserID, testLogs[i].Data)
		assert.NoError(t, err)
		testLogs[i] = *createdLog
	}

	var respBody []HistoryLog

	req, _ := http.NewRequest("GET", "/test/logs?level=TestLogsByLevel_level1,TestLogsByLevel_level2,TestLogsByLevel_level3", nil)
	resp := suite.performRequest(req)
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 6, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsByLevel_level1,TestLogsByLevel_level2", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 5, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsByLevel_level1,TestLogsByLevel_level3", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 4, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsByLevel_level3,TestLogsByLevel_level2", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsByLevel_level1", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsByLevel_level2,TestLogsByLevel_level2", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsByLevel_level3", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, len(respBody))

	assert.Equal(t, testLogs[5].Level, respBody[0].Level)
	assert.Equal(t, testLogs[5].Service, respBody[0].Service)
	assert.Equal(t, testLogs[5].UserID, respBody[0].UserID)
	assert.Equal(t, testLogs[5].Data, respBody[0].Data)

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsByLevel_level4", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 0, len(respBody))
}

func (suite *AdminHandlerTestSuite) TestLogsByService() {
	t := suite.T()
	testLogs := []HistoryLog{
		{
			Level:   "TestLogsByService_level1",
			Service: "TestLogsByService_service1",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByService_level2",
			Service: "TestLogsByService_service2",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByService_level2",
			Service: "TestLogsByService_service1",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByService_level3",
			Service: "TestLogsByService_service2",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByService_level3",
			Service: "TestLogsByService_service1",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByService_level4",
			Service: "TestLogsByService_service3",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key2": "value2",
			},
		},
	}

	for i := range testLogs {
		createdLog, err := suite.service.Create(suite.ctx, testLogs[i].Level, testLogs[i].Service, testLogs[i].UserID, testLogs[i].Data)
		assert.NoError(t, err)
		testLogs[i] = *createdLog
	}

	var respBody []HistoryLog

	req, _ := http.NewRequest("GET", "/test/logs?service=TestLogsByService_service1,TestLogsByService_service2,TestLogsByService_service3", nil)
	resp := suite.performRequest(req)
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 6, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?service=TestLogsByService_service1,TestLogsByService_service2", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 5, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?service=TestLogsByService_service1,TestLogsByService_service3", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 4, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?service=TestLogsByService_service3,TestLogsByService_service2", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?service=TestLogsByService_service1", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?service=TestLogsByService_service2,TestLogsByService_service2", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?service=TestLogsByService_service3", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, len(respBody))

	assert.Equal(t, testLogs[5].Level, respBody[0].Level)
	assert.Equal(t, testLogs[5].Service, respBody[0].Service)
	assert.Equal(t, testLogs[5].UserID, respBody[0].UserID)
	assert.Equal(t, testLogs[5].Data, respBody[0].Data)

	req, _ = http.NewRequest("GET", "/test/logs?service=TestLogsByService_service4", nil)
	resp = suite.performRequest(req)
	respBody = make([]HistoryLog, 0)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 0, len(respBody))
}

func (suite *AdminHandlerTestSuite) TestLogsByUserId() {
	t := suite.T()

	id1 := uuid.New()
	id2 := uuid.New()
	testLogs := []HistoryLog{
		{
			Level:   "TestLogsByUserId",
			Service: "TestLogsByUserId",
			UserID:  &id1,
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByUserId",
			Service: "TestLogsByUserId",
			UserID:  &id2,
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByUserId",
			Service: "TestLogsByUserId",
			UserID:  &id2,
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByUserId",
			Service: "TestLogsByUserId",
			UserID:  &id2,
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByUserId",
			Service: "TestLogsByUserId",
			UserID:  nil,
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsByUserId",
			Service: "TestLogsByUserId",
			UserID:  nil,
			Data: map[string]any{
				"key2": "value2",
			},
		},
	}

	for i := range testLogs {
		createdLog, err := suite.service.Create(suite.ctx, testLogs[i].Level, testLogs[i].Service, testLogs[i].UserID, testLogs[i].Data)
		assert.NoError(t, err)
		testLogs[i] = *createdLog
	}

	var respBody []HistoryLog

	req, _ := http.NewRequest("GET", "/test/logs?user=none,"+id1.String()+","+id2.String(), nil)
	resp := suite.performRequest(req)
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 6, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?user="+id2.String()+",none", nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 5, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?user="+id2.String()+","+id1.String(), nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 4, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?user="+id2.String(), nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 3, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?user=none", nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 2, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?user="+id1.String(), nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?user="+uuid.NewString(), nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 0, len(respBody))

}

func (suite *AdminHandlerTestSuite) TestLogsByTime() {
	t := suite.T()

	timeMinus12 := time.Now().UTC().Add(-12 * time.Hour)
	timeMinus11 := time.Now().UTC().Add(-11 * time.Hour)
	timeMinus10 := time.Now().UTC().Add(-10 * time.Hour)
	timeMinus8 := time.Now().UTC().Add(-8 * time.Hour)
	timeMinus6 := time.Now().UTC().Add(-6 * time.Hour)
	timeMinus4 := time.Now().UTC().Add(-4 * time.Hour)
	timeMinus3 := time.Now().UTC().Add(-3 * time.Hour)
	timeMinus2 := time.Now().UTC().Add(-2 * time.Hour)
	timeMinus1 := time.Now().UTC().Add(-1 * time.Hour)

	testLogs := []HistoryLog{
		{
			ID:      uuid.New(),
			Level:   "TestLogsByTime",
			Service: "TestLogsByTime",
			UserID:  testhelpers.NewUUIDPtr(),
			DataRaw: "some data",
			Created: timeMinus12,
		},
		{
			ID:      uuid.New(),
			Level:   "TestLogsByTime",
			Service: "TestLogsByTime",
			UserID:  testhelpers.NewUUIDPtr(),
			DataRaw: "some data",
			Created: timeMinus10,
		},
		{
			ID:      uuid.New(),
			Level:   "TestLogsByTime",
			Service: "TestLogsByTime",
			UserID:  testhelpers.NewUUIDPtr(),
			DataRaw: "some data",
			Created: timeMinus8,
		},
		{
			ID:      uuid.New(),
			Level:   "TestLogsByTime",
			Service: "TestLogsByTime",
			UserID:  testhelpers.NewUUIDPtr(),
			DataRaw: "some data",
			Created: timeMinus6,
		},
		{
			ID:      uuid.New(),
			Level:   "TestLogsByTime",
			Service: "TestLogsByTime",
			UserID:  testhelpers.NewUUIDPtr(),
			DataRaw: "some data",
			Created: timeMinus4,
		},
		{
			ID:      uuid.New(),
			Level:   "TestLogsByTime",
			Service: "TestLogsByTime",
			UserID:  testhelpers.NewUUIDPtr(),
			DataRaw: "some data",
			Created: timeMinus2,
		},
	}

	for i := range testLogs {
		err := suite.repository.Create(suite.ctx, &testLogs[i])
		assert.NoError(t, err)
	}

	var respBody []HistoryLog

	req, _ := http.NewRequest("GET", "/test/logs?start="+timeMinus12.Format(TimeFormat)+"&end="+timeMinus1.Format(TimeFormat), nil)
	resp := suite.performRequest(req)
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 6, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?end="+timeMinus3.Format(TimeFormat), nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 5, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?start="+timeMinus11.Format(TimeFormat)+"&end="+timeMinus3.Format(TimeFormat), nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 4, len(respBody))

}

func (suite *AdminHandlerTestSuite) TestLogsLimitSkip() {
	t := suite.T()

	testLogs := []HistoryLog{
		{
			Level:   "TestLogsLimitSkip",
			Service: "TestLogsLimitSkip",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsLimitSkip",
			Service: "TestLogsLimitSkip",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsLimitSkip",
			Service: "TestLogsLimitSkip",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsLimitSkip",
			Service: "TestLogsLimitSkip",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsLimitSkip",
			Service: "TestLogsLimitSkip",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key1": "value1",
			},
		},
		{
			Level:   "TestLogsLimitSkip",
			Service: "TestLogsLimitSkip",
			UserID:  testhelpers.NewUUIDPtr(),
			Data: map[string]any{
				"key2": "value2",
			},
		},
	}

	for i := range testLogs {
		createdLog, err := suite.service.Create(suite.ctx, testLogs[i].Level, testLogs[i].Service, testLogs[i].UserID, testLogs[i].Data)
		assert.NoError(t, err)
		testLogs[i] = *createdLog
	}

	var respBody []HistoryLog

	req, _ := http.NewRequest("GET", "/test/logs?level=TestLogsLimitSkip&limit=5", nil)
	resp := suite.performRequest(req)
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 5, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsLimitSkip&limit=5&skip=1", nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 5, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsLimitSkip&limit=5&skip=2", nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 4, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsLimitSkip&limit=2&skip=5", nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 1, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsLimitSkip&limit=2&skip=10", nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 0, len(respBody))

	req, _ = http.NewRequest("GET", "/test/logs?level=TestLogsLimitSkip&limit=7", nil)
	resp = suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, 6, len(respBody))
}

func (suite *AdminHandlerTestSuite) TestLogsById() {
	t := suite.T()

	data := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}

	testLog, err := suite.service.Create(suite.ctx, "TestLogsById", "TestLogsById", testhelpers.NewUUIDPtr(), data)
	assert.NoError(t, err)

	var respBody HistoryLog

	req, _ := http.NewRequest("GET", "/test/logs/"+testLog.ID.String(), nil)
	resp := suite.performRequest(req)
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, testLog.Level, respBody.Level)
	assert.Equal(t, testLog.Service, respBody.Service)
	assert.Equal(t, testLog.UserID, respBody.UserID)
	assert.Equal(t, testLog.Data, respBody.Data)
}
