package email

import (
	"email-service/pkg/filelogger"
	"email-service/pkg/queuelogger"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	EmailConsumerName = "email-service"
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
		EmailConsumerName,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	go c.listener()
	return nil
}

func (c *Consumer) Close() {
	_ = c.channel.Cancel(EmailConsumerName, false)
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

func (c *Consumer) listener() {
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
	msgBodyBytes := msg.Body
	loggingMap["body_bytes"] = string(msgBodyBytes)

	var emailMsg QueueMessage
	err := json.Unmarshal(msgBodyBytes, &emailMsg)
	if err != nil {
		loggingMap["message"] = "failed to unmarshal email message"
		loggingMap["error"] = err.Error()
		c.fileLogger.Error("error occurred", loggingMap)
		err1 := c.queueLogger.Error(nil, loggingMap)
		if err1 != nil {
			c.fileLogger.Error("failed to log email msg unmarshal error into log queue", map[string]any{
				"error": err1.Error(),
			})
		}
		_ = msg.Ack(false)
		return
	}

	err = c.service.SendEmail(emailMsg)
	if err != nil {
		loggingMap["message"] = "failed to send email"
		loggingMap["error"] = err.Error()
		c.fileLogger.Error("error occurred", loggingMap)
		err1 := c.queueLogger.Error(nil, loggingMap)
		if err1 != nil {
			c.fileLogger.Error("failed to log email send error into log queue", map[string]any{
				"error": err1.Error(),
			})
		}
		_ = msg.Nack(false, true)
		return
	}

	_ = msg.Ack(false)
	loggingMap["message"] = "new email sent"
	c.fileLogger.Info("success", loggingMap)
	err = c.queueLogger.Info(nil, loggingMap)
	if err != nil {
		c.fileLogger.Error("failed to log success email send into log queue", map[string]any{
			"error": err.Error(),
		})
	}
}
