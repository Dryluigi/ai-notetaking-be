package main

import (
	embeddingrepository "ai-notetaking-be/internal/repository/embedding"
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

	embeddingRepository := embeddingrepository.NewEmbeddingRepository(db)
	consumer := consumerservice.NewEmbedNoteConsumerService(
		os.Getenv("RABBITMQ_CONNECTION_STRING"),
		"embed-note-content",
		db,
		embeddingRepository,
	)
	err := consumer.Consume(ctx)
	if err != nil {
		panic(err)
	}
}
