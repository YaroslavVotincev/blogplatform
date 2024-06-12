package notifications

import (
	"context"
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

func NewSender(ctx context.Context, mqUrl string, queue string) (*Sender, error) {
	mqConn, err := amqp.Dial(mqUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %v", err)
	}

	ch, err := mqConn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open channel: %v", err)
	}

	// check if queue exists
	_, err = ch.QueueDeclarePassive(queue, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("mq queue might not exist: %v", err)
	}

	sender := &Sender{
		ctx:   ctx,
		conn:  mqConn,
		ch:    ch,
		queue: queue,
	}

	return sender, nil
}

func (s *Sender) publishMessage(userId string, event string, body []byte) error {

	headers := make(amqp.Table)
	headers["user_id"] = userId
	headers["event"] = event

	cancelCtx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()
	var err error
	if body != nil {
		err = s.ch.PublishWithContext(
			cancelCtx,
			"",
			s.queue,
			false,
			false,
			amqp.Publishing{
				Headers: headers,
				Body:    body,
			},
		)
	} else {
		err = s.ch.PublishWithContext(
			cancelCtx,
			"",
			s.queue,
			false,
			false,
			amqp.Publishing{
				Headers: headers,
			},
		)
	}

	return err
}

func (s *Sender) Close() {
	_ = s.ch.Close()
	_ = s.conn.Close()
}
