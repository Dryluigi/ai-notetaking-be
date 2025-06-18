package note

import (
	"time"

	"github.com/google/uuid"
)

type Notebook struct {
	Id        uuid.UUID
	Name      string
	ParentId  *uuid.UUID
	CreatedAt time.Time
	CreatedBy string
	UpdatedAt *time.Time
	UpdatedBy *string
	DeletedAt *time.Time
	DeletedBy *string
	IsDeleted bool
}
