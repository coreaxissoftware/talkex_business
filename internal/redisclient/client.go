// Package redisclient wraps the go-redis client so callers can share
// one connection pool and consistently no-op when REDIS_URL is unset.
//
// Every consumer (rate limiter, OTP store, events hub) accepts a
// *redis.Client that may be nil — nil means "fall back to in-memory".
// This keeps the dev experience zero-config while letting production
// deployments scale horizontally by setting one env var.
package redisclient

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	shared     *redis.Client
	sharedOnce sync.Once
)

// Init connects on first call. Returns nil (no error) when url is
// blank so callers can treat "no redis" as a first-class option.
// A failed URL parse or a ping timeout also returns nil and logs —
// we prefer in-memory fallback over crashing the process.
func Init(url string) *redis.Client {
	sharedOnce.Do(func() {
		if url == "" {
			return
		}
		opt, err := redis.ParseURL(url)
		if err != nil {
			log.Printf("redis: bad REDIS_URL, falling back to in-memory: %v", err)
			return
		}
		c := redis.NewClient(opt)
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := c.Ping(ctx).Err(); err != nil {
			log.Printf("redis: ping failed, falling back to in-memory: %v", err)
			_ = c.Close()
			return
		}
		shared = c
		log.Printf("redis: connected to %s", opt.Addr)
	})
	return shared
}

// Get returns the shared client if Init succeeded, else nil.
func Get() *redis.Client { return shared }
