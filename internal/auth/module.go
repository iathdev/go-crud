package auth

import (
	"learning-go/internal/auth/adapter/handler"
	"learning-go/internal/auth/adapter/repository"
	"learning-go/internal/auth/adapter/service"
	"learning-go/internal/auth/application/port"
	"learning-go/internal/auth/application/usecase"
	"learning-go/internal/infrastructure/circuitbreaker"
	"learning-go/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Module struct {
	handler         *handler.AuthHandler
	PrepUserService port.PrepUserServicePort
	UserRepo        port.UserRepositoryPort
}

func NewModule(db *gorm.DB, cfg *config.Config) *Module {
	userRepo := repository.NewUserRepository(db)

	breaker := circuitbreaker.NewBreaker(circuitbreaker.BreakerConfig{
		Name: "prep-user-service",
	}, nil)
	prepUserService := service.NewPrepUserService(cfg.PrepUserServiceURL, cfg.GetPrepMeEndpoint(), cfg.GetPrepHTTPClientTimeout(), breaker)

	authUseCase := usecase.NewAuthUseCase(userRepo)
	authHandler := handler.NewAuthHandler(authUseCase)

	return &Module{
		handler:         authHandler,
		PrepUserService: prepUserService,
		UserRepo:        userRepo,
	}
}

func (module *Module) RegisterRoutes(public, protected *gin.RouterGroup) {
	protected.GET("/me", module.handler.GetMe)
}
