package usecase

import (
	"context"
	"learning-go/internal/auth/application/dto"
	"learning-go/internal/auth/application/port"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthUseCase struct {
	userRepo port.UserRepositoryPort
}

func NewAuthUseCase(userRepo port.UserRepositoryPort) port.AuthUseCasePort {
	return &AuthUseCase{userRepo: userRepo}
}

func (uc *AuthUseCase) GetMe(ctx context.Context, userID uuid.UUID, isFirstLogin bool) (*dto.MeResponse, error) {
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.WithContext(ctx).Error("error finding user", zap.Error(err))
		return nil, sharederror.ErrInternal
	}
	if user == nil {
		return nil, sharederror.ErrNotFound
	}

	return &dto.MeResponse{
		ID:           user.ID.String(),
		PrepUserID:   user.PrepUserID,
		Name:         user.Name,
		Email:        user.Email,
		IsFirstLogin: isFirstLogin,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}, nil
}
