package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	rateWindow  = time.Minute
	maxRequests = 60
)

type ipRateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
}

var limiter = &ipRateLimiter{
	requests: make(map[string][]time.Time),
}

// RateLimit limits each IP to 60 requests per minute.
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		windowStart := now.Add(-rateWindow)

		limiter.mu.Lock()
		times := limiter.requests[ip]
		valid := times[:0]
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}
		if len(valid) >= maxRequests {
			limiter.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "คำขอมากเกินไป กรุณาลองใหม่ในภายหลัง",
			})
			return
		}
		limiter.requests[ip] = append(valid, now)
		limiter.mu.Unlock()
		c.Next()
	}
}
