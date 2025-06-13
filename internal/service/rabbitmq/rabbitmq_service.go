package rabbitmq

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

type IRabbitMqService interface {
	Publish(ctx context.Context, payload []byte) error
}

type rabbitMqService struct {
	ch *amqp.Channel
	q  amqp.Queue
}

func (mq *rabbitMqService) Publish(ctx context.Context, payload []byte) error {
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

func NewRabbitMqService(connectionString string, queueName string) IRabbitMqService {
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

	return &rabbitMqService{
		ch: ch,
		q:  q,
	}
}
