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
	"time"

	"github.com/google/uuid"
)

type EmbeddingModelRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type EmbeddingModelResponse struct {
	Embedding []float32 `json:"embedding"`
}

type INoteService interface {
	Create(ctx context.Context, request *CreateNoteRequest) (*CreateNoteResponse, error)
	Search(ctx context.Context, request *SearchNoteRequest) ([]*SearchNoteResponse, error)
}

type noteService struct {
	noteRepository      noterepository.INoteRepository
	embeddingRepository embeddingrepository.IEmbeddingRepository
	rabbitMqService     publisherservice.IRabbitMqPublisherService

	embeddingModelName      string
	embeddingServiceBaseUrl string
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

	msg := EmbedCreatedNoteMessage{
		NoteId:  id,
		Title:   noteEntity.Title,
		Content: noteEntity.Content,
	}
	msgJson, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	ns.rabbitMqService.Publish(
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

func NewNoteService(
	noteRepository noterepository.INoteRepository,
	embeddingRepository embeddingrepository.IEmbeddingRepository,
	rabbitMqService publisherservice.IRabbitMqPublisherService,
	embeddingServiceBaseUrl string,
	embeddingModelName string,
) INoteService {
	return &noteService{
		noteRepository:          noteRepository,
		rabbitMqService:         rabbitMqService,
		embeddingRepository:     embeddingRepository,
		embeddingModelName:      embeddingModelName,
		embeddingServiceBaseUrl: embeddingServiceBaseUrl,
	}
}
