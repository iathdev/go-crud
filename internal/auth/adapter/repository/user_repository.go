package repository

import (
	"context"
	"errors"
	"learning-go/internal/auth/adapter/repository/model"
	"learning-go/internal/auth/application/port"
	"learning-go/internal/auth/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) port.UserRepositoryPort {
	return &UserRepository{db: db}
}

func (repo *UserRepository) Upsert(ctx context.Context, user *domain.User) error {
	m := model.FromUserEntity(user)
	result := repo.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "prep_user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"email", "name", "updated_at"}),
		}).
		Create(m)
	if result.Error != nil {
		return result.Error
	}
	user.ID = m.ID
	user.CreatedAt = m.CreatedAt
	user.UpdatedAt = m.UpdatedAt
	return nil
}

func (repo *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var m model.UserModel
	if err := repo.db.WithContext(ctx).Where("id = ?", id).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.ToEntity(), nil
}

func (repo *UserRepository) FindByPrepUserID(ctx context.Context, prepUserID int64) (*domain.User, error) {
	var m model.UserModel
	if err := repo.db.WithContext(ctx).Where("prep_user_id = ?", prepUserID).First(&m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.ToEntity(), nil
}

func (repo *UserRepository) Update(ctx context.Context, user *domain.User) error {
	m := model.FromUserEntity(user)
	if err := repo.db.WithContext(ctx).Save(m).Error; err != nil {
		return err
	}
	user.UpdatedAt = m.UpdatedAt
	return nil
}
