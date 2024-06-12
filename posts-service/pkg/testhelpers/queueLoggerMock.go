package testhelpers

import "github.com/google/uuid"

type QueueLoggerMock struct{}

func (QueueLoggerMock) Debug(service string, userId *uuid.UUID, data map[string]any) error {
	return nil
}

func (QueueLoggerMock) Error(service string, userId *uuid.UUID, data map[string]any) error {
	return nil
}

func (QueueLoggerMock) Info(service string, userId *uuid.UUID, data map[string]any) error {
	return nil
}

func (QueueLoggerMock) Warning(service string, userId *uuid.UUID, data map[string]any) error {
	return nil
}
