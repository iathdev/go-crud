package product

import (
	"context"
	"errors"
	"learning-go/internal/application/dto"
	"learning-go/internal/application/port/input"
	"learning-go/internal/application/port/output"
	"learning-go/internal/domain/entity"
	domainerror "learning-go/internal/domain/error"
	"log"
	"math"

	"github.com/google/uuid"
)

type ProductQueryUseCase struct {
	productRepo output.ProductRepositoryPort
}

func NewProductQueryUseCase(productRepo output.ProductRepositoryPort) input.ProductQueryPort {
	return &ProductQueryUseCase{productRepo: productRepo}
}

func (u *ProductQueryUseCase) GetProduct(ctx context.Context, id string) (*dto.ProductResponse, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainerror.ErrInvalidInput
	}

	product, err := u.productRepo.FindByID(ctx, uuidID)
	if err != nil {
		if errors.Is(err, domainerror.ErrNotFound) {
			return nil, domainerror.ErrNotFound
		}
		log.Printf("error finding product by id: %v", err)
		return nil, domainerror.ErrInternal
	}

	return toProductResponse(product), nil
}

func (u *ProductQueryUseCase) ListProducts(ctx context.Context, pagination dto.PaginationRequest) (*dto.PaginatedResponse, error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}
	if pagination.PageSize > 100 {
		pagination.PageSize = 100
	}

	offset := (pagination.Page - 1) * pagination.PageSize

	total, err := u.productRepo.Count(ctx)
	if err != nil {
		log.Printf("error counting products: %v", err)
		return nil, domainerror.ErrInternal
	}

	products, err := u.productRepo.FindAll(ctx, offset, pagination.PageSize)
	if err != nil {
		log.Printf("error listing products: %v", err)
		return nil, domainerror.ErrInternal
	}

	items := make([]*dto.ProductResponse, 0, len(products))
	for _, p := range products {
		items = append(items, toProductResponse(p))
	}

	totalPages := int(math.Ceil(float64(total) / float64(pagination.PageSize)))

	return &dto.PaginatedResponse{
		Items: items,
		Metadata: dto.PaginationMeta{
			Total:      total,
			Page:       pagination.Page,
			PageSize:   pagination.PageSize,
			TotalPages: totalPages,
		},
	}, nil
}

func toProductResponse(p *entity.Product) *dto.ProductResponse {
	return &dto.ProductResponse{
		ID:          p.ID.String(),
		Name:        p.Name,
		Description: p.Description,
		Price:       p.Price,
		CreatedAt:   p.CreatedAt,
	}
}
