package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	noterepository "ai-notetaking-be/internal/repository/note"
	publisherservice "ai-notetaking-be/internal/service/publisher"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type INotebookService interface {
	Create(ctx context.Context, request *CreateNotebookRequest) (*CreateNotebookResponse, error)
	Update(ctx context.Context, id uuid.UUID, request *UpdateNotebookRequest) (*UpdateNotebookResponse, error)
	UpdateParent(ctx context.Context, id uuid.UUID, request *UpdateNotebookParentRequest) (*UpdateNotebookParentResponse, error)
}

type notebookService struct {
	noteRepository     noterepository.INoteRepository
	notebookRepository noterepository.INotebookRepository
	rabbitMqService    publisherservice.IRabbitMqPublisherService
}

func (ns *notebookService) Create(ctx context.Context, request *CreateNotebookRequest) (*CreateNotebookResponse, error) {
	id := uuid.New()
	notebookEntity := noteentity.Notebook{
		Id:        id,
		Name:      request.Name,
		ParentId:  request.ParentId,
		CreatedAt: time.Now(),
		CreatedBy: "System",
	}
	err := ns.notebookRepository.Create(ctx, &notebookEntity)
	if err != nil {
		return nil, err
	}

	return &CreateNotebookResponse{Id: id}, nil
}

func (ns *notebookService) Update(ctx context.Context, id uuid.UUID, request *UpdateNotebookRequest) (*UpdateNotebookResponse, error) {
	notebook, err := ns.notebookRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	updatedBy := "System"
	notebook.Name = request.Name
	notebook.UpdatedAt = &now
	notebook.UpdatedBy = &updatedBy

	err = ns.notebookRepository.Update(ctx, notebook)
	if err != nil {
		return nil, err
	}

	notes, err := ns.noteRepository.GetByNotebookId(ctx, notebook.Id)
	if err != nil {
		return nil, err
	}

	for _, note := range notes {
		msg := EmbedCreatedNoteMessage{
			NoteId:             note.Id,
			DeleteOldEmbedding: true,
		}
		msgJson, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		err = ns.rabbitMqService.Publish(
			ctx,
			msgJson,
		)
		if err != nil {
			return nil, err
		}
	}

	return &UpdateNotebookResponse{Id: id}, nil
}

func (ns *notebookService) UpdateParent(ctx context.Context, id uuid.UUID, request *UpdateNotebookParentRequest) (*UpdateNotebookParentResponse, error) {
	notebook, err := ns.notebookRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	updatedBy := "System"
	notebook.ParentId = &request.ParentId
	notebook.UpdatedAt = &now
	notebook.UpdatedBy = &updatedBy

	err = ns.notebookRepository.UpdateParent(ctx, notebook)
	if err != nil {
		return nil, err
	}

	return &UpdateNotebookParentResponse{Id: id}, nil
}

func NewNotebookService(
	notebookRepository noterepository.INotebookRepository,
	noteRepository noterepository.INoteRepository,
	rabbitMqService publisherservice.IRabbitMqPublisherService,
) INotebookService {
	return &notebookService{
		notebookRepository: notebookRepository,
		noteRepository:     noteRepository,
		rabbitMqService:    rabbitMqService,
	}
}
