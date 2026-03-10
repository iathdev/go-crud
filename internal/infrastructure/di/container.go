package di

import (
	"learning-go/internal/adapter/driven/persistence/postgres"
	"learning-go/internal/adapter/driving/http"
	"learning-go/internal/adapter/driving/http/handler"
	"learning-go/internal/application/usecase/auth"
	"learning-go/internal/application/usecase/product"
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/infrastructure/database"
	"learning-go/internal/infrastructure/security"
	"log"
)

func NewApp() (*http.Server, func(), error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// Repositories
	userRepo := postgres.NewUserRepository(db)
	productRepo := postgres.NewProductRepository(db)

	// Services
	tokenService := security.NewJWTService(cfg)

	// UseCases
	authUseCase := auth.NewAuthUseCase(userRepo, tokenService)
	productCommandUseCase := product.NewProductCommandUseCase(productRepo)
	productQueryUseCase := product.NewProductQueryUseCase(productRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	productHandler := handler.NewProductHandler(productCommandUseCase, productQueryUseCase)
	healthHandler := handler.NewHealthHandler(db)

	// Router
	router := http.NewRouter(authHandler, productHandler, healthHandler, cfg)

	// Server
	server := http.NewServer(cfg, router)

	log.Println("App initialized successfully")
	return server, cleanup, nil
}
