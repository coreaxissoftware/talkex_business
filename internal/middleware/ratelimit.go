package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type bucket struct {
	tokens    float64
	lastFill  time.Time
}

type RateLimiterConfig struct {
	Rate     float64 // tokens per second
	Burst    int     // max tokens (bucket capacity)
	KeyFunc  func(*gin.Context) string
}

func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		Rate:  1, // 60 per minute
		Burst: 60,
		KeyFunc: func(c *gin.Context) string {
			if uid, ok := c.Get("user_id"); ok {
				return uid.(string)
			}
			return c.ClientIP()
		},
	}
}

func RateLimit(cfg RateLimiterConfig) gin.HandlerFunc {
	var mu sync.Mutex
	buckets := make(map[string]*bucket)

	go func() {
		for range time.NewTicker(5 * time.Minute).C {
			mu.Lock()
			cutoff := time.Now().Add(-10 * time.Minute)
			for k, b := range buckets {
				if b.lastFill.Before(cutoff) {
					delete(buckets, k)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		key := cfg.KeyFunc(c)
		now := time.Now()

		mu.Lock()
		b, ok := buckets[key]
		if !ok {
			b = &bucket{tokens: float64(cfg.Burst), lastFill: now}
			buckets[key] = b
		}

		elapsed := now.Sub(b.lastFill).Seconds()
		b.tokens += elapsed * cfg.Rate
		if b.tokens > float64(cfg.Burst) {
			b.tokens = float64(cfg.Burst)
		}
		b.lastFill = now

		if b.tokens < 1 {
			retryAfter := (1 - b.tokens) / cfg.Rate
			mu.Unlock()
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)+1))
			c.JSON(http.StatusTooManyRequests, gin.H{"detail": "Rate limit exceeded"})
			c.Abort()
			return
		}

		b.tokens--
		mu.Unlock()
		c.Next()
	}
}
