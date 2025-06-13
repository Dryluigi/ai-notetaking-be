package database

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(connectionString string) *pgxpool.Pool {
	var err error
	db, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}

	_, err = db.Exec(context.Background(), `
		CREATE EXTENSION IF NOT EXISTS vector;
	`)
	if err != nil {
		log.Fatalf("Failed to create vector extension: %v", err)
	}

	return db
}
