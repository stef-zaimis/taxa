package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"

	"github.com/stef-zaimis/taxa/internal/db"
	"github.com/stef-zaimis/taxa/internal/api"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found (fallback to system env)")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	pool := db.Connect(dbURL)
	defer pool.Close()

	r := api.SetupRouter(pool)

	
	log.Println("Starting server on:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
