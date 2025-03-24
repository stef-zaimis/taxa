package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5"
)

func SetupRouter(conn *pgx.Conn) *chi.Mux {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:5173"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge: 300,
	}))

	r.Get("/api/quiz", MakeStartQuizHandler(conn))
	r.Get("/api/search", MakeSearchHandler(conn))

	return r
}
