package main

import (
	embeddingrepository "ai-notetaking-be/internal/repository/embedding"
	noterepository "ai-notetaking-be/internal/repository/note"
	consumerservice "ai-notetaking-be/internal/service/consumer"
	"ai-notetaking-be/pkg/database"
	"context"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	ctx := context.Background()

	db := database.ConnectDB(os.Getenv("DB_CONNECTION_STRING"))

	noteRepository := noterepository.NewNoteRepository(db)
	embeddingRepository := embeddingrepository.NewEmbeddingRepository(db)
	consumer := consumerservice.NewEmbedNoteConsumerService(
		os.Getenv("RABBITMQ_CONNECTION_STRING"),
		"embed-note-content",
		db,
		os.Getenv("EMBEDDING_SERVER_BASE_URL"),
		os.Getenv("EMBEDDING_MODEL_NAME"),
		embeddingRepository,
		noteRepository,
	)
	err := consumer.Consume(ctx)
	if err != nil {
		panic(err)
	}
}
