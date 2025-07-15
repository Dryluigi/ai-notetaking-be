package publisher

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type IPublisherService interface {
	Publish(ctx context.Context, payload []byte) error
}

type rabbitMqPublisherService struct {
	ch *amqp.Channel
	q  amqp.Queue
}

func (mq *rabbitMqPublisherService) Publish(ctx context.Context, payload []byte) error {
	err := mq.ch.PublishWithContext(
		ctx,
		"",
		mq.q.Name,
		true,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        payload,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func NewRabbitMqPublisherService(connectionString string, queueName string) IPublisherService {
	conn, err := amqp.Dial(connectionString)
	if err != nil {
		panic(fmt.Sprintf("RabbitMQ connection error, %s", err))
	}

	ch, err := conn.Channel()
	if err != nil {
		panic(fmt.Sprintf("RabbitMQ creating channel error, %s", err))
	}

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		panic(fmt.Sprintf("RabbitMQ declaring queue error, %s", err))
	}

	return &rabbitMqPublisherService{
		ch: ch,
		q:  q,
	}
}
