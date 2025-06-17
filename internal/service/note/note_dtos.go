package note

import "github.com/google/uuid"

type CreateNoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreateNoteResponse struct {
	Id uuid.UUID `json:"id"`
}

type SearchNoteRequest struct {
	Query string `query:"query"`
}

type SearchNoteResponse struct {
	Id    uuid.UUID `json:"id"`
	Title string    `json:"title"`
}

type EmbedCreatedNoteMessage struct {
	NoteId  uuid.UUID `json:"note_id"`
	Title   string    `json:"title"`
	Content string    `json:"content"`
}
