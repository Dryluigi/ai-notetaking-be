package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	"ai-notetaking-be/pkg/database"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INoteRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) INoteRepository
	Create(ctx context.Context, noteEntity *noteentity.Note) error
	Update(ctx context.Context, noteEntity *noteentity.Note) error
	UpdateNoteNotebook(ctx context.Context, noteId uuid.UUID, notebookId *uuid.UUID, updatedBy string) error
	GetById(ctx context.Context, id uuid.UUID) (*noteentity.Note, error)
	GetByIds(ctx context.Context, ids []uuid.UUID) ([]*noteentity.Note, error)
	GetByNotebookId(ctx context.Context, notebookId uuid.UUID) ([]*noteentity.Note, error)
}

type noteRepository struct {
	db database.DatabaseQueryer
}

func (n *noteRepository) UsingTx(ctx context.Context, tx database.DatabaseQueryer) INoteRepository {
	return &noteRepository{
		db: tx,
	}
}

func (n *noteRepository) Create(ctx context.Context, noteEntity *noteentity.Note) error {
	_, err := n.db.Exec(
		ctx,
		"INSERT INTO notes (id, title, content, notebook_id, created_at, created_by) VALUES ($1, $2, $3, $4, $5, $6)",
		noteEntity.Id,
		noteEntity.Title,
		noteEntity.Content,
		noteEntity.NotebookId,
		noteEntity.CreatedAt,
		noteEntity.CreatedBy,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *noteRepository) GetById(ctx context.Context, id uuid.UUID) (*noteentity.Note, error) {
	row := n.db.QueryRow(
		ctx,
		`
			SELECT
				n.id,
				n.title,
				n.content,
				n.notebook_id,
				n.created_at,
				n.created_by,
				nb.id,
				nb.name
			FROM
				notes n
			LEFT JOIN notebook nb
				ON nb.id = n.notebook_id
			WHERE n.id = $1
				AND n.is_deleted = false
		`,
		id,
	)
	noteEntity := noteentity.Note{}
	var notebookId *uuid.UUID
	var notebookName *string
	err := row.Scan(
		&noteEntity.Id,
		&noteEntity.Title,
		&noteEntity.Content,
		&noteEntity.NotebookId,
		&noteEntity.CreatedAt,
		&noteEntity.CreatedBy,
		&notebookId,
		&notebookName,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, err
	}

	if notebookId != nil {
		noteEntity.Notebook = &noteentity.Notebook{
			Id:   *notebookId,
			Name: *notebookName,
		}
	}

	return &noteEntity, nil
}

func (n *noteRepository) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*noteentity.Note, error) {
	if len(ids) == 0 {
		return make([]*noteentity.Note, 0), nil
	}

	whereIds := make([]string, 0)
	for _, id := range ids {
		whereIds = append(whereIds, fmt.Sprintf("'%s'", id.String()))
	}
	whereQuery := strings.Join(whereIds, ", ")

	rows, err := n.db.Query(
		ctx,
		fmt.Sprintf(
			"SELECT id, title, content FROM notes WHERE id IN (%s) AND is_deleted = false",
			whereQuery,
		),
	)
	if err != nil {
		return nil, err
	}

	var result []*noteentity.Note = make([]*noteentity.Note, 0)
	for rows.Next() {
		noteEntity := noteentity.Note{}
		err = rows.Scan(
			&noteEntity.Id,
			&noteEntity.Title,
			&noteEntity.Content,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &noteEntity)
	}

	return result, nil
}

func (n *noteRepository) GetByNotebookId(ctx context.Context, notebookId uuid.UUID) ([]*noteentity.Note, error) {
	rows, err := n.db.Query(
		ctx,
		"SELECT id FROM notes WHERE notebook_id = $1 AND is_deleted = false",
		notebookId,
	)
	if err != nil {
		return nil, err
	}

	var result []*noteentity.Note = make([]*noteentity.Note, 0)
	for rows.Next() {
		noteEntity := noteentity.Note{}
		err = rows.Scan(
			&noteEntity.Id,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, &noteEntity)
	}

	return result, nil
}

func (n *noteRepository) Update(ctx context.Context, noteEntity *noteentity.Note) error {
	_, err := n.db.Exec(
		ctx,
		"UPDATE notes SET title = $1, content = $2, notebook_id = $3, updated_at = $4, updated_by = $5 WHERE id = $6",
		noteEntity.Title,
		noteEntity.Content,
		noteEntity.NotebookId,
		noteEntity.UpdatedAt,
		noteEntity.UpdatedBy,
		noteEntity.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (n *noteRepository) UpdateNoteNotebook(ctx context.Context, noteId uuid.UUID, notebookId *uuid.UUID, updatedBy string) error {
	_, err := n.db.Exec(
		ctx,
		"UPDATE notes SET notebook_id = $1, updated_at = $2, updated_by = $3 WHERE id = $4",
		notebookId,
		time.Now(),
		updatedBy,
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
