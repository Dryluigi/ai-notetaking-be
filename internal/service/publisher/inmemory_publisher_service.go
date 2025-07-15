package publisher

import (
	"context"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

type inMemoryPublisherService struct {
	queueName string
	pubSub    *gochannel.GoChannel
}

func (mq *inMemoryPublisherService) Publish(ctx context.Context, payload []byte) error {
	err := mq.pubSub.Publish(mq.queueName, message.NewMessage(watermill.NewUUID(), payload))
	if err != nil {
		return err
	}

	return nil
}

func NewInMemoryPublisherService(pubSub *gochannel.GoChannel, queueName string) IPublisherService {
	return &inMemoryPublisherService{
		pubSub:    pubSub,
		queueName: queueName,
	}
}
