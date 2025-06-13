package note

import "github.com/gofiber/fiber/v2"

func AssignNoteRoutes(app *fiber.App, noteController INoteController) {
	group := app.Group("/api/v1/note")
	group.Post("", noteController.Create)
}
