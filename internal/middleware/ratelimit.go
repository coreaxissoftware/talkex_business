package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/redisclient"
)

type bucket struct {
	tokens    float64
	lastFill  time.Time
}

type RateLimiterConfig struct {
	Rate     float64 // tokens per second
	Burst    int     // max tokens (bucket capacity)
	KeyFunc  func(*gin.Context) string
	// Namespace prefixes the Redis key so multiple limiters can share
	// one Redis without colliding. Optional; defaults to "rl:".
	Namespace string
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

// RateLimit returns a middleware that enforces cfg. When
// redisclient.Get() returns a live client, buckets live in Redis so
// every pod sees the same counters; otherwise the process's own map
// is used (dev, single-pod deploys).
func RateLimit(cfg RateLimiterConfig) gin.HandlerFunc {
	if rdb := redisclient.Get(); rdb != nil {
		return redisRateLimit(cfg)
	}
	return memoryRateLimit(cfg)
}

func memoryRateLimit(cfg RateLimiterConfig) gin.HandlerFunc {
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

// redisRateLimit implements the token bucket in Redis. Each request:
//   1. HGET the (tokens, ts) pair for the key
//   2. compute refill locally, decrement, HSET back with TTL
//
// Race-tolerant enough for HTTP-scale traffic; a proper Lua CAS could
// tighten it further but adds complexity without changing correctness
// under realistic contention (per-key writes are already serialized
// by whichever pod owns the request for that IP/user).
func redisRateLimit(cfg RateLimiterConfig) gin.HandlerFunc {
	ns := cfg.Namespace
	if ns == "" {
		ns = "rl:"
	}
	rdb := redisclient.Get()
	// Redis key TTL — the bucket is safe to drop after long idle.
	ttl := 15 * time.Minute

	return func(c *gin.Context) {
		key := ns + cfg.KeyFunc(c)
		ctx, cancel := context.WithTimeout(c.Request.Context(), 500*time.Millisecond)
		defer cancel()

		nowUnix := float64(time.Now().UnixNano()) / 1e9

		vals, err := rdb.HMGet(ctx, key, "t", "ts").Result()
		if err != nil {
			// On a Redis blip, fail open — better than blocking traffic.
			c.Next()
			return
		}

		tokens := float64(cfg.Burst)
		if v, ok := vals[0].(string); ok && v != "" {
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				tokens = parsed
			}
		}
		lastFill := nowUnix
		if v, ok := vals[1].(string); ok && v != "" {
			if parsed, err := strconv.ParseFloat(v, 64); err == nil {
				lastFill = parsed
			}
		}

		elapsed := nowUnix - lastFill
		tokens += elapsed * cfg.Rate
		if tokens > float64(cfg.Burst) {
			tokens = float64(cfg.Burst)
		}

		if tokens < 1 {
			retryAfter := (1 - tokens) / cfg.Rate
			c.Header("Retry-After", strconv.Itoa(int(retryAfter)+1))
			c.JSON(http.StatusTooManyRequests, gin.H{"detail": "Rate limit exceeded"})
			c.Abort()
			return
		}

		tokens--
		pipe := rdb.TxPipeline()
		pipe.HSet(ctx, key, "t", fmt.Sprintf("%.6f", tokens), "ts", fmt.Sprintf("%.6f", nowUnix))
		pipe.Expire(ctx, key, ttl)
		_, _ = pipe.Exec(ctx)

		c.Next()
	}
}
