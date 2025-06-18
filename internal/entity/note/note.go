package note

import (
	"time"

	"github.com/google/uuid"
)

type Note struct {
	Id         uuid.UUID
	Title      string
	Content    string
	NotebookId *uuid.UUID
	CreatedAt  time.Time
	CreatedBy  string
	UpdatedAt  *time.Time
	UpdatedBy  *string
	DeletedAt  *time.Time
	DeletedBy  *string
	IsDeleted  bool

	Notebook *Notebook
}
