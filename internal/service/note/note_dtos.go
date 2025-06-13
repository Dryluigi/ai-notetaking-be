package note

import "github.com/google/uuid"

type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateNoteResponse struct {
	Id uuid.UUID `json:"id"`
}
