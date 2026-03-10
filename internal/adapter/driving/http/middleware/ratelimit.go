package middleware

import (
	"learning-go/internal/adapter/driving/http/response"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	tokens    float64
	lastSeen  time.Time
	maxTokens float64
	rate      float64
}

func (v *visitor) allow() bool {
	now := time.Now()
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.lastSeen = now
	v.tokens += elapsed * v.rate
	if v.tokens > v.maxTokens {
		v.tokens = v.maxTokens
	}
	if v.tokens < 1 {
		return false
	}
	v.tokens--
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
		ip := c.ClientIP()
		v := limiter.getVisitor(ip)

		limiter.mu.Lock()
		allowed := v.allow()
		limiter.mu.Unlock()

		if !allowed {
			response.Error(c, 429, "common.too_many_requests")
			c.Abort()
			return
		}
		c.Next()
	}
}
