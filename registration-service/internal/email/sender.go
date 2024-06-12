package email

import (
	"context"
	"encoding/json"
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type Sender struct {
	ctx   context.Context
	conn  *amqp.Connection
	ch    *amqp.Channel
	queue string
}

func NewSender(ctx context.Context, mqUrl string, emailQueue string) (*Sender, error) {
	mqConn, err := amqp.Dial(mqUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %v", err)
	}

	ch, err := mqConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// check if queue exists
	_, err = ch.QueueDeclarePassive(emailQueue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("mq queue might not exist: %v", err)
	}

	sender := &Sender{
		ctx:   ctx,
		conn:  mqConn,
		ch:    ch,
		queue: emailQueue,
	}

	return sender, nil
}

func (s *Sender) handleMessage(msg QueueMessage) error {

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}
	cancelCtx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	err = s.ch.PublishWithContext(
		cancelCtx,
		"",
		s.queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        msgBytes,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}
	return nil
}

func (s *Sender) Close() {
	_ = s.ch.Close()
	_ = s.conn.Close()
}
