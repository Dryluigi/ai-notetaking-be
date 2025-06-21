package note

import (
	noteservice "ai-notetaking-be/internal/service/note"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type INoteController interface {
	Create(c *fiber.Ctx) error
	Search(c *fiber.Ctx) error
	Ask(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	UpdateNotebook(c *fiber.Ctx) error
}

type noteController struct {
	noteService noteservice.INoteService
}

func (nc *noteController) Create(c *fiber.Ctx) error {
	var request noteservice.CreateNoteRequest
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}

	res, err := nc.noteService.Create(c.UserContext(), &request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (nc *noteController) Search(c *fiber.Ctx) error {
	var request noteservice.SearchNoteRequest
	err := c.QueryParser(&request)
	if err != nil {
		return err
	}

	res, err := nc.noteService.Search(c.UserContext(), &request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (nc *noteController) Ask(c *fiber.Ctx) error {
	var request noteservice.AskNoteRequest
	err := c.QueryParser(&request)
	if err != nil {
		return err
	}

	res, err := nc.noteService.Ask(c.UserContext(), &request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (nc *noteController) Update(c *fiber.Ctx) error {
	id := c.Params("id")
	idUuid, _ := uuid.Parse(id)

	var request noteservice.UpdateNoteRequest
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}

	res, err := nc.noteService.Update(c.UserContext(), idUuid, &request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (nc *noteController) UpdateNotebook(c *fiber.Ctx) error {
	id := c.Params("id")
	idUuid, _ := uuid.Parse(id)

	var request noteservice.UpdateNoteNotebookRequest
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}

	res, err := nc.noteService.UpdateNoteNotebook(c.UserContext(), idUuid, &request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func NewNoteController(noteService noteservice.INoteService) INoteController {
	return &noteController{
		noteService: noteService,
	}
}
