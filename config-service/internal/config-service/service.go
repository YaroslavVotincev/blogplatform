package configservice

import (
	"context"
	"time"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) NewService(ctx context.Context, name string) (*ServiceModel, error) {
	now := time.Now().UTC()
	service := &ServiceModel{Service: name, Created: now, Updated: now}
	err := s.repository.CreateService(ctx, service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (s *Service) AllServices(ctx context.Context) ([]ServiceModel, error) {
	return s.repository.AllServices(ctx)
}

func (s *Service) ServiceByName(ctx context.Context, name string) (*ServiceModel, error) {
	return s.repository.ServiceByName(ctx, name)
}

func (s *Service) UpdateService(ctx context.Context, service *ServiceModel, newName string) error {
	err := s.repository.UpdateService(ctx, service.Service, newName)
	service.Service = newName
	service.Updated = time.Now().UTC()
	return err
}

func (s *Service) DeleteService(ctx context.Context, name string) error {
	return s.repository.DeleteService(ctx, name)
}

func (s *Service) SettingsByService(ctx context.Context, service string) ([]Setting, error) {
	return s.repository.SettingsByService(ctx, service)
}

func (s *Service) SetSettingsToService(ctx context.Context, service string, items []CreateSettingRequest) error {
	timeNow := time.Now().UTC()
	newItems := make([]Setting, len(items))
	for i := range items {
		newItems[i] = Setting{
			Key:     items[i].Key,
			Value:   items[i].Value,
			Updated: timeNow,
			Created: timeNow,
			Service: service,
		}
	}
	return s.repository.SetSettingsToService(ctx, service, newItems)
}
