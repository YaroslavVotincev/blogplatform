package queuelogger

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

type logMessage struct {
	Level   string         `json:"level"`
	Service string         `json:"service"`
	UserId  *uuid.UUID     `json:"user_id,omitempty"`
	Data    map[string]any `json:"data"`
	Created time.Time      `json:"created"`
}

type QueueLogger interface {
	Info(userId *uuid.UUID, data map[string]any) error  // level = info
	Debug(userId *uuid.UUID, data map[string]any) error // level = debug
	Warn(userId *uuid.UUID, data map[string]any) error  // level = warning
	Error(userId *uuid.UUID, data map[string]any) error
}

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

type RemoteLogger struct {
	service string
	queue   string
	conn    *amqp.Connection
	ch      *amqp.Channel
}

func NewRemoteLogger(url string, queue string, service string) (*RemoteLogger, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	return &RemoteLogger{
		service: service,
		queue:   queue,
		conn:    conn,
		ch:      ch,
	}, nil
}

func (r *RemoteLogger) Debug(userId *uuid.UUID, data map[string]any) error {
	return r.send(logMessage{
		Level:   LogLevelDebug,
		Service: r.service,
		UserId:  userId,
		Data:    data,
		Created: time.Now().UTC(),
	})
}

func (r *RemoteLogger) Info(userId *uuid.UUID, data map[string]any) error {
	return r.send(logMessage{
		Level:   LogLevelInfo,
		Service: r.service,
		UserId:  userId,
		Data:    data,
		Created: time.Now().UTC(),
	})
}

func (r *RemoteLogger) Warn(userId *uuid.UUID, data map[string]any) error {
	return r.send(logMessage{
		Level:   LogLevelWarn,
		Service: r.service,
		UserId:  userId,
		Data:    data,
		Created: time.Now().UTC(),
	})
}

func (r *RemoteLogger) Error(userId *uuid.UUID, data map[string]any) error {
	return r.send(logMessage{
		Level:   LogLevelError,
		Service: r.service,
		UserId:  userId,
		Data:    data,
		Created: time.Now().UTC(),
	})
}

func (r *RemoteLogger) send(msg logMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.ch.PublishWithContext(ctx,
		"",
		r.queue,
		false,
		false,
		amqp.Publishing{ContentType: "text/plain", Body: data})
}

func (r *RemoteLogger) Close() {
	_ = r.ch.Close()
	_ = r.conn.Close()
}
