package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	embeddingrepository "ai-notetaking-be/internal/repository/embedding"
	noterepository "ai-notetaking-be/internal/repository/note"
	publisherservice "ai-notetaking-be/internal/service/publisher"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type EmbeddingModelRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingModelResponse struct {
	Embedding []float32 `json:"embedding"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type ChatResponse struct {
	Model              string      `json:"model"`
	CreatedAt          string      `json:"created_at"`
	Message            ChatMessage `json:"message"`
	DoneReason         string      `json:"done_reason"`
	Done               bool        `json:"done"`
	TotalDuration      int64       `json:"total_duration"`
	LoadDuration       int64       `json:"load_duration"`
	PromptEvalCount    int         `json:"prompt_eval_count"`
	PromptEvalDuration int64       `json:"prompt_eval_duration"`
	EvalCount          int         `json:"eval_count"`
	EvalDuration       int64       `json:"eval_duration"`
}

type INoteService interface {
	Create(ctx context.Context, request *CreateNoteRequest) (*CreateNoteResponse, error)
	Search(ctx context.Context, request *SearchNoteRequest) ([]*SearchNoteResponse, error)
	Ask(ctx context.Context, request *AskNoteRequest) (*AskNoteResponse, error)
	Update(ctx context.Context, id uuid.UUID, request *UpdateNoteRequest) (*UpdateNoteResponse, error)
	UpdateNoteNotebook(ctx context.Context, id uuid.UUID, request *UpdateNoteNotebookRequest) (*UpdateNoteNotebookResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Show(ctx context.Context, id uuid.UUID) (*ShowNoteResponse, error)
}

type noteService struct {
	noteRepository      noterepository.INoteRepository
	embeddingRepository embeddingrepository.IEmbeddingRepository
	publisherService    publisherservice.IPublisherService

	embeddingModelName      string
	embeddingServiceBaseUrl string

	db *pgxpool.Pool
}

func (ns *noteService) Create(ctx context.Context, request *CreateNoteRequest) (*CreateNoteResponse, error) {
	id := uuid.New()
	noteEntity := noteentity.Note{
		Id:         id,
		Title:      request.Title,
		Content:    request.Content,
		NotebookId: request.NotebookId,
		CreatedAt:  time.Now(),
		CreatedBy:  "System",
	}
	err := ns.noteRepository.Create(ctx, &noteEntity)
	if err != nil {
		return nil, err
	}

	msg := EmbedCreatedNoteMessage{
		NoteId: id,
	}
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	ns.publisherService.Publish(
		ctx,
		msgJson,
	)

	return &CreateNoteResponse{Id: id}, nil
}

func (ns *noteService) Search(ctx context.Context, request *SearchNoteRequest) ([]*SearchNoteResponse, error) {
	req := EmbeddingModelRequest{
		Model:  ns.embeddingModelName,
		Prompt: request.Query,
	}
	reqJson, _ := json.Marshal(req)
	res, err := http.Post(fmt.Sprintf("%s/api/embeddings", ns.embeddingServiceBaseUrl), "application/json", bytes.NewBuffer(reqJson))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var embeddingResponse EmbeddingModelResponse
	err = json.NewDecoder(res.Body).Decode(&embeddingResponse)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ids, err := ns.embeddingRepository.FindMostSimilarNoteIds(
		ctx,
		embeddingResponse.Embedding,
	)
	if err != nil {
		return nil, err
	}

	notes, err := ns.noteRepository.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	var response = make([]*SearchNoteResponse, 0)
	for _, n := range notes {
		response = append(response, &SearchNoteResponse{
			Id:    n.Id,
			Title: n.Title,
		})
		if len(response) == 5 {
			break
		}
	}

	return response, nil
}

func (ns *noteService) Ask(ctx context.Context, request *AskNoteRequest) (*AskNoteResponse, error) {
	req := EmbeddingModelRequest{
		Model:  ns.embeddingModelName,
		Prompt: request.Question,
	}
	reqJson, _ := json.Marshal(req)
	res, err := http.Post(fmt.Sprintf("%s/api/embeddings", ns.embeddingServiceBaseUrl), "application/json", bytes.NewBuffer(reqJson))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var embeddingResponse EmbeddingModelResponse
	err = json.NewDecoder(res.Body).Decode(&embeddingResponse)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	ids, err := ns.embeddingRepository.FindMostSimilarNoteIds(
		ctx,
		embeddingResponse.Embedding,
	)
	if err != nil {
		return nil, err
	}

	notes, err := ns.noteRepository.GetByIds(ctx, ids)
	if err != nil {
		return nil, err
	}

	references := make([]string, 0)
	for i := 0; i < len(notes); i++ {
		references = append(references, fmt.Sprintf("Reference %d", i+1))
		references = append(references, notes[i].Title)
		references = append(references, notes[i].Content)
	}
	referencesString := strings.Join(references, "\n")

	prompt := fmt.Sprintf(`
		Given references and question below. Answer the question directly without asking again with question language

		%s

		Question:
		%s
	
		Your answer: ...
	`, referencesString, request.Question)

	chatRequest := ChatRequest{
		Model: "llama3.2",
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}
	chatRequestJson, _ := json.Marshal(&chatRequest)
	res, err = http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(chatRequestJson))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var answerResponse ChatResponse
	err = json.NewDecoder(res.Body).Decode(&answerResponse)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &AskNoteResponse{
		Answer: answerResponse.Message.Content,
	}, nil
}

func (ns *noteService) Update(ctx context.Context, id uuid.UUID, request *UpdateNoteRequest) (*UpdateNoteResponse, error) {
	n, err := ns.noteRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	updatedBy := "System"
	n.Title = request.Title
	n.Content = request.Content
	n.UpdatedAt = &now
	n.UpdatedBy = &updatedBy

	err = ns.noteRepository.Update(ctx, n)
	if err != nil {
		return nil, err
	}

	msg := EmbedCreatedNoteMessage{
		NoteId:             id,
		DeleteOldEmbedding: true,
	}
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	ns.publisherService.Publish(
		ctx,
		msgJson,
	)

	return &UpdateNoteResponse{Id: id}, nil
}

func (ns *noteService) UpdateNoteNotebook(ctx context.Context, id uuid.UUID, request *UpdateNoteNotebookRequest) (*UpdateNoteNotebookResponse, error) {
	_, err := ns.noteRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}

	err = ns.noteRepository.UpdateNoteNotebook(ctx, id, request.NewNotebookId, "System")
	if err != nil {
		return nil, err
	}

	msg := EmbedCreatedNoteMessage{
		NoteId:             id,
		DeleteOldEmbedding: true,
	}
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	ns.publisherService.Publish(
		ctx,
		msgJson,
	)

	return &UpdateNoteNotebookResponse{Id: id}, nil
}

func (ns *noteService) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := ns.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if e := recover(); e != nil {
			tx.Rollback(ctx)
		}

		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	noteRepo := ns.noteRepository.UsingTx(ctx, tx)
	embedRepo := ns.embeddingRepository.UsingTx(ctx, tx)
	err = noteRepo.DeleteNote(ctx, id, "System")
	if err != nil {
		return err
	}

	err = embedRepo.DeleteByNoteId(ctx, id, "System")
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil
	}

	return nil
}

func (ns *noteService) Show(ctx context.Context, id uuid.UUID) (*ShowNoteResponse, error) {
	note, err := ns.noteRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	res := ShowNoteResponse{
		Id:         note.Id,
		Title:      note.Title,
		Content:    note.Content,
		NotebookId: note.NotebookId,
		CreatedAt:  note.CreatedAt,
		CreatedBy:  note.CreatedBy,
		UpdatedAt:  note.UpdatedAt,
		UpdatedBy:  note.UpdatedBy,
	}

	return &res, nil
}

func NewNoteService(
	noteRepository noterepository.INoteRepository,
	embeddingRepository embeddingrepository.IEmbeddingRepository,
	publisherService publisherservice.IPublisherService,
	embeddingServiceBaseUrl string,
	embeddingModelName string,
	db *pgxpool.Pool,
) INoteService {
	return &noteService{
		noteRepository:          noteRepository,
		publisherService:        publisherService,
		embeddingRepository:     embeddingRepository,
		embeddingModelName:      embeddingModelName,
		embeddingServiceBaseUrl: embeddingServiceBaseUrl,
		db:                      db,
	}
}
