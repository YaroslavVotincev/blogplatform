package historylogs

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	service := Service{repository: repository}
	return &service
}

func (s *Service) Count(ctx context.Context) (int, error) {
	return s.repository.Count(ctx)
}

func (s *Service) All(ctx context.Context, levels, services []string, userIDs []*uuid.UUID, startTime, endTime *time.Time, limit, skip *int) ([]HistoryLog, error) {
	logs, err := s.repository.All(ctx, levels, services, userIDs, startTime, endTime, limit, skip)
	if err != nil {
		return nil, err
	}
	for i := range logs {
		logs[i].Data = make(map[string]any)
		if err = json.Unmarshal([]byte(logs[i].DataRaw), &logs[i].Data); err != nil {
			logs[i].Data = make(map[string]any)
			logs[i].Data["history_logs_service_error"] = err
			logs[i].Data["history_logs_service_message"] = "failed to unmarshal raw data from db, attaching raw string here instead"
			logs[i].Data["log_raw"] = logs[i].DataRaw
		}
	}
	return logs, nil
}

func (s *Service) ByID(ctx context.Context, id uuid.UUID) (*HistoryLog, error) {
	log, err := s.repository.ByID(ctx, id)
	if err != nil {
		return nil, err
	}
	log.Data = make(map[string]any)
	if err = json.Unmarshal([]byte(log.DataRaw), &log.Data); err != nil {
		log.Data = make(map[string]any)
		log.Data["history_logs_service_error"] = err
		log.Data["history_logs_service_message"] = "failed to unmarshal raw data from db, attaching raw string here instead"
		log.Data["log_raw"] = log.DataRaw
	}
	return log, nil
}

func (s *Service) Create(ctx context.Context, level, service string, userID *uuid.UUID, data map[string]any) (*HistoryLog, error) {
	dataRaw, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	historyLog := &HistoryLog{
		ID:      uuid.New(),
		Level:   level,
		Service: service,
		UserID:  userID,
		Data:    data,
		DataRaw: string(dataRaw),
		Created: time.Now().UTC(),
	}
	return historyLog, s.repository.Create(ctx, historyLog)
}

func (s *Service) Levels(ctx context.Context) ([]string, error) {
	return s.repository.Levels(ctx)
}

func (s *Service) Services(ctx context.Context) ([]string, error) {
	return s.repository.Services(ctx)
}
