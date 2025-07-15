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

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type embedNoteInMemoryConsumerService struct {
	queueName              string
	embeddingServerBaseUrl string
	embeddingModelName     string

	pubSub *gochannel.GoChannel

	semaphore     chan struct{}
	maxConcurrent int

	embeddingRepository embeddingrepository.IEmbeddingRepository
	noterepository      noterepository.INoteRepository
	db                  *pgxpool.Pool
}

func (mq *embedNoteInMemoryConsumerService) Consume(ctx context.Context) error {
	messages, err := mq.pubSub.Subscribe(ctx, mq.queueName)
	if err != nil {
		return err
	}

	// go func() {
	// 	for msg := range messages {
	// 		log.Print("Consuming...")
	// 		mq.processMessage(ctx, msg)
	// 		log.Print("Finished consuming.")
	// 	}
	// }()

	// Concurrent
	go func() {
		for msg := range messages {
			log.Print("Consuming...")
			mq.semaphore <- struct{}{}

			go mq.processMessage(ctx, msg)
			log.Print("Finished consuming.")
		}
	}()

	return nil
}

func (mq *embedNoteInMemoryConsumerService) processMessage(ctx context.Context, msg *message.Message) error {
	// Activate if concurrent
	defer func() {
		<-mq.semaphore
	}()
	defer msg.Nack()

	var dest noteservice.EmbedCreatedNoteMessage
	err := json.Unmarshal(msg.Payload, &dest)
	if err != nil {
		return err
	}

	tx, err := mq.db.Begin(ctx)
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback(ctx)

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
	msg.Ack()

	return nil
}

func NewInMemoryConsumer(
	pubSub *gochannel.GoChannel,
	queueName string,
	db *pgxpool.Pool,
	embeddingServerBaseUrl string,
	embeddingModelName string,
	embeddingRepository embeddingrepository.IEmbeddingRepository,
	noteRepository noterepository.INoteRepository,
) IEmbedNoteConsumerService {
	return &embedNoteInMemoryConsumerService{
		queueName:              queueName,
		embeddingServerBaseUrl: embeddingServerBaseUrl,
		embeddingModelName:     embeddingModelName,
		db:                     db,
		maxConcurrent:          100,
		semaphore:              make(chan struct{}, 100),
		embeddingRepository:    embeddingRepository,
		noterepository:         noteRepository,
		pubSub:                 pubSub,
	}
}
