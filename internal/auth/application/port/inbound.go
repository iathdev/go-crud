package port

import (
	"context"
	"learning-go/internal/auth/application/dto"

	"github.com/google/uuid"
)

// Input ports (driving) — used by handlers to call usecases

type AuthUseCasePort interface {
	GetMe(ctx context.Context, userID uuid.UUID, isFirstLogin bool) (*dto.MeResponse, error)
}
