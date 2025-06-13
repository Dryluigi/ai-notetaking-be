package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	noterepository "ai-notetaking-be/internal/repository/note"
	"context"
	"time"

	"github.com/google/uuid"
)

type INoteService interface {
	Create(ctx context.Context, request *CreateNoteRequest) (*CreateNoteResponse, error)
}

type noteService struct {
	noteRepository noterepository.INoteRepository
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

	return &CreateNoteResponse{Id: id}, nil
}

func NewNoteService(noteRepository noterepository.INoteRepository) INoteService {
	return &noteService{
		noteRepository: noteRepository,
	}
}
