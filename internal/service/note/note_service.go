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

type INoteService interface {
	Create(ctx context.Context, request *CreateNoteRequest) (*CreateNoteResponse, error)
}

type noteService struct {
	noteRepository  noterepository.INoteRepository
	rabbitMqService publisherservice.IRabbitMqPublisherService
}

func (ns *noteService) Create(ctx context.Context, request *CreateNoteRequest) (*CreateNoteResponse, error) {
	id := uuid.New()
	noteEntity := noteentity.Note{
		Id:        id,
		Title:     request.Title,
		Content:   request.Content,
		CreatedAt: time.Now(),
		CreatedBy: "System",
	}
	err := ns.noteRepository.Create(ctx, &noteEntity)
	if err != nil {
		return nil, err
	}

	msg := EmbedCreatedNoteMessage{
		NoteId:  id,
		Title:   noteEntity.Title,
		Content: noteEntity.Content,
	}
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	ns.rabbitMqService.Publish(
		ctx,
		msgJson,
	)

	return &CreateNoteResponse{Id: id}, nil
}

func NewNoteService(noteRepository noterepository.INoteRepository, rabbitMqService publisherservice.IRabbitMqPublisherService) INoteService {
	return &noteService{
		noteRepository:  noteRepository,
		rabbitMqService: rabbitMqService,
	}
}
