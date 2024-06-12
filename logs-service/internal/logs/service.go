package logs

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"time"
)

const (
	DebugLevel   = "debug"
	InfoLevel    = "info"
	WarningLevel = "warn"
	ErrorLevel   = "error"
)

type Service struct {
	repository  *Repository
	ctx         context.Context
	serviceName string
}

func NewService(repository *Repository, serviceName string) *Service {
	return &Service{repository: repository, ctx: context.TODO(), serviceName: serviceName}
}

func (s *Service) FailedToUnmarshall(msg []byte, err error) error {
	data := make(map[string]any)
	data["log_bytes"] = string(msg)
	data["error"] = err.Error()
	data["message"] = "failed to unmarshall log message from log queue"
	dataRaw, _ := json.Marshal(data)
	logObj := &HistoryLog{
		ID:      uuid.New(),
		Level:   ErrorLevel,
		Service: s.serviceName,
		UserID:  nil,
		Data:    data,
		DataRaw: string(dataRaw),
		Created: time.Now().UTC(),
	}
	return s.repository.Create(s.ctx, logObj)
}

func (s *Service) ValidateLogMessage(msg *LogMessage) error {
	if msg.Level == "" {
		return fmt.Errorf("message level cannot be empty")
	}
	if msg.Service == "" {
		return fmt.Errorf("message service cannot be empty")
	}
	if !(msg.Level == DebugLevel || msg.Level == InfoLevel || msg.Level == WarningLevel || msg.Level == ErrorLevel) {
		return fmt.Errorf("message level must be one of %s, %s, %s, %s", DebugLevel, InfoLevel, WarningLevel, ErrorLevel)
	}
	if msg.Created == nil {
		t := time.Now().UTC()
		msg.Created = &t
	}
	return nil
}

func (s *Service) FailedToValidate(msg []byte, err error) error {
	data := make(map[string]any)
	data["log_bytes"] = string(msg)
	data["error"] = err.Error()
	data["message"] = "failed to validate log message from log queue"
	dataRaw, _ := json.Marshal(data)
	logObj := &HistoryLog{
		ID:      uuid.New(),
		Level:   ErrorLevel,
		Service: s.serviceName,
		UserID:  nil,
		Data:    data,
		DataRaw: string(dataRaw),
		Created: time.Now().UTC(),
	}
	return s.repository.Create(s.ctx, logObj)
}

func (s *Service) Create(logMsg *LogMessage) error {
	dataRaw, err := json.Marshal(logMsg.Data)
	if err != nil {
		return err
	}
	historyLog := &HistoryLog{
		ID:      uuid.New(),
		Level:   logMsg.Level,
		Service: logMsg.Service,
		UserID:  logMsg.UserID,
		Data:    logMsg.Data,
		DataRaw: string(dataRaw),
		Created: *logMsg.Created,
	}
	return s.repository.Create(s.ctx, historyLog)
}
