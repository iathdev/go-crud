package port

import (
	"context"
	"learning-go/internal/auth/application/dto"
	"learning-go/internal/auth/domain"

	"github.com/google/uuid"
)

type AuthUseCasePort interface {
	GetMe(ctx context.Context, userID uuid.UUID, prepUser *domain.PrepUser) (*dto.MeResponse, error)
}
