package output

import (
	"context"
	"learning-go/internal/domain/entity"

	"github.com/google/uuid"
)

type ProductRepositoryPort interface {
	Save(ctx context.Context, product *entity.Product) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error)
	FindAll(ctx context.Context, offset, limit int) ([]*entity.Product, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, product *entity.Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}
