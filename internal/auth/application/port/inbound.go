package port

import (
	"context"
	"learning-go/internal/auth/application/dto"
	"learning-go/internal/auth/domain"

	"github.com/google/uuid"
)

type AuthUseCasePort interface {
	GetMe(ctx context.Context, userID uuid.UUID, prepUser *domain.PrepUser) (*dto.AuthMeResponse, error)
	UpsertFromPrepUser(ctx context.Context, prepUser *domain.PrepUser) (*domain.User, error)
}
