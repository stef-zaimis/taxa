package db

import (
	"context"
	"log"
	
	"github.com/jackc/pgx/v5/pgxpool"
)

// Establish a connection to the postgres db
func Connect(dbURL string) *pgxpool.Pool {
	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return pool 
}
