package middleware

import (
	"learning-go/internal/adapter/driving/http/response"
	"log"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		log.Printf("[PANIC RECOVERED] %v\n%s", recovered, debug.Stack())
		response.InternalServerError(c, "")
		c.Abort()
	})
}
