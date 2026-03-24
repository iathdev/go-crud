package server

import (
	"learning-go/internal/auth"
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/shared/middleware"
	"learning-go/internal/shared/response"
	"learning-go/internal/vocabulary"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"gorm.io/gorm"
)

func NewRouter(
	authModule *auth.Module,
	vocabularyModule *vocabulary.Module,
	db *gorm.DB,
	cfg *config.Config,
) *gin.Engine {
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}
	r := gin.New()
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestIDMiddleware())

	// OpenTelemetry tracing (no-op if OTLP_ENDPOINT is not configured)
	if cfg.OTLPEndpoint != "" {
		r.Use(otelgin.Middleware(cfg.GetServiceName()))
	}

	r.Use(middleware.RequestLoggerMiddleware())
	r.Use(middleware.LanguageMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// Health check
	r.GET("/health", healthHandler(db))

	// Public routes with rate limiting
	public := r.Group("/")
	public.Use(middleware.RateLimitMiddleware(cfg.GetRateLimitRPS(), cfg.GetRateLimitBurst()))

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(authModule.PrepUserService, authModule.UserRepo))

	// Register modules
	authModule.RegisterRoutes(public, api)
	vocabularyModule.RegisterRoutes(public, api)

	r.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "common.route_not_found")
	})

	return r
}

func healthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "cannot get database connection",
			})
			return
		}

		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"error":  "database ping failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	}
}
