package usecase

import (
	"context"
	"learning-go/internal/auth/application/dto"
	"learning-go/internal/auth/application/port"
	sharederror "learning-go/internal/shared/error"

	"github.com/google/uuid"
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
		return nil, sharederror.InternalServerError("auth.find_user_failed", err)
	}
	if user == nil {
		return nil, sharederror.NotFound("auth.user_not_found")
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
