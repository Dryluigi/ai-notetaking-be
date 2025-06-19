package note

import "github.com/gofiber/fiber/v2"

func AssignNoteRoutes(app *fiber.App, noteController INoteController, notebookController INotebookController) {
	group := app.Group("/api/v1/note")
	group.Get("", noteController.Search)
	group.Get("ask", noteController.Ask)
	group.Post("", noteController.Create)
	group.Put(":id", noteController.Update)

	notebookGroup := app.Group("/api/v1/notebook")
	notebookGroup.Post("", notebookController.Create)
}
