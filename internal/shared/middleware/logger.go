package middleware

import (
	"bytes"
	"io"
	"learning-go/internal/shared/common"
	"learning-go/internal/shared/logger"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxBodyLogSize = 10 * 1024 // 10KB

var defaultSkipPaths = map[string]bool{
	"/health": true,
}

// Paths whose request body may contain sensitive data (passwords, tokens).
var sensitiveBodyPaths = map[string]bool{
	"/login":    true,
	"/register": true,
	"/refresh":  true,
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if defaultSkipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()

		// Capture request body for debug logging (skip sensitive paths)
		var requestBody string
		if !sensitiveBodyPaths[c.Request.URL.Path] {
			if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength <= maxBodyLogSize {
				bodyBytes, err := io.ReadAll(c.Request.Body)
				if err == nil {
					requestBody = string(bodyBytes)
					c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			}
		}

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		ctx := c.Request.Context()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.String("client_ip", common.ResolveClientIP(c.Request)),
		}

		if requestBody != "" {
			fields = append(fields, zap.String("body", requestBody))
		}

		logger.WithContext(ctx).Info("[SERVER] http_request", fields...)
	}
}
