package usecase

import (
	"context"
	"learning-go/internal/auth/application/dto"
	"learning-go/internal/auth/application/port"
	"learning-go/internal/auth/domain"
	sharederror "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthUseCase struct {
	prepService       port.PrepUserServicePort
	userRepo          port.UserRepositoryPort
	tokenService      port.TokenServicePort
	refreshTokenStore port.RefreshTokenStorePort
	refreshExpiry     time.Duration
}

func NewAuthUseCase(
	prepService port.PrepUserServicePort,
	userRepo port.UserRepositoryPort,
	tokenService port.TokenServicePort,
	refreshTokenStore port.RefreshTokenStorePort,
	refreshExpiry time.Duration,
) port.AuthUseCasePort {
	return &AuthUseCase{
		prepService:       prepService,
		userRepo:          userRepo,
		tokenService:      tokenService,
		refreshTokenStore: refreshTokenStore,
		refreshExpiry:     refreshExpiry,
	}
}

func (useCase *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	prepUser, err := useCase.prepService.ValidateToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}

	existing, err := useCase.userRepo.FindByPrepUserID(ctx, prepUser.PrepUserID)
	if err != nil {
		logger.WithContext(ctx).Error("error finding user by prep user id", zap.Error(err))
		return nil, sharederror.ErrInternal
	}

	isFirstLogin := existing == nil
	var user *domain.User

	if isFirstLogin {
		user = domain.NewUser(prepUser.PrepUserID, prepUser.Email, prepUser.Name)
		if err := useCase.userRepo.Upsert(ctx, user); err != nil {
			logger.WithContext(ctx).Error("error creating user", zap.Error(err))
			return nil, sharederror.ErrInternal
		}
	} else {
		existing.Email = prepUser.Email
		existing.Name = prepUser.Name
		if err := useCase.userRepo.Update(ctx, existing); err != nil {
			logger.WithContext(ctx).Error("error updating user", zap.Error(err))
			return nil, sharederror.ErrInternal
		}
		user = existing
	}

	accessToken, refreshToken, err := useCase.generateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsFirstLogin: isFirstLogin,
		Profile:      toUserResponse(user),
	}, nil
}

func (useCase *AuthUseCase) RefreshToken(ctx context.Context, req dto.RefreshTokenRequest) (*dto.TokenResponse, error) {
	userIDStr, err := useCase.refreshTokenStore.Find(ctx, req.RefreshToken)
	if err != nil {
		if _, ok := sharederror.IsAppError(err); ok {
			logger.WithContext(ctx).Debug("invalid refresh token", zap.Error(err))
			return nil, sharederror.ErrUnauthorized
		}
		logger.WithContext(ctx).Error("error finding refresh token", zap.Error(err))
		return nil, sharederror.ErrInternal
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.WithContext(ctx).Error("invalid user id in refresh token", zap.Error(err))
		return nil, sharederror.ErrInternal
	}

	user, err := useCase.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.WithContext(ctx).Error("error finding user", zap.Error(err))
		return nil, sharederror.ErrInternal
	}
	if user == nil {
		return nil, sharederror.ErrUnauthorized
	}

	_ = useCase.refreshTokenStore.Delete(ctx, req.RefreshToken)

	accessToken, refreshToken, err := useCase.generateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (useCase *AuthUseCase) Logout(ctx context.Context, req dto.LogoutRequest) error {
	return useCase.refreshTokenStore.Delete(ctx, req.RefreshToken)
}

func (useCase *AuthUseCase) GetProfile(ctx context.Context, userID uuid.UUID) (*dto.UserResponse, error) {
	user, err := useCase.userRepo.FindByID(ctx, userID)
	if err != nil {
		logger.WithContext(ctx).Error("error finding user", zap.Error(err))
		return nil, sharederror.ErrInternal
	}
	if user == nil {
		return nil, sharederror.ErrNotFound
	}
	return toUserResponse(user), nil
}

func (useCase *AuthUseCase) generateTokenPair(ctx context.Context, user *domain.User) (string, string, error) {
	accessToken, err := useCase.tokenService.GenerateToken(user)
	if err != nil {
		logger.WithContext(ctx).Error("error generating access token", zap.Error(err))
		return "", "", sharederror.ErrInternal
	}

	refreshToken, err := useCase.tokenService.GenerateRefreshToken()
	if err != nil {
		logger.WithContext(ctx).Error("error generating refresh token", zap.Error(err))
		return "", "", sharederror.ErrInternal
	}

	if err := useCase.refreshTokenStore.Save(ctx, user.ID.String(), refreshToken, useCase.refreshExpiry); err != nil {
		logger.WithContext(ctx).Error("error saving refresh token", zap.Error(err))
		return "", "", sharederror.ErrInternal
	}

	return accessToken, refreshToken, nil
}

func toUserResponse(user *domain.User) *dto.UserResponse {
	return &dto.UserResponse{
		UserID: user.ID.String(),
		Email:  user.Email,
		Name:   user.Name,
	}
}
