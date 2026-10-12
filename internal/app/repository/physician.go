package repository

import (
	"github.com/google/uuid"

	"wells-risk-backend/internal/app/ds"
)

// RegisterPhysician регистрирует нового врача.
func (r *Repository) RegisterPhysician(physician *ds.Physician) error {
	if physician.PhysicianUUID == uuid.Nil {
		physician.PhysicianUUID = uuid.New()
	}

	return r.db.Create(physician).Error
}

// PhysicianExists проверяет, занят ли логин.
func (r *Repository) PhysicianExists(login string) (bool, error) {
	var count int64

	err := r.db.Model(&ds.Physician{}).
		Where("login = ?", login).
		Count(&count).Error

	return count > 0, err
}

// GetPhysicianByLogin возвращает врача по логину.
func (r *Repository) GetPhysicianByLogin(login string) (*ds.Physician, error) {
	physician := &ds.Physician{}

	err := r.db.Where("login = ?", login).First(physician).Error
	if err != nil {
		return nil, err
	}

	return physician, nil
}

// GetPhysicianByUUID возвращает врача по uuid из JWT.
func (r *Repository) GetPhysicianByUUID(physicianUUID uuid.UUID) (*ds.Physician, error) {
	physician := &ds.Physician{}

	err := r.db.Where("physician_uuid = ?", physicianUUID).First(physician).Error
	if err != nil {
		return nil, err
	}

	return physician, nil
}
