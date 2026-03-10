package http

import (
	"learning-go/internal/adapter/driving/http/handler"
	"learning-go/internal/adapter/driving/http/middleware"
	"learning-go/internal/infrastructure/config"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	authHandler *handler.AuthHandler,
	productHandler *handler.ProductHandler,
	healthHandler *handler.HealthHandler,
	cfg *config.Config,
) *gin.Engine {
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(middleware.LanguageMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// Health check
	r.GET("/health", healthHandler.Health)

	// Public routes with rate limiting
	public := r.Group("/")
	public.Use(middleware.RateLimitMiddleware(5, 10))
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
	}

	// Private routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg))
	{
		api.POST("/products", productHandler.CreateProduct)
		api.GET("/products/:id", productHandler.GetProduct)
		api.GET("/products", productHandler.ListProducts)
		api.PUT("/products/:id", productHandler.UpdateProduct)
		api.DELETE("/products/:id", productHandler.DeleteProduct)
	}

	return r
}
