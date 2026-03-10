package output

import (
	"context"
	"learning-go/internal/domain/entity"

	"github.com/google/uuid"
)

type UserRepositoryPort interface {
	Save(ctx context.Context, user *entity.User) error
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id uuid.UUID) (*entity.User, error)
}
