package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNameRequired            = errors.New("product name is required")
	ErrProductPriceMustBeGreaterThan0 = errors.New("product price must be greater than 0")
)

type Product struct {
	ID          uuid.UUID
	Name        string
	Description string
	Price       float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewProduct(name, description string, price float64) (*Product, error) {
	if name == "" {
		return nil, ErrProductNameRequired
	}
	if price <= 0 {
		return nil, ErrProductPriceMustBeGreaterThan0
	}

	return &Product{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Price:       price,
	}, nil
}

func (p *Product) Update(name, description string, price float64) error {
	if name == "" {
		return ErrProductNameRequired
	}
	if price <= 0 {
		return ErrProductPriceMustBeGreaterThan0
	}

	p.Name = name
	p.Description = description
	p.Price = price
	return nil
}
