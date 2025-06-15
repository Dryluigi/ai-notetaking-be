package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
)

type INoteRepository interface {
	Create(ctx context.Context, noteEntity *noteentity.Note) error
	UpdateNoteEmbeddingValue(ctx context.Context, noteId uuid.UUID, embeddingValue []float32) error
}

type noteRepository struct {
	db *pgxpool.Pool
}

func (n *noteRepository) Create(ctx context.Context, noteEntity *noteentity.Note) error {
	_, err := n.db.Exec(
		ctx,
		"INSERT INTO notes (id, title, content, created_at, created_by) VALUES ($1, $2, $3, $4, $5)",
		noteEntity.Id,
		noteEntity.Title,
		noteEntity.Content,
		noteEntity.CreatedAt,
		noteEntity.CreatedBy,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *noteRepository) UpdateNoteEmbeddingValue(ctx context.Context, noteId uuid.UUID, embeddingValue []float32) error {
	_, err := n.db.Exec(
		ctx,
		"UPDATE notes SET embedding = $1 WHERE id = $2",
		pgvector.NewVector(embeddingValue),
		noteId,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewNoteRepository(db *pgxpool.Pool) INoteRepository {
	return &noteRepository{
		db: db,
	}
}
