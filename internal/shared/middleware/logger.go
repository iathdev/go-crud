package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"learning-go/internal/shared/common"
	"learning-go/internal/shared/logger"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const maxBodyLogSize = 10 * 1024 // 10KB

var defaultSkipPaths = map[string]bool{
	"/health": true,
}

var sensitiveKeys = map[string]bool{
	"password":      true,
	"token":         true,
	"secret":        true,
	"authorization": true,
	"cookie":        true,
	"credit_card":   true,
	"ssn":           true,
	"api_key":       true,
}

func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if defaultSkipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		start := time.Now()

		// Capture request body for debug logging (sensitive fields are masked)
		var requestBody string
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength <= maxBodyLogSize {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = sanitizeBody(bodyBytes)
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
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
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("protocol", c.Request.Proto),
			zap.Int("response_size", c.Writer.Size()),
		}

		if userID, exists := c.Get("user_id"); exists {
			fields = append(fields, zap.String("user_id", userID.(string)))
		}

		if requestBody != "" {
			fields = append(fields, zap.String("body", requestBody))
		}

		log := logger.WithContext(ctx)
		switch {
		case status >= 500:
			log.Error("[SERVER] http_request", fields...)
		case status >= 400:
			log.Warn("[SERVER] http_request", fields...)
		default:
			log.Info("[SERVER] http_request", fields...)
		}
	}
}

func sanitizeBody(body []byte) string {
	var data map[string]any
	if err := json.Unmarshal(body, &data); err != nil {
		return string(body)
	}

	for key := range data {
		if sensitiveKeys[strings.ToLower(key)] {
			data[key] = "***"
		}
	}

	sanitized, err := json.Marshal(data)
	if err != nil {
		return string(body)
	}
	return string(sanitized)
}
