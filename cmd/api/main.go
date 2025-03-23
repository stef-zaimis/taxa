package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
	"github.com/go-chi/cors"


	"github.com/stef-zaimis/taxa/internal/db"
	"github.com/stef-zaimis/taxa/internal/quiz"
)

func main() {
	conn := db.Connect()
	defer conn.Close(context.Background())

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	r.Get("/api/quiz", makeStartQuizHandler(conn))

	log.Println("Starting server on:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func makeStartQuizHandler(conn *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parentRank := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("rank")))
		parentName := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("name")))
		targetRank := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("targetRank")))
		optionCountStr := strings.TrimSpace(r.URL.Query().Get("optionCount"))


		if parentRank == "" || parentName == "" || targetRank == "" {
			log.Println("Missing required query parameters, defaulting to 'kingdom', 'animalia' and 'order'")
			parentRank = "kingdom"
			parentName = "animalia"
			targetRank = "order"
		}

		optionCount := 4 // default
		if optionCountStr != "" {
			if val, err := strconv.Atoi(optionCountStr); err == nil && val > 1 {
				optionCount = val
			}
		}

		question, err := quiz.GenerateQuestion(conn, parentRank, parentName, targetRank, optionCount)
		if err != nil {
			http.Error(w, "Failed to generate question: " + err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Got the full question, sending it over")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(question)
	}
}
