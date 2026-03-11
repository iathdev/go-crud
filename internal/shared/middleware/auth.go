package middleware

import (
	"fmt"
	"learning-go/internal/infrastructure/config"
	"learning-go/internal/shared/common"
	"learning-go/internal/shared/ctxlog"
	"learning-go/internal/shared/logger"
	"learning-go/internal/shared/response"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.WithContext(c.Request.Context()).Debug("auth rejected", zap.String("reason", "missing authorization header"), zap.String("client_ip", common.ResolveClientIP(c.Request)))
			response.Unauthorized(c, "auth.unauthorized")
			c.Abort()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			logger.WithContext(c.Request.Context()).Debug("auth rejected", zap.String("reason", "invalid bearer prefix"), zap.String("client_ip", common.ResolveClientIP(c.Request)))
			response.Unauthorized(c, "auth.unauthorized")
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			logger.WithContext(c.Request.Context()).Debug("auth rejected", zap.String("reason", "invalid token"), zap.String("client_ip", common.ResolveClientIP(c.Request)), zap.Error(err))
			response.Unauthorized(c, "auth.unauthorized")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			logger.WithContext(c.Request.Context()).Debug("auth rejected", zap.String("reason", "invalid claims"), zap.String("client_ip", common.ResolveClientIP(c.Request)))
			response.Unauthorized(c, "auth.unauthorized")
			c.Abort()
			return
		}

		if userID, ok := claims["user_id"].(string); ok {
			c.Set("user_id", userID)
			ctx := ctxlog.WithFields(c.Request.Context(), zap.String("user_id", userID))
			c.Request = c.Request.WithContext(ctx)
		}
		if email, ok := claims["email"].(string); ok {
			c.Set("email", email)
		}

		c.Next()
	}
}
