package main

import (
	"log"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/dsn"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("файл .env не найден, читаю переменные окружения")
	}

	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		log.Fatalf("не удалось подключиться к базе: %v", err)
	}

	err = db.AutoMigrate(
		&ds.Physician{},
		&ds.WellsCriterion{},
		&ds.CriterionLike{},
	)
	if err != nil {
		log.Fatalf("ошибка миграции: %v", err)
	}

	log.Println("Миграция выполнена успешно")
}
