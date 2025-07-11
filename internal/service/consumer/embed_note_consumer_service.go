package consumer

import (
	embeddingentity "ai-notetaking-be/internal/entity/embedding"
	embeddingrepository "ai-notetaking-be/internal/repository/embedding"
	noterepository "ai-notetaking-be/internal/repository/note"
	noteservice "ai-notetaking-be/internal/service/note"
	"ai-notetaking-be/pkg/gemini"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	embeddingServerBaseUrl string
	embeddingModelName     string

	ch *amqp.Channel
	q  amqp.Queue

	semaphore     chan struct{}
	maxConcurrent int

	embeddingRepository embeddingrepository.IEmbeddingRepository
	noterepository      noterepository.INoteRepository
	db                  *pgxpool.Pool
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
		mq.semaphore <- struct{}{}

		go mq.processMessage(ctx, msg)
	}

	return nil
}

func (mq *embedNoteConsumerService) processMessage(ctx context.Context, msg amqp.Delivery) error {
	defer func() {
		<-mq.semaphore
	}()
	var err error
	defer func() {
		if err != nil {
			nackErr := msg.Nack(false, false)
			if nackErr != nil {
				panic(nackErr)
			}
		}
	}()
	var dest noteservice.EmbedCreatedNoteMessage
	err = json.Unmarshal(msg.Body, &dest)
	if err != nil {
		return err
	}

	tx, err := mq.db.Begin(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	defer func() {
		if err != nil {
			log.Println(err)
			rollbackErr := tx.Rollback(ctx)
			if rollbackErr != nil {
				panic(rollbackErr)
			}
		}
	}()
	embedRepo := mq.embeddingRepository.UsingTx(ctx, tx)
	noteRepo := mq.noterepository.UsingTx(ctx, tx)

	note, err := noteRepo.GetById(ctx, dest.NoteId)
	if err != nil {
		return err
	}
	if dest.DeleteOldEmbedding {
		err = embedRepo.DeleteNoteEmbeddings(ctx, dest.NoteId, "System")
		if err != nil {
			return err
		}
	}

	notebookName := "-"
	if note.Notebook != nil {
		notebookName = note.Notebook.Name
	}
	document := fmt.Sprintf(
		`Notebook: %s\nTitle: %s\nContent: %s\nCreated at: %s`,
		notebookName,
		note.Title,
		note.Content,
		note.CreatedAt.Format(time.RFC3339),
	)
	// req := EmbeddingModelRequest{
	// 	Model:  mq.embeddingModelName,
	// 	Prompt: document,
	// }
	// reqJson, _ := json.Marshal(req)
	// res, err := http.Post(fmt.Sprintf("%s/api/embeddings", mq.embeddingServerBaseUrl), "application/json", bytes.NewBuffer(reqJson))
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

	// var embeddingResponse EmbeddingModelResponse
	// err = json.NewDecoder(res.Body).Decode(&embeddingResponse)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

	embeddingValue, err := gemini.GetEmbedding(os.Getenv("GEMINI_API_KEY"), document, "RETRIEVAL_DOCUMENT")
	if err != nil {
		log.Println(err)
		return err
	}

	embeddingText := embeddingentity.NoteEmbedding{
		Id:           uuid.New(),
		NoteId:       dest.NoteId,
		OriginalText: document,
		Embedding:    embeddingValue.Embedding.Values,
		CreatedAt:    time.Now(),
		CreatedBy:    "System",
	}
	err = embedRepo.CreateNoteEmbedding(ctx, &embeddingText)
	if err != nil {
		log.Println(err)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	err = msg.Ack(false)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func NewEmbedNoteConsumerService(
	connectionString string,
	queueName string,
	db *pgxpool.Pool,
	embeddingServerBaseUrl string,
	embeddingModelName string,
	embeddingRepository embeddingrepository.IEmbeddingRepository,
	noteRepository noterepository.INoteRepository,
) IEmbedNoteConsumerService {
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
		ch:                     ch,
		q:                      q,
		embeddingServerBaseUrl: embeddingServerBaseUrl,
		embeddingModelName:     embeddingModelName,
		db:                     db,
		maxConcurrent:          100,
		semaphore:              make(chan struct{}, 100),
		embeddingRepository:    embeddingRepository,
		noterepository:         noteRepository,
	}
}
