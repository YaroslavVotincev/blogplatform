package queuelogger

import (
	"github.com/google/uuid"
)

type Mock struct{}

func (Mock) Debug(userId *uuid.UUID, data map[string]any) error {
	return nil
}

func (Mock) Error(userId *uuid.UUID, data map[string]any) error {
	return nil
}

func (Mock) Info(userId *uuid.UUID, data map[string]any) error {
	return nil
}

func (Mock) Warn(userId *uuid.UUID, data map[string]any) error {
	return nil
}
