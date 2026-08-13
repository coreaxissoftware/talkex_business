package middleware

import (
	"bytes"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	idempotencyHeader = "X-Idempotency-Key"
	idempotencyTTL    = 24 * time.Hour
)

type idempotencyEntry struct {
	status int
	body   []byte
	at     time.Time
}

type idempotencyStore struct {
	mu    sync.RWMutex
	items map[string]*idempotencyEntry
}

func newIdempotencyStore() *idempotencyStore {
	s := &idempotencyStore{items: make(map[string]*idempotencyEntry)}
	go s.cleanup()
	return s
}

func (s *idempotencyStore) get(key string) (*idempotencyEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	e, ok := s.items[key]
	if ok && time.Since(e.at) > idempotencyTTL {
		return nil, false
	}
	return e, ok
}

func (s *idempotencyStore) set(key string, status int, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = &idempotencyEntry{status: status, body: body, at: time.Now()}
}

func (s *idempotencyStore) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		for k, e := range s.items {
			if time.Since(e.at) > idempotencyTTL {
				delete(s.items, k)
			}
		}
		s.mu.Unlock()
	}
}

type responseRecorder struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

var globalIdempotencyStore = newIdempotencyStore()

// Idempotency middleware — if a request carries an X-Idempotency-Key header
// and the same key was seen before (within TTL), the cached response is
// replayed without re-executing the handler.
func Idempotency() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(idempotencyHeader)
		if key == "" {
			c.Next()
			return
		}

		userID, _ := c.Get("user_id")
		compositeKey := key
		if uid, ok := userID.(string); ok {
			compositeKey = uid + ":" + key
		}

		if cached, ok := globalIdempotencyStore.get(compositeKey); ok {
			c.Data(cached.status, "application/json", cached.body)
			c.Abort()
			return
		}

		rec := &responseRecorder{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = rec

		c.Next()

		if c.Writer.Status() >= 200 && c.Writer.Status() < 500 {
			globalIdempotencyStore.set(compositeKey, c.Writer.Status(), rec.body.Bytes())
		}
	}
}
