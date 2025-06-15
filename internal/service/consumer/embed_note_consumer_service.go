package consumer

import (
	embeddingentity "ai-notetaking-be/internal/entity/embedding"
	embeddingrepository "ai-notetaking-be/internal/repository/embedding"
	noteservice "ai-notetaking-be/internal/service/note"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EmbeddingModelRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingModelResponse struct {
	Embedding []float32 `json:"embedding"`
}

type IEmbedNoteConsumerService interface {
	Consume(ctx context.Context) error
}

type embedNoteConsumerService struct {
	ch *amqp.Channel
	q  amqp.Queue

	embeddingRepository embeddingrepository.IEmbeddingRepository
}

func (mq *embedNoteConsumerService) Consume(ctx context.Context) error {
	msgs, err := mq.ch.ConsumeWithContext(
		ctx,
		"",
		mq.q.Name,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for msg := range msgs {
		var dest noteservice.EmbedCreatedNoteMessage
		err = json.Unmarshal(msg.Body, &dest)
		if err != nil {
			return err
		}

		req := EmbeddingModelRequest{
			Model:  "nomic-embed-text",
			Prompt: dest.Content,
		}
		reqJson, _ := json.Marshal(req)
		res, err := http.Post("http://localhost:11434/api/embeddings", "application/json", bytes.NewBuffer(reqJson))
		if err != nil {
			log.Println(err)
			return err
		}

		var embeddingResponse EmbeddingModelResponse
		err = json.NewDecoder(res.Body).Decode(&embeddingResponse)
		if err != nil {
			log.Println(err)
			return err
		}

		embeddingText := embeddingentity.NoteEmbedding{
			Id:           uuid.New(),
			NoteId:       dest.NoteId,
			OriginalText: dest.Content,
			Embedding:    embeddingResponse.Embedding,
			CreatedAt:    time.Now(),
			CreatedBy:    "System",
		}
		err = mq.embeddingRepository.CreateNoteEmbedding(ctx, &embeddingText)
		if err != nil {
			return err
		}

		err = msg.Ack(false)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewEmbedNoteConsumerService(connectionString string, queueName string, embeddingRepository embeddingrepository.IEmbeddingRepository) IEmbedNoteConsumerService {
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

	return &embedNoteConsumerService{
		ch:                  ch,
		q:                   q,
		embeddingRepository: embeddingRepository,
	}
}
