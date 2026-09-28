package main

import (
	"fmt"
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

	var count int64

	if err := db.Model(&ds.Physician{}).Where("physician_id = ?", 1).Count(&count).Error; err != nil {
		log.Fatalf("ошибка проверки врача по умолчанию: %v", err)
	}

	if count == 0 {
		defaultPhysician := ds.Physician{
			PhysicianID: 1,
			Login:       "doctor",
			// sha256 от строки "doctor123"
			Password:    "f348d5628621f3d8f59c8cabda0f8eb0aa7e0514a90be7571020b1336f26c113",
			FullName:    "Довлетов Сейран Батырович",
			IsModerator: false,
		}

		if err := db.Create(&defaultPhysician).Error; err != nil {
			log.Fatalf("ошибка создания врача по умолчанию: %v", err)
		}
	}

	sequences := []struct {
		table  string
		column string
	}{
		{"physicians", "physician_id"},
		{"wells_criteria", "criterion_id"},
		{"criterion_likes", "like_id"},
	}

	for _, sequence := range sequences {
		query := fmt.Sprintf(
			"SELECT setval(pg_get_serial_sequence('%s', '%s'), COALESCE((SELECT MAX(%s) FROM %s), 0) + 1, false)",
			sequence.table, sequence.column, sequence.column, sequence.table,
		)

		if err := db.Exec(query).Error; err != nil {
			log.Fatalf("ошибка сброса последовательности %s: %v", sequence.table, err)
		}
	}

	log.Println("Миграция выполнена успешно")
}
