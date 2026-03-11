package middleware

import (
	"learning-go/internal/shared/ctxlog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const RequestIDHeader = "X-Request-ID"

func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.Must(uuid.NewV7()).String()
		}

		c.Set("request_id", requestID)
		c.Header(RequestIDHeader, requestID)

		ctx := ctxlog.WithFields(c.Request.Context(), zap.String("request_id", requestID))
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
