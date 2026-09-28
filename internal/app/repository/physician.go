package repository

import (
	"crypto/sha256"
	"encoding/hex"

	"wells-risk-backend/internal/app/ds"
)

// hashPassword — хеш пароля врача.
func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))

	return hex.EncodeToString(sum[:])
}

// CreatePhysician регистрирует нового врача.
func (r *Repository) CreatePhysician(physician *ds.Physician) error {
	physician.Password = hashPassword(physician.Password)

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

// GetPhysician возвращает врача по ид.
func (r *Repository) GetPhysician(physicianID int) (ds.Physician, error) {
	var physician ds.Physician

	err := r.db.
		Where("physician_id = ?", physicianID).
		First(&physician).Error
	if err != nil {
		return ds.Physician{}, err
	}

	return physician, nil
}
