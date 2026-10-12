package main

import (
	"context"
	"log"
	"os"

	"wells-risk-backend/internal/pkg/app"
)

func main() {
	ctx := context.Background()

	log.Println("Initializing server")
	application, err := app.New(ctx)
	if err != nil {
		log.Println("cant create application:", err)
		os.Exit(2)
	}

	log.Println("Application start!")
	application.Run()
	log.Println("Application terminated!")
}
