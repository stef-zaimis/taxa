package db

import (
	"context"
	"log"
	
	"github.com/jackc/pgx/v5"
)

const dbURL = "postgres://postgres:toor@127.0.0.1:5432/taxa"

// Establish a connection to the postgres db
func Connect() *pgx.Conn {
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return conn
}
