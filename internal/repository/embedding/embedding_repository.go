package embedding

import (
	embeddingentity "ai-notetaking-be/internal/entity/embedding"
	"ai-notetaking-be/pkg/database"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	pgvector "github.com/pgvector/pgvector-go"
)

type IEmbeddingRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) IEmbeddingRepository
	CreateNoteEmbedding(ctx context.Context, noteEmbedding *embeddingentity.NoteEmbedding) error
	FindMostSimilarNoteIds(ctx context.Context, embeddingValue []float32) ([]uuid.UUID, error)
}

type embeddingRepository struct {
	db database.DatabaseQueryer
}

func (n *embeddingRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) IEmbeddingRepository {
	return &embeddingRepository{
		db: tx,
	}
}

func (n *embeddingRepository) CreateNoteEmbedding(ctx context.Context, noteEmbedding *embeddingentity.NoteEmbedding) error {
	_, err := n.db.Exec(
		ctx,
		"INSERT INTO embedding_notes (id, original_text, embedding, note_id, created_at, created_by) VALUES ($1, $2, $3, $4, $5, $6)",
		noteEmbedding.Id,
		noteEmbedding.OriginalText,
		pgvector.NewVector(noteEmbedding.Embedding),
		noteEmbedding.NoteId,
		noteEmbedding.CreatedAt,
		noteEmbedding.CreatedBy,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *embeddingRepository) FindMostSimilarNoteIds(ctx context.Context, embeddingValue []float32) ([]uuid.UUID, error) {
	rows, err := n.db.Query(
		ctx,
		"SELECT note_id, original_text, embedding <-> $1 AS similarity FROM embedding_notes ORDER BY similarity LIMIT 10",
		pgvector.NewVector(embeddingValue),
	)
	if err != nil {
		return nil, err
	}

	result := make([]uuid.UUID, 10)
	i := 9
	for rows.Next() {
		var id uuid.UUID
		var distance float32

		err = rows.Scan(&id, &distance)
		if err != nil {
			return nil, err
		}

		result[i] = id
		i--
	}

	return result, nil
}

func NewEmbeddingRepository(db *pgxpool.Pool) IEmbeddingRepository {
	return &embeddingRepository{
		db: db,
	}
}
