package note

import "github.com/google/uuid"

type CreateNotebookRequest struct {
	Name     string     `json:"name"`
	ParentId *uuid.UUID `json:"parent_id"`
}

type CreateNotebookResponse struct {
	Id uuid.UUID `json:"id"`
}
