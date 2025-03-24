package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stef-zaimis/taxa/internal/quiz"
	"github.com/stef-zaimis/taxa/internal/search"
)

func MakeStartQuizHandler(pool *pgxpool.Pool) http.HandlerFunc {
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

		question, err := quiz.GenerateQuestion(pool, parentRank, parentName, targetRank, optionCount)
		if err != nil {
			http.Error(w, "Failed to generate question: " + err.Error(), http.StatusInternalServerError)
			return
		}

		log.Println("Got the full question, sending it over")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(question)
	}
}

func MakeSearchHandler(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := strings.TrimSpace(r.URL.Query().Get("q"))
		log.Println("Search query received:", query)

		if query == "" {
			http.Error(w, "Missing search query", http.StatusBadRequest)
			return
		}

		results, err := search.SearchTaxa(pool, query, 10)
		if err != nil {
			log.Println("Search error:", err)
			http.Error(w, "Search failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}
