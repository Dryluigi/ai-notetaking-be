package main

import (
	notecontroller "ai-notetaking-be/internal/controller/note"
	embeddingrepository "ai-notetaking-be/internal/repository/embedding"
	noterepository "ai-notetaking-be/internal/repository/note"
	noteservice "ai-notetaking-be/internal/service/note"
	publisherservice "ai-notetaking-be/internal/service/publisher"
	"ai-notetaking-be/pkg/database"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	app := fiber.New()

	db := database.ConnectDB(os.Getenv("DB_CONNECTION_STRING"))
	rabbitMqService := publisherservice.NewRabbitMqPublisherService(os.Getenv("RABBITMQ_CONNECTION_STRING"), "embed-note-content")
	embeddingRepository := embeddingrepository.NewEmbeddingRepository(db)

	noteRepository := noterepository.NewNoteRepository(db)
	notebookRepository := noterepository.NewNotebookRepository(db)
	noteService := noteservice.NewNoteService(
		noteRepository,
		embeddingRepository,
		rabbitMqService,
		os.Getenv("EMBEDDING_SERVER_BASE_URL"),
		os.Getenv("EMBEDDING_MODEL_NAME"),
	)
	notebookService := noteservice.NewNotebookService(
		notebookRepository,
		noteRepository,
		rabbitMqService,
	)
	noteController := notecontroller.NewNoteController(noteService)
	notebookController := notecontroller.NewNotebookController(notebookService)

	notecontroller.AssignNoteRoutes(app, noteController, notebookController)

	log.Fatal(app.Listen(":3000"))
}
