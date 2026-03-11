package service

import (
	"crypto/rand"
	"encoding/base64"
	"learning-go/internal/auth/application/port"
	"learning-go/internal/auth/domain"
	"learning-go/internal/infrastructure/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secretKey    string
	accessExpiry time.Duration
}

func NewJWTService(cfg *config.Config) port.TokenServicePort {
	return &JWTService{
		secretKey:    cfg.JWTSecret,
		accessExpiry: time.Duration(cfg.GetAccessTokenExpiry()) * time.Minute,
	}
}

func (service *JWTService) GenerateToken(user *domain.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"user_id": user.ID.String(),
		"email":   user.Email,
		"iat":     now.Unix(),
		"exp":     now.Add(service.accessExpiry).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(service.secretKey))
}

func (service *JWTService) GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
