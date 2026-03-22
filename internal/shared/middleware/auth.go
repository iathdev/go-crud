package middleware

import (
	"learning-go/internal/auth/application/port"
	"learning-go/internal/auth/domain"
	"learning-go/internal/shared/common"
	"learning-go/internal/shared/ctxlog"
	"learning-go/internal/shared/logger"
	"learning-go/internal/shared/response"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware(prepService port.PrepUserServicePort, userRepo port.UserRepositoryPort) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			logger.WithContext(c.Request.Context()).Debug("[AUTH] auth rejected",
				zap.String("reason", "missing or invalid authorization header"),
				zap.String("client_ip", common.ResolveClientIP(c.Request)),
			)
			response.Unauthorized(c, "auth.unauthorized")
			c.Abort()
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")

		prepUser, err := prepService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			response.HandleError(c, err)
			c.Abort()
			return
		}

		user, isFirstLogin, err := upsertUser(c, userRepo, prepUser)
		if err != nil {
			logger.WithContext(c.Request.Context()).Error("[AUTH] upsert failed", zap.Error(err))
			response.InternalServerError(c, "")
			c.Abort()
			return
		}

		c.Set("user_id", user.ID.String())
		c.Set("email", user.Email)
		c.Set("prep_user_id", user.PrepUserID)
		c.Set("is_first_login", isFirstLogin)

		ctx := ctxlog.WithFields(c.Request.Context(), zap.String("user_id", user.ID.String()))
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func upsertUser(c *gin.Context, userRepo port.UserRepositoryPort, prepUser *domain.PrepUser) (*domain.User, bool, error) {
	ctx := c.Request.Context()

	existing, err := userRepo.FindByPrepUserID(ctx, prepUser.PrepUserID)
	if err != nil {
		return nil, false, err
	}

	if existing == nil {
		user := domain.NewUser(prepUser.PrepUserID, prepUser.Email, prepUser.Name)
		if err := userRepo.Upsert(ctx, user); err != nil {
			return nil, false, err
		}
		return user, true, nil
	}

	if existing.Email != prepUser.Email || existing.Name != prepUser.Name {
		existing.Email = prepUser.Email
		existing.Name = prepUser.Name
		if err := userRepo.Update(ctx, existing); err != nil {
			return nil, false, err
		}
	}

	return existing, false, nil
}
