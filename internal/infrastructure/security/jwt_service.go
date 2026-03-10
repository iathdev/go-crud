package security

import (
	"learning-go/internal/application/port/output"
	"learning-go/internal/domain/entity"
	"learning-go/internal/infrastructure/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey string
}

func NewJWTService(cfg *config.Config) output.TokenServicePort {
	return &JWTService{
		secretKey: cfg.JWTSecret,
	}
}

func (s *JWTService) GenerateToken(user *entity.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"iat":     now.Unix(),
		"exp":     now.Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.secretKey))
}
