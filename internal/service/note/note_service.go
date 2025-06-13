package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	noterepository "ai-notetaking-be/internal/repository/note"
	rabbitqservice "ai-notetaking-be/internal/service/rabbitmq"
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
	rabbitMqService rabbitqservice.IRabbitMqService
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

func NewNoteService(noteRepository noterepository.INoteRepository, rabbitMqService rabbitqservice.IRabbitMqService) INoteService {
	return &noteService{
		noteRepository:  noteRepository,
		rabbitMqService: rabbitMqService,
	}
}
