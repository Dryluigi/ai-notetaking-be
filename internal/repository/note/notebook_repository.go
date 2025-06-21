package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INotebookRepository interface {
	GetById(ctx context.Context, id uuid.UUID) (*noteentity.Notebook, error)
	Create(ctx context.Context, notebookEntity *noteentity.Notebook) error
	Update(ctx context.Context, notebookEntity *noteentity.Notebook) error
}

type notebookRepository struct {
	db *pgxpool.Pool
}

func (n *notebookRepository) GetById(ctx context.Context, id uuid.UUID) (*noteentity.Notebook, error) {
	var notebook noteentity.Notebook
	row := n.db.QueryRow(
		ctx,
		"SELECT id, name, parent_id, created_at, created_by, updated_at, updated_by FROM notebook WHERE id = $1 AND is_deleted = false",
		id,
	)
	err := row.Scan(
		&notebook.Id,
		&notebook.Name,
		&notebook.ParentId,
		&notebook.CreatedAt,
		&notebook.CreatedBy,
		&notebook.UpdatedAt,
		&notebook.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &notebook, nil
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

func (n *notebookRepository) Update(ctx context.Context, notebookEntity *noteentity.Notebook) error {
	_, err := n.db.Exec(
		ctx,
		"UPDATE notebook SET name = $1, updated_at = $2, updated_by = $3 WHERE id = $4",
		notebookEntity.Name,
		notebookEntity.UpdatedAt,
		notebookEntity.UpdatedBy,
		notebookEntity.Id,
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
