package entity

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNewProduct(t *testing.T) {
	product, err := NewProduct("Test Product", "A description", 9.99)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if product.Name != "Test Product" {
		t.Errorf("expected name Test Product, got %s", product.Name)
	}
	if product.Description != "A description" {
		t.Errorf("expected description A description, got %s", product.Description)
	}
	if product.Price != 9.99 {
		t.Errorf("expected price 9.99, got %f", product.Price)
	}
	if product.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}
}

func TestNewProduct_EmptyName(t *testing.T) {
	_, err := NewProduct("", "desc", 10)
	if !errors.Is(err, ErrProductNameRequired) {
		t.Errorf("expected ErrProductNameRequired, got %v", err)
	}
}

func TestNewProduct_ZeroPrice(t *testing.T) {
	_, err := NewProduct("Test", "desc", 0)
	if !errors.Is(err, ErrProductPriceMustBeGreaterThan0) {
		t.Errorf("expected ErrProductPriceMustBeGreaterThan0, got %v", err)
	}
}

func TestNewProduct_NegativePrice(t *testing.T) {
	_, err := NewProduct("Test", "desc", -5)
	if !errors.Is(err, ErrProductPriceMustBeGreaterThan0) {
		t.Errorf("expected ErrProductPriceMustBeGreaterThan0, got %v", err)
	}
}

func TestProduct_Update(t *testing.T) {
	product, _ := NewProduct("Old Name", "Old Desc", 10)
	err := product.Update("New Name", "New Desc", 20)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if product.Name != "New Name" {
		t.Errorf("expected New Name, got %s", product.Name)
	}
	if product.Price != 20 {
		t.Errorf("expected price 20, got %f", product.Price)
	}
}

func TestProduct_Update_EmptyName(t *testing.T) {
	product, _ := NewProduct("Name", "Desc", 10)
	err := product.Update("", "Desc", 10)
	if !errors.Is(err, ErrProductNameRequired) {
		t.Errorf("expected ErrProductNameRequired, got %v", err)
	}
}

func TestProduct_Update_InvalidPrice(t *testing.T) {
	product, _ := NewProduct("Name", "Desc", 10)
	err := product.Update("Name", "Desc", -1)
	if !errors.Is(err, ErrProductPriceMustBeGreaterThan0) {
		t.Errorf("expected ErrProductPriceMustBeGreaterThan0, got %v", err)
	}
}
