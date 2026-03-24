package usecase

import (
	"context"
	"learning-go/internal/auth/application/dto"
	"learning-go/internal/auth/application/port"
	"learning-go/internal/auth/domain"
	apperr "learning-go/internal/shared/error"

	"github.com/google/uuid"
)

type AuthUseCase struct {
	userRepo port.UserRepositoryPort
}

func NewAuthUseCase(userRepo port.UserRepositoryPort) port.AuthUseCasePort {
	return &AuthUseCase{userRepo: userRepo}
}

func (uc *AuthUseCase) GetMe(ctx context.Context, userID uuid.UUID, prepUser *domain.PrepUser) (*dto.MeResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperr.InternalServerError("auth.find_user_failed", err)
	}
	if user == nil {
		return nil, apperr.NotFound("auth.user_not_found")
	}

	// Prep is source of truth for core profile fields; local DB provides app-specific fields
	return &dto.MeResponse{
		ID:           user.ID.String(),
		PrepUserID:   prepUser.PrepUserID,
		Name:         prepUser.Name,
		Email:        prepUser.Email,
		IsFirstLogin: prepUser.IsFirstLogin,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}
