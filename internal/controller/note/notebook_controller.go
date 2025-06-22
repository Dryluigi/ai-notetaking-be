package note

import (
	noteservice "ai-notetaking-be/internal/service/note"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type INotebookController interface {
	Create(c *fiber.Ctx) error
	Update(c *fiber.Ctx) error
	UpdateParent(c *fiber.Ctx) error
	Delete(c *fiber.Ctx) error
	Show(c *fiber.Ctx) error
}

type notebookController struct {
	notebookService noteservice.INotebookService
}

func (nc *notebookController) Create(c *fiber.Ctx) error {
	var request noteservice.CreateNotebookRequest
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}

	res, err := nc.notebookService.Create(c.UserContext(), &request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(res)
}

func (nc *notebookController) Update(c *fiber.Ctx) error {
	var request noteservice.UpdateNotebookRequest
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}
	id := c.Params("id")
	idUuid := uuid.MustParse(id)

	res, err := nc.notebookService.Update(c.UserContext(), idUuid, &request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (nc *notebookController) UpdateParent(c *fiber.Ctx) error {
	var request noteservice.UpdateNotebookParentRequest
	err := c.BodyParser(&request)
	if err != nil {
		return err
	}
	id := c.Params("id")
	idUuid := uuid.MustParse(id)

	res, err := nc.notebookService.UpdateParent(c.UserContext(), idUuid, &request)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func (nc *notebookController) Delete(c *fiber.Ctx) error {
	id := c.Params("id")
	idUuid := uuid.MustParse(id)

	err := nc.notebookService.Delete(c.UserContext(), idUuid)
	if err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

func (nc *notebookController) Show(c *fiber.Ctx) error {
	id := c.Params("id")
	idUuid := uuid.MustParse(id)

	res, err := nc.notebookService.Show(c.UserContext(), idUuid)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusOK).JSON(res)
}

func NewNotebookController(notebookService noteservice.INotebookService) INotebookController {
	return &notebookController{
		notebookService: notebookService,
	}
}
