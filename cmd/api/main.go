package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func main() {
	r := chi.NewRouter()
	r.Get("/api/quiz", handleStartQuiz)
	r.Get("/api/search", handleSearchTaxon)

	log.Pringln("Starting server on:8080")
	http.ListenAndServe(":8080", r)
