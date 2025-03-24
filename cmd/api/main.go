package main

import (
	"context"
	"log"
	"net/http"

	"github.com/stef-zaimis/taxa/internal/db"
	"github.com/stef-zaimis/taxa/internal/api"
)

func main() {
	conn := db.Connect()
	defer conn.Close(context.Background())

	r := api.SetupRouter(conn)

	
	log.Println("Starting server on:8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
