package main

import (
	"log"

	"github.com/joho/godotenv"

	"wells-risk-backend/internal/api"
)

func main() {
	log.Println("Application start!")

	if err := godotenv.Load(); err != nil {
		log.Println("файл .env не найден, читаю переменные окружения")
	}

	api.StartServer()

	log.Println("Application terminated")
}
