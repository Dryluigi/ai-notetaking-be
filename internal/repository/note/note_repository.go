package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	"ai-notetaking-be/pkg/database"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INoteRepository interface {
	UsingTx(ctx context.Context, tx database.DatabaseQueryer) INoteRepository
	Create(ctx context.Context, noteEntity *noteentity.Note) error
	GetById(ctx context.Context, id uuid.UUID) (*noteentity.Note, error)
	GetByIds(ctx context.Context, ids []uuid.UUID) ([]*noteentity.Note, error)
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
			"SELECT id, title, content FROM notes WHERE id IN (%s)",
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

func NewNoteRepository(db *pgxpool.Pool) INoteRepository {
	return &noteRepository{
		db: db,
	}
}
