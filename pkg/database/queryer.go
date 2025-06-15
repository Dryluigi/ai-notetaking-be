package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseQueryer interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
}
