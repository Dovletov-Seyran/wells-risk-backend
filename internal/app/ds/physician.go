package ds

import (
	"github.com/google/uuid"

	"wells-risk-backend/internal/app/role"
)

type Physician struct {
	PhysicianID   int       `gorm:"primaryKey;size:32;column:physician_id" json:"physician_id"`
	PhysicianUUID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex;column:physician_uuid" json:"physician_uuid"`
	Login         string    `gorm:"type:varchar(50);not null;unique;column:login" json:"login"`
	Password      string    `gorm:"type:varchar(100);not null;column:password" json:"-"`
	FullName      string    `gorm:"type:varchar(150);column:full_name" json:"full_name"`
	Role          role.Role `gorm:"not null;default:0;column:role" json:"role"`
}

func (Physician) TableName() string {
	return "physicians"
}
