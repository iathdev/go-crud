package port

import (
	"context"
	"learning-go/internal/auth/domain"

	"github.com/google/uuid"
)

// Output ports (driven) — implemented by adapters (repositories, services)

type PrepUserServicePort interface {
	ValidateToken(ctx context.Context, token string) (*domain.PrepUser, error)
}

type UserRepositoryPort interface {
	Upsert(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	FindByPrepUserID(ctx context.Context, prepUserID int64) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
}