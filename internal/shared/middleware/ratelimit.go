package middleware

import (
	"learning-go/internal/shared/common"
	"learning-go/internal/shared/logger"
	"learning-go/internal/shared/response"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type visitor struct {
	tokens    float64
	lastSeen  time.Time
	maxTokens float64
	rate      float64
}

func (vis *visitor) allow() bool {
	now := time.Now()
	elapsed := now.Sub(vis.lastSeen).Seconds()
	vis.lastSeen = now
	vis.tokens += elapsed * vis.rate
	if vis.tokens > vis.maxTokens {
		vis.tokens = vis.maxTokens
	}
	if vis.tokens < 1 {
		return false
	}
	vis.tokens--
	return true
}

type RateLimiter struct {
	mu        sync.Mutex
	visitors  map[string]*visitor
	rate      float64
	maxTokens float64
}

func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors:  make(map[string]*visitor),
		rate:      requestsPerSecond,
		maxTokens: float64(burst),
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getVisitor(ip string) *visitor {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		v = &visitor{
			tokens:    rl.maxTokens,
			lastSeen:  time.Now(),
			maxTokens: rl.maxTokens,
			rate:      rl.rate,
		}
		rl.visitors[ip] = v
	}
	return v
}

func RateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(requestsPerSecond, burst)

	return func(c *gin.Context) {
		ip := common.ResolveClientIP(c.Request)
		v := limiter.getVisitor(ip)

		limiter.mu.Lock()
		allowed := v.allow()
		limiter.mu.Unlock()

		if !allowed {
			logger.WithContext(c.Request.Context()).Warn("rate limit exceeded", zap.String("client_ip", ip))
			response.Error(c, 429, "common.too_many_requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
