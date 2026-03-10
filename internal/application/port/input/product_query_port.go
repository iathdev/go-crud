package input

import (
	"context"
	"learning-go/internal/application/dto"
)

type ProductQueryPort interface {
	GetProduct(ctx context.Context, id string) (*dto.ProductResponse, error)
	ListProducts(ctx context.Context, pagination dto.PaginationRequest) (*dto.PaginatedResponse, error)
}
