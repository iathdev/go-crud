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

	"github.com/google/uuid"
)

type ProductCommandUseCase struct {
	productRepo output.ProductRepositoryPort
}

func NewProductCommandUseCase(productRepo output.ProductRepositoryPort) input.ProductCommandPort {
	return &ProductCommandUseCase{productRepo: productRepo}
}

func (u *ProductCommandUseCase) CreateProduct(ctx context.Context, req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product, err := entity.NewProduct(req.Name, req.Description, req.Price)
	if err != nil {
		return nil, mapProductEntityError(err)
	}

	if err := u.productRepo.Save(ctx, product); err != nil {
		log.Printf("error saving product: %v", err)
		return nil, domainerror.ErrInternal
	}

	return toProductResponse(product), nil
}

func (u *ProductCommandUseCase) UpdateProduct(ctx context.Context, id string, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return nil, domainerror.ErrInvalidInput
	}

	product, err := u.productRepo.FindByID(ctx, uuidID)
	if err != nil {
		if errors.Is(err, domainerror.ErrNotFound) {
			return nil, domainerror.ErrNotFound
		}
		log.Printf("error finding product: %v", err)
		return nil, domainerror.ErrInternal
	}

	if err := product.Update(req.Name, req.Description, req.Price); err != nil {
		return nil, mapProductEntityError(err)
	}

	if err := u.productRepo.Update(ctx, product); err != nil {
		log.Printf("error updating product: %v", err)
		return nil, domainerror.ErrInternal
	}

	return toProductResponse(product), nil
}

func (u *ProductCommandUseCase) DeleteProduct(ctx context.Context, id string) error {
	uuidID, err := uuid.Parse(id)
	if err != nil {
		return domainerror.ErrInvalidInput
	}

	if _, err := u.productRepo.FindByID(ctx, uuidID); err != nil {
		if errors.Is(err, domainerror.ErrNotFound) {
			return domainerror.ErrNotFound
		}
		log.Printf("error finding product: %v", err)
		return domainerror.ErrInternal
	}

	if err := u.productRepo.Delete(ctx, uuidID); err != nil {
		log.Printf("error deleting product: %v", err)
		return domainerror.ErrInternal
	}

	return nil
}

func mapProductEntityError(err error) error {
	switch {
	case errors.Is(err, entity.ErrProductNameRequired):
		return domainerror.ErrInvalidInput
	case errors.Is(err, entity.ErrProductPriceMustBeGreaterThan0):
		return domainerror.ErrInvalidInput
	default:
		return domainerror.ErrInternal
	}
}
