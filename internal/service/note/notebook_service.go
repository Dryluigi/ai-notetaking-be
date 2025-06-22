package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	embeddingrepository "ai-notetaking-be/internal/repository/embedding"
	noterepository "ai-notetaking-be/internal/repository/note"
	publisherservice "ai-notetaking-be/internal/service/publisher"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INotebookService interface {
	Create(ctx context.Context, request *CreateNotebookRequest) (*CreateNotebookResponse, error)
	Update(ctx context.Context, id uuid.UUID, request *UpdateNotebookRequest) (*UpdateNotebookResponse, error)
	UpdateParent(ctx context.Context, id uuid.UUID, request *UpdateNotebookParentRequest) (*UpdateNotebookParentResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Show(ctx context.Context, id uuid.UUID) (*ShowNotebookResponse, error)
}

type notebookService struct {
	noteRepository      noterepository.INoteRepository
	notebookRepository  noterepository.INotebookRepository
	embeddingRepository embeddingrepository.IEmbeddingRepository
	rabbitMqService     publisherservice.IRabbitMqPublisherService

	db *pgxpool.Pool
}

func (ns *notebookService) Create(ctx context.Context, request *CreateNotebookRequest) (*CreateNotebookResponse, error) {
	id := uuid.New()
	notebookEntity := noteentity.Notebook{
		Id:        id,
		Name:      request.Name,
		ParentId:  request.ParentId,
		CreatedAt: time.Now(),
		CreatedBy: "System",
	}
	err := ns.notebookRepository.Create(ctx, &notebookEntity)
	if err != nil {
		return nil, err
	}

	return &CreateNotebookResponse{Id: id}, nil
}

func (ns *notebookService) Update(ctx context.Context, id uuid.UUID, request *UpdateNotebookRequest) (*UpdateNotebookResponse, error) {
	notebook, err := ns.notebookRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	updatedBy := "System"
	notebook.Name = request.Name
	notebook.UpdatedAt = &now
	notebook.UpdatedBy = &updatedBy

	err = ns.notebookRepository.Update(ctx, notebook)
	if err != nil {
		return nil, err
	}

	notes, err := ns.noteRepository.GetByNotebookId(ctx, notebook.Id)
	if err != nil {
		return nil, err
	}

	for _, note := range notes {
		msg := EmbedCreatedNoteMessage{
			NoteId:             note.Id,
			DeleteOldEmbedding: true,
		}
		msgJson, err := json.Marshal(msg)
		if err != nil {
			return nil, err
		}
		err = ns.rabbitMqService.Publish(
			ctx,
			msgJson,
		)
		if err != nil {
			return nil, err
		}
	}

	return &UpdateNotebookResponse{Id: id}, nil
}

func (ns *notebookService) UpdateParent(ctx context.Context, id uuid.UUID, request *UpdateNotebookParentRequest) (*UpdateNotebookParentResponse, error) {
	notebook, err := ns.notebookRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	updatedBy := "System"
	notebook.ParentId = &request.ParentId
	notebook.UpdatedAt = &now
	notebook.UpdatedBy = &updatedBy

	err = ns.notebookRepository.UpdateParent(ctx, notebook)
	if err != nil {
		return nil, err
	}

	return &UpdateNotebookParentResponse{Id: id}, nil
}

func (ns *notebookService) Delete(ctx context.Context, id uuid.UUID) error {
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
	notebookRepo := ns.notebookRepository.UsingTx(ctx, tx)
	embedRepo := ns.embeddingRepository.UsingTx(ctx, tx)

	deletedBy := "System"
	err = notebookRepo.Delete(ctx, id, deletedBy)
	if err != nil {
		return err
	}

	err = noteRepo.DeleteByNotebookId(ctx, id, deletedBy)
	if err != nil {
		return err
	}

	err = embedRepo.DeleteByNotebookId(ctx, id, deletedBy)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil
	}

	return nil
}

func (ns *notebookService) Show(ctx context.Context, id uuid.UUID) (*ShowNotebookResponse, error) {
	notebook, err := ns.notebookRepository.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	res := ShowNotebookResponse{
		Id:        notebook.Id,
		Name:      notebook.Name,
		ParentId:  notebook.ParentId,
		CreatedAt: notebook.CreatedAt,
		CreatedBy: notebook.CreatedBy,
		UpdatedAt: notebook.UpdatedAt,
		UpdatedBy: notebook.UpdatedBy,
	}

	return &res, nil
}

func NewNotebookService(
	notebookRepository noterepository.INotebookRepository,
	noteRepository noterepository.INoteRepository,
	embeddingRepository embeddingrepository.IEmbeddingRepository,
	rabbitMqService publisherservice.IRabbitMqPublisherService,
	db *pgxpool.Pool,
) INotebookService {
	return &notebookService{
		notebookRepository:  notebookRepository,
		noteRepository:      noteRepository,
		embeddingRepository: embeddingRepository,
		rabbitMqService:     rabbitMqService,
		db:                  db,
	}
}
