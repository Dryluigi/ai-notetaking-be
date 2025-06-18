package note

import (
	noteentity "ai-notetaking-be/internal/entity/note"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type INoteRepository interface {
	Create(ctx context.Context, noteEntity *noteentity.Note) error
	GetByIds(ctx context.Context, ids []uuid.UUID) ([]*noteentity.Note, error)
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
