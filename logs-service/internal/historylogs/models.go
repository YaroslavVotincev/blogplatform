package historylogs

import (
	"time"

	"github.com/google/uuid"
)

type HistoryLog struct {
	ID      uuid.UUID      `json:"id"`
	Level   string         `json:"level"`
	Service string         `json:"service"`
	UserID  *uuid.UUID     `json:"user_id"`
	Data    map[string]any `json:"data"`
	DataRaw string         `json:"-"`
	Created time.Time      `json:"created"`
}

const (
	TimeFormat = "2006-01-02T15:04"
)
