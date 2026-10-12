package main

import (
	"context"
	"log"
	"os"

	"wells-risk-backend/internal/pkg/app"
)

// @title Критерии шкалы Уэллса
// @version 1.0
// @description Веб-сервис справочника критериев шкалы Уэллса для оценки риска тромбоза глубоких вен и ТЭЛА

// @contact.name Довлетов Сейран
// @contact.email dovletovsejran7@gmail.com

// @license.name AS IS (NO WARRANTY)

// @host localhost:8080
// @schemes http
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Токен из ответа аутентификации в виде: Bearer <access_token>

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
