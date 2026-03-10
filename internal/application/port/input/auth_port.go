package input

import (
	"context"
	"learning-go/internal/application/dto"
)

type AuthUseCasePort interface {
	Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error)
	Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error)
}
