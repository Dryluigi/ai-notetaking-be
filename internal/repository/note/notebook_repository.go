package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type INotebookRepository interface {
	Create(ctx context.Context, notebookEntity *noteentity.Notebook) error
}

type notebookRepository struct {
	db *pgxpool.Pool
}

func (n *notebookRepository) Create(ctx context.Context, notebookEntity *noteentity.Notebook) error {
	_, err := n.db.Exec(
		ctx,
		"INSERT INTO notebook (id, name, parent_id, created_at, created_by) VALUES ($1, $2, $3, $4, $5)",
		notebookEntity.Id,
		notebookEntity.Name,
		notebookEntity.ParentId,
		notebookEntity.CreatedAt,
		notebookEntity.CreatedBy,
	)
	if err != nil {
		return err
	}

	return nil
}

func NewNotebookRepository(db *pgxpool.Pool) INotebookRepository {
	return &notebookRepository{
		db: db,
	}
}
