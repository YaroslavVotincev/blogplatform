package logs

import (
	"github.com/google/uuid"
	"time"
)

type LogMessage struct {
	Level   string         `json:"level"`
	Service string         `json:"service"`
	UserID  *uuid.UUID     `json:"user_id"`
	Data    map[string]any `json:"data"`
	Created *time.Time     `json:"created"`
}
