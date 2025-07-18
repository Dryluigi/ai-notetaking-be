package note

import (
	"time"

	"github.com/google/uuid"
)

type CreateNotebookRequest struct {
	Name     string     `json:"name"`
	ParentId *uuid.UUID `json:"parent_id"`
}

type CreateNotebookResponse struct {
	Id uuid.UUID `json:"id"`
}

type UpdateNotebookRequest struct {
	Name string `json:"name"`
}

type UpdateNotebookResponse struct {
	Id uuid.UUID `json:"id"`
}

type UpdateNotebookParentRequest struct {
	ParentId uuid.UUID `json:"parent_id"`
}

type UpdateNotebookParentResponse struct {
	Id uuid.UUID `json:"id"`
}

type ShowNotebookResponse struct {
	Id        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	ParentId  *uuid.UUID `json:"parent_id"`
	CreatedAt time.Time  `json:"created_at"`
	CreatedBy string     `json:"created_by"`
	UpdatedAt *time.Time `json:"updated_at"`
	UpdatedBy *string    `json:"updated_by"`
}

type GetAllNotebookResponseNote struct {
	Id         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	NotebookId *uuid.UUID `json:"notebook_id"`
	CreatedAt  time.Time  `json:"created_at"`
	CreatedBy  string     `json:"created_by"`
	UpdatedAt  *time.Time `json:"updated_at"`
	UpdatedBy  *string    `json:"updated_by"`
}

type GetAllNotebookResponseNotebook struct {
	Id        uuid.UUID                    `json:"id"`
	Name      string                       `json:"name"`
	ParentId  *uuid.UUID                   `json:"parent_id"`
	CreatedAt time.Time                    `json:"created_at"`
	CreatedBy string                       `json:"created_by"`
	UpdatedAt *time.Time                   `json:"updated_at"`
	UpdatedBy *string                      `json:"updated_by"`
	Notes     []GetAllNotebookResponseNote `json:"books"`
}

type GetAllNotebookResponse struct {
	Notebooks []GetAllNotebookResponseNotebook `json:"notebooks"`
	Notes     []GetAllNotebookResponseNote     `json:"notes"`
}
