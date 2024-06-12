package notifications

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"notifications-service/pkg/filelogger"
	"notifications-service/pkg/queuelogger"
	"strings"
	"time"
)

const (
	ConsumerName = "notifications-service"
)

type Consumer struct {
	service      *Service
	conn         *amqp.Connection
	channel      *amqp.Channel
	closeChan    chan struct{}
	messagesChan <-chan amqp.Delivery
	queue        string
	fileLogger   *filelogger.FileLogger
	queueLogger  *queuelogger.RemoteLogger
}

func NewConsumer(service *Service, mqConn *amqp.Connection, queue string,
	fileLogger *filelogger.FileLogger, queueLogger *queuelogger.RemoteLogger) (*Consumer, error) {
	ch, err := mqConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open mq channel: %v", err)
	}

	// check if queue exists
	_, err = ch.QueueDeclarePassive(queue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("mq queue might not exist: %v", err)
	}

	consumer := &Consumer{
		service:      service,
		conn:         mqConn,
		channel:      ch,
		closeChan:    make(chan struct{}),
		messagesChan: make(chan amqp.Delivery),
		queue:        queue,
		fileLogger:   fileLogger,
		queueLogger:  queueLogger,
	}

	return consumer, err
}

func (c *Consumer) Start() error {
	var err error
	c.messagesChan, err = c.channel.Consume(
		c.queue,
		ConsumerName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	go c.listen()
	return nil
}

func (c *Consumer) Close() {
	_ = c.channel.Cancel(ConsumerName, false)
	c.closeChan <- struct{}{}
	for {
		select {
		case msg := <-c.messagesChan:
			c.handleMessage(msg)
		default:
			_ = c.channel.Close()
			return
		}
	}
}

func (c *Consumer) listen() {
	for {
		select {
		case msg := <-c.messagesChan:
			c.handleMessage(msg)
		case <-c.closeChan:
			return
		}
	}
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	loggingMap := map[string]any{}
	timeNow := time.Now().UTC()

	userIdStr, ok := msg.Headers["user_id"].(string)
	if !ok {
		loggingMap["message"] = "failed to get user id from queue message header"
		c.fileLogger.Error("", loggingMap)
		_ = c.queueLogger.Error(nil, loggingMap)
		return
	}
	loggingMap["user_id_str"] = userIdStr
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		loggingMap["message"] = "failed to parse user id from queue message header"
		loggingMap["error"] = err.Error()
		c.fileLogger.Error("", loggingMap)
		_ = c.queueLogger.Error(nil, loggingMap)
		return
	}

	eventCode, ok := msg.Headers["event"].(string)
	if !ok {
		loggingMap["message"] = "failed to get event code from queue message header"
		loggingMap["error"] = err.Error()
		c.fileLogger.Error("", loggingMap)
		_ = c.queueLogger.Error(nil, loggingMap)
		return
	}
	eventCode = strings.ToUpper(eventCode)

	loggingMap["body_bytes"] = string(msg.Body)

	notification := Notification{
		ID:        uuid.New(),
		EventCode: eventCode,
		UserID:    userId,
		Seen:      false,
		Data:      nil,
		DataBytes: msg.Body,
		Created:   timeNow,
		Updated:   timeNow,
	}

	err = c.service.ValidateNotificationData(&notification)
	if err != nil {
		loggingMap["message"] = "failed to validate notification data"
		loggingMap["error"] = err.Error()
		c.fileLogger.Error("", loggingMap)
		_ = c.queueLogger.Error(nil, loggingMap)
		return
	}

	err = c.service.repository.Create(context.Background(), notification)
	if err != nil {
		loggingMap["message"] = "failed to create notification"
		loggingMap["error"] = err.Error()
		c.fileLogger.Error("", loggingMap)
		_ = c.queueLogger.Error(nil, loggingMap)
		return
	}

	loggingMap["user_id"] = userId
	loggingMap["event_code"] = eventCode
	loggingMap["message"] = "successfully processed queue message"
	c.fileLogger.Info("", loggingMap)
	_ = c.queueLogger.Info(nil, loggingMap)
}
