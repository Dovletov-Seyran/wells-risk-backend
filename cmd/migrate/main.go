package main

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"wells-risk-backend/internal/app/ds"
	"wells-risk-backend/internal/app/dsn"
	"wells-risk-backend/internal/app/role"
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

	// AutoMigrate не добавляет NOT NULL к уже созданным столбцам, поэтому меняем их явно.
	for _, field := range []string{"ImageKey", "VideoKey", "CreatorID"} {
		if err := db.Migrator().AlterColumn(&ds.WellsCriterion{}, field); err != nil {
			log.Fatalf("ошибка изменения столбца %s: %v", field, err)
		}
	}

	// Врачам, созданным до ЛР4, выдаём uuid для JWT.
	var withoutUUID []ds.Physician
	if err := db.Where("physician_uuid IS NULL").Find(&withoutUUID).Error; err != nil {
		log.Fatalf("ошибка поиска врачей без uuid: %v", err)
	}

	for _, physician := range withoutUUID {
		err := db.Model(&ds.Physician{}).
			Where("physician_id = ?", physician.PhysicianID).
			Update("physician_uuid", uuid.New()).Error
		if err != nil {
			log.Fatalf("ошибка заполнения uuid врача %d: %v", physician.PhysicianID, err)
		}
	}

	if err := db.Migrator().AlterColumn(&ds.Physician{}, "PhysicianUUID"); err != nil {
		log.Fatalf("ошибка изменения столбца physician_uuid: %v", err)
	}

	// Роль модератора переносим из прежнего признака is_moderator, после чего столбец удаляем.
	if db.Migrator().HasColumn(&ds.Physician{}, "is_moderator") {
		err := db.Model(&ds.Physician{}).
			Where("is_moderator = ?", true).
			Update("role", role.Moderator).Error
		if err != nil {
			log.Fatalf("ошибка переноса роли модератора: %v", err)
		}

		if err := db.Migrator().DropColumn(&ds.Physician{}, "is_moderator"); err != nil {
			log.Fatalf("ошибка удаления столбца is_moderator: %v", err)
		}
	}

	var count int64

	if err := db.Model(&ds.Physician{}).Where("physician_id = ?", 1).Count(&count).Error; err != nil {
		log.Fatalf("ошибка проверки врача по умолчанию: %v", err)
	}

	if count == 0 {
		defaultPhysician := ds.Physician{
			PhysicianID:   1,
			PhysicianUUID: uuid.New(),
			Login:         "doctor",
			// sha1 от строки "doctor123"
			Password: "a04fd5431e6c2b3130dd7609794a56b22b4661ec",
			FullName: "Довлетов Сейран Батырович",
			Role:     role.Physician,
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
