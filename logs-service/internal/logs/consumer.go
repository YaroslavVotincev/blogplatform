package logs

import (
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"hidepost-history-logs-consumer/pkg/filelogger"
)

type Consumer struct {
	service    *Service
	conn       *amqp.Connection
	channel    *amqp.Channel
	closeChan  chan struct{}
	queue      string
	fileLogger *filelogger.FileLogger
}

func NewConsumer(service *Service, mqConn *amqp.Connection, queue string, fileLogger *filelogger.FileLogger) (*Consumer, error) {
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
		service:    service,
		conn:       mqConn,
		channel:    ch,
		queue:      queue,
		fileLogger: fileLogger,
	}
	return consumer, err
}

func (c *Consumer) Start() error {
	messages, err := c.channel.Consume(
		c.queue,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	go c.listener(messages)
	return nil
}

func (c *Consumer) Close() {
	c.closeChan <- struct{}{}
	<-c.closeChan
	_ = c.channel.Close()
}

func (c *Consumer) listener(messages <-chan amqp.Delivery) {
	for {
		select {
		case msg := <-messages:
			go c.handleMessage(msg)
		case <-c.closeChan:
			c.closeChan <- struct{}{}
			return
		}
	}
}

func (c *Consumer) handleMessage(msg amqp.Delivery) {
	var logMsg LogMessage
	err := json.Unmarshal(msg.Body, &logMsg)
	if err != nil {
		err1 := c.service.FailedToUnmarshall(msg.Body, err)
		if err1 != nil {
			c.fileLogger.Error("failed to handle message: message couldn't be unmarshalled",
				map[string]interface{}{"error": err1.Error()})
		}
		err2 := msg.Ack(false)
		if err2 != nil {
			c.fileLogger.Error("failed to handle message: failed to acknowledge bad message from queue",
				map[string]interface{}{"error": err2.Error()})
		}
		return
	}

	err = c.service.ValidateLogMessage(&logMsg)
	if err != nil {
		err1 := c.service.FailedToValidate(msg.Body, err)
		if err1 != nil {
			c.fileLogger.Error("failed to handle message: failed to record that message couldn't be validated",
				map[string]interface{}{"error": err1.Error()})
		}
		err2 := msg.Ack(false)
		if err2 != nil {
			c.fileLogger.Error("failed to handle message: failed to acknowledge bad message from queue",
				map[string]interface{}{"error": err2.Error()})
		}
		return
	}

	err = c.service.Create(&logMsg)
	if err != nil {
		c.fileLogger.Error("failed to handle message: failed to record message",
			map[string]interface{}{"error": err.Error()})
		_ = msg.Nack(false, true)
	} else {
		_ = msg.Ack(false)
	}
}
