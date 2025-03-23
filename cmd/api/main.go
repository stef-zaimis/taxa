package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/stef-zaimis/taxa/internal/db"
	"github.com/stef-zaimis/taxa/internal/quiz"
)

func main() {
	conn := db.Connect()
	defer conn.Close(context.Background())

	r := chi.NewRouter()
	r.Get("/api/quiz", makeStartQuizHandler(conn))

	log.Println("Starting server on:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func makeStartQuizHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		question, err := quiz.GenerateQuestion(conn, "kingdom", "Animalia", "order", 3)
		if err != nil {
			http.Error(w, "Failed to generate question: " + err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(question)
	}
}
