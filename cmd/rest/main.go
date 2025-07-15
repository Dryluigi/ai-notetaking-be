package main

import (
	notecontroller "ai-notetaking-be/internal/controller/note"
	embeddingrepository "ai-notetaking-be/internal/repository/embedding"
	noterepository "ai-notetaking-be/internal/repository/note"
	"ai-notetaking-be/internal/service/consumer"
	noteservice "ai-notetaking-be/internal/service/note"
	publisherservice "ai-notetaking-be/internal/service/publisher"
	"ai-notetaking-be/pkg/database"
	"context"
	"log"
	"os"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app := fiber.New()

	app.Use(cors.New())

	db := database.ConnectDB(os.Getenv("DB_CONNECTION_STRING"))
	// rabbitMqService := publisherservice.NewRabbitMqPublisherService(os.Getenv("RABBITMQ_CONNECTION_STRING"), "embed-note-content")
	embeddingRepository := embeddingrepository.NewEmbeddingRepository(db)

	pubSubLogger := watermill.NewStdLogger(false, false)
	pubsub := gochannel.NewGoChannel(gochannel.Config{}, pubSubLogger)
	publisherService := publisherservice.NewInMemoryPublisherService(pubsub, "embed-note-content")

	noteRepository := noterepository.NewNoteRepository(db)
	notebookRepository := noterepository.NewNotebookRepository(db)
	noteService := noteservice.NewNoteService(
		noteRepository,
		embeddingRepository,
		publisherService,
		os.Getenv("EMBEDDING_SERVER_BASE_URL"),
		os.Getenv("EMBEDDING_MODEL_NAME"),
		db,
	)
	notebookService := noteservice.NewNotebookService(
		notebookRepository,
		noteRepository,
		embeddingRepository,
		publisherService,
		db,
	)
	noteController := notecontroller.NewNoteController(noteService)
	notebookController := notecontroller.NewNotebookController(notebookService)

	notecontroller.AssignNoteRoutes(app, noteController, notebookController)

	cons := consumer.NewInMemoryConsumer(
		pubsub,
		"embed-note-content",
		db,
		os.Getenv("EMBEDDING_SERVER_BASE_URL"),
		os.Getenv("EMBEDDING_MODEL_NAME"),
		embeddingRepository,
		noteRepository,
	)
	err := cons.Consume(context.Background())
	if err != nil {
		log.Panic(err)
	}

	log.Fatal(app.Listen(":3000"))
}
