package auth

import (
	"learning-go/internal/auth/adapter/handler"
	"learning-go/internal/auth/adapter/repository"
	"learning-go/internal/auth/adapter/service"
	"learning-go/internal/auth/application/usecase"
	"learning-go/internal/infrastructure/circuitbreaker"
	"learning-go/internal/infrastructure/config"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type Module struct {
	handler *handler.AuthHandler
}

func NewModule(db *gorm.DB, redisClient *redis.Client, cfg *config.Config) *Module {
	userRepo := repository.NewUserRepository(db)
	refreshStore := repository.NewRefreshTokenRepository(redisClient)
	tokenService := service.NewJWTService(cfg)

	breaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
		Name: "prep-user-service",
	}, nil)
	prepUserService := service.NewPrepUserService(cfg.PrepUserServiceURL, breaker)

	refreshExpiry := time.Duration(cfg.GetRefreshTokenExpiry()) * 24 * time.Hour
	authUseCase := usecase.NewAuthUseCase(prepUserService, userRepo, tokenService, refreshStore, refreshExpiry)
	authHandler := handler.NewAuthHandler(authUseCase)

	return &Module{handler: authHandler}
}

func (module *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	public.POST("/login", module.handler.Login)
	public.POST("/refresh", module.handler.RefreshToken)

	protected.POST("/logout", module.handler.Logout)
	protected.GET("/profile", module.handler.GetProfile)
}
