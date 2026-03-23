package model

import (
	"learning-go/internal/auth/domain"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key"`
	PrepUserID int64     `gorm:"uniqueIndex;not null"`
	Email      string    `gorm:"not null;default:''"`
	Name       string    `gorm:"not null;default:''"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

func (UserModel) TableName() string {
	return "users"
}

func (model *UserModel) ToEntity() *domain.User {
	return &domain.User{
		ID:         model.ID,
		PrepUserID: model.PrepUserID,
		Email:      model.Email,
		Name:       model.Name,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}

func FromUserEntity(user *domain.User) *UserModel {
	return &UserModel{
		ID:         user.ID,
		PrepUserID: user.PrepUserID,
		Email:      user.Email,
		Name:       user.Name,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	}
}
