package main

import (
	"log"
	"net/http"

	"github.com/stef-zaimis/taxa/internal/db"
	"github.com/stef-zaimis/taxa/internal/api"
)

func main() {
	pool := db.Connect()
	defer pool.Close()

	r := api.SetupRouter(pool)

	
	log.Println("Starting server on:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
