package port

import (
	"context"
	"learning-go/internal/auth/application/dto"

	"github.com/google/uuid"
)

// Input ports (driving) — used by handlers to call usecases

type AuthUseCasePort interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.TokenResponse, error)
	Logout(ctx context.Context, req dto.LogoutRequest) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error)
}
