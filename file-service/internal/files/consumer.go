package files

import (
	"file-service/pkg/filelogger"
	"file-service/pkg/queuelogger"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ConsumerName = "files-service"
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

	fileId, ok := msg.Headers["id"].(string)
	if !ok {
		loggingMap["message"] = "failed to get id from queue message header"
		c.fileLogger.Error("", loggingMap)
		_ = c.queueLogger.Error(nil, loggingMap)
		//_ = msg.Ack(false)
		return
	}
	loggingMap["file_id"] = fileId
	method, ok := msg.Headers["method"].(string)
	if !ok {
		loggingMap["message"] = "failed to get method from queue message header"
		c.fileLogger.Error("", loggingMap)
		_ = c.queueLogger.Error(nil, loggingMap)
		//_ = msg.Ack(false)
		return
	}
	loggingMap["method"] = method

	switch method {
	case "set":
		err := c.service.SetFile(fileId, msg.Body)
		if err != nil {
			loggingMap["message"] = "failed to set file"
			loggingMap["error"] = err.Error()
			c.fileLogger.Error("error occurred", loggingMap)
			_ = c.queueLogger.Error(nil, loggingMap)
			//_ = msg.Ack(false)
			return
		}
	case "delete":
		err := c.service.DeleteFile(fileId)
		if err != nil {
			loggingMap["message"] = "failed to delete file"
			loggingMap["error"] = err.Error()
			c.fileLogger.Error("error occurred", loggingMap)
			_ = c.queueLogger.Error(nil, loggingMap)
			//_ = msg.Ack(false)
			return
		}
	default:
		loggingMap["message"] = "unknown method on queue message header"
		c.fileLogger.Error("", loggingMap)
		_ = c.queueLogger.Error(nil, loggingMap)
		//_ = msg.Ack(false)
		return
	}
}
