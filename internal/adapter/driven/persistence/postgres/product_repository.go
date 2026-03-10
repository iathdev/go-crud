package postgres

import (
	"context"
	"errors"
	"learning-go/internal/adapter/driven/persistence/postgres/model"
	"learning-go/internal/application/port/output"
	"learning-go/internal/domain/entity"
	domainerror "learning-go/internal/domain/error"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) output.ProductRepositoryPort {
	return &ProductRepository{db: db}
}

func (r *ProductRepository) Save(ctx context.Context, product *entity.Product) error {
	m := model.FromProductEntity(product)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	product.CreatedAt = m.CreatedAt
	product.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *ProductRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Product, error) {
	var m model.Product
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domainerror.ErrNotFound
		}
		return nil, err
	}
	return m.ToEntity(), nil
}

func (r *ProductRepository) FindAll(ctx context.Context, offset, limit int) ([]*entity.Product, error) {
	var models []model.Product
	if err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Order("created_at DESC").Find(&models).Error; err != nil {
		return nil, err
	}
	products := make([]*entity.Product, 0, len(models))
	for _, m := range models {
		products = append(products, m.ToEntity())
	}
	return products, nil
}

func (r *ProductRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Product{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *ProductRepository) Update(ctx context.Context, product *entity.Product) error {
	m := model.FromProductEntity(product)
	if err := r.db.WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	product.UpdatedAt = m.UpdatedAt
	return nil
}

func (r *ProductRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Product{}, "id = ?", id).Error
}
