package embedding

import (
	"time"

	"github.com/google/uuid"
)

type NoteEmbedding struct {
	Id           uuid.UUID
	OriginalText string
	NoteId       uuid.UUID
	Embedding    []float32
	CreatedAt    time.Time
	CreatedBy    string
	UpdatedAt    *time.Time
	UpdatedBy    *string
	DeletedAt    *time.Time
	DeletedBy    *string
	IsDeleted    bool
}
