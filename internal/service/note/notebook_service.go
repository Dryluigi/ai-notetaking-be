package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	noterepository "ai-notetaking-be/internal/repository/note"
	publisherservice "ai-notetaking-be/internal/service/publisher"
	"context"
	"time"

	"github.com/google/uuid"
)

type INotebookService interface {
	Create(ctx context.Context, request *CreateNotebookRequest) (*CreateNotebookResponse, error)
}

type notebookService struct {
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

func NewNotebookService(
	notebookRepository noterepository.INotebookRepository,
	rabbitMqService publisherservice.IRabbitMqPublisherService,
) INotebookService {
	return &notebookService{
		notebookRepository: notebookRepository,
		rabbitMqService:    rabbitMqService,
	}
}
