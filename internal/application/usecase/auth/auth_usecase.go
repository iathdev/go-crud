package auth

import (
	"context"
	"learning-go/internal/application/dto"
	"learning-go/internal/application/port/input"
	"learning-go/internal/application/port/output"
	"learning-go/internal/domain/entity"
	domainerror "learning-go/internal/domain/error"
	"log"
)

type AuthUseCase struct {
	userRepo     output.UserRepositoryPort
	tokenService output.TokenServicePort
}

func NewAuthUseCase(userRepo output.UserRepositoryPort, tokenService output.TokenServicePort) input.AuthUseCasePort {
	return &AuthUseCase{
		userRepo:     userRepo,
		tokenService: tokenService,
	}
}

func (u *AuthUseCase) Register(ctx context.Context, req dto.RegisterRequest) (*dto.RegisterResponse, error) {
	existingUser, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("error finding user by email: %v", err)
		return nil, domainerror.ErrInternal
	}
	if existingUser != nil {
		return nil, domainerror.ErrEmailAlreadyExists
	}

	user, err := entity.NewUser(req.Email, req.Password, req.Name)
	if err != nil {
		log.Printf("error creating user entity: %v", err)
		return nil, domainerror.ErrInternal
	}

	if err := u.userRepo.Save(ctx, user); err != nil {
		log.Printf("error saving user: %v", err)
		return nil, domainerror.ErrInternal
	}

	token, err := u.tokenService.GenerateToken(user)
	if err != nil {
		log.Printf("error generating token: %v", err)
		return nil, domainerror.ErrInternal
	}

	return &dto.RegisterResponse{
		User: dto.UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
		},
		AccessToken: token,
	}, nil
}

func (u *AuthUseCase) Login(ctx context.Context, req dto.LoginRequest) (*dto.TokenResponse, error) {
	user, err := u.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		log.Printf("error finding user by email: %v", err)
		return nil, domainerror.ErrInternal
	}
	if user == nil {
		return nil, domainerror.ErrUnauthorized
	}

	if !user.CheckPassword(req.Password) {
		return nil, domainerror.ErrUnauthorized
	}

	token, err := u.tokenService.GenerateToken(user)
	if err != nil {
		log.Printf("error generating token: %v", err)
		return nil, domainerror.ErrInternal
	}

	return &dto.TokenResponse{
		AccessToken: token,
	}, nil
}
