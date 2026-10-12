package ds

import (
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"wells-risk-backend/internal/app/role"
)

type JWTClaims struct {
	jwt.StandardClaims           // все что точно необходимо по RFC
	UserUUID           uuid.UUID `json:"user_uuid"` // наши данные - uuid этого врача в базе данных
	Role               role.Role
}
