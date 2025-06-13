package main

import (
	notecontroller "ai-notetaking-be/internal/controller/note"
	noterepository "ai-notetaking-be/internal/repository/note"
	noteservice "ai-notetaking-be/internal/service/note"
	"ai-notetaking-be/internal/service/rabbitmq"
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
	rabbitMqService := rabbitmq.NewRabbitMqService(os.Getenv("RABBITMQ_CONNECTION_STRING"), "embed-note-content")
	noteRepository := noterepository.NewNoteRepository(db)
	noteService := noteservice.NewNoteService(noteRepository, rabbitMqService)
	noteController := notecontroller.NewNoteController(noteService)

	notecontroller.AssignNoteRoutes(app, noteController)

	log.Fatal(app.Listen(":3000"))
}
