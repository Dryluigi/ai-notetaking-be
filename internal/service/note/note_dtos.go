package note

import (
	"time"

	"github.com/google/uuid"
)

type CreateNoteRequest struct {
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	NotebookId *uuid.UUID `json:"notebook_id"`
}

type CreateNoteResponse struct {
	Id uuid.UUID `json:"id"`
}

type UpdateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type UpdateNoteResponse struct {
	Id uuid.UUID `json:"id"`
}

type UpdateNoteNotebookRequest struct {
	NewNotebookId *uuid.UUID `json:"new_notebook_id"`
}

type UpdateNoteNotebookResponse struct {
	Id uuid.UUID `json:"id"`
}

type SearchNoteRequest struct {
	Query string `query:"query"`
}

type SearchNoteResponse struct {
	Id    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}

type AskNoteRequest struct {
	Question string `query:"question"`
}

type AskNoteResponse struct {
	Answer string `json:"answer"`
}

type ShowNoteResponse struct {
	Id         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	Content    string     `json:"content"`
	NotebookId *uuid.UUID `json:"notebook_id"`
	CreatedAt  time.Time  `json:"created_at"`
	CreatedBy  string     `json:"created_by"`
	UpdatedAt  *time.Time `json:"updated_at"`
	UpdatedBy  *string    `json:"updated_by"`
}

type EmbedCreatedNoteMessage struct {
	NoteId             uuid.UUID `json:"note_id"`
	DeleteOldEmbedding bool      `json:"delete_old_embedding"`
}
