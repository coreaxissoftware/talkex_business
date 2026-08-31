package middleware

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/redisclient"
)

// LoginGuard tracks failed login attempts by (email, ip) and progressively
// slows / locks the account after successive failures.
//
// Thresholds (banking-industry-standard baseline):
//   3 fails  → 30-second cool-off
//   5 fails  → 5-minute lock
//   10 fails → 30-minute lock
//
// The counter clears on a successful login (recorded via
// LoginSuccessHook the auth handler must call).
type LoginGuard struct {
	mu   sync.Mutex
	fail map[string]*failRecord
}

type failRecord struct {
	count      int
	lockedUnti time.Time
}

var defaultGuard = &LoginGuard{fail: make(map[string]*failRecord)}

const (
	guardTTL       = 60 * time.Minute
	failRedisNS    = "auth:fail:"
	failRedisTTL   = 60 * time.Minute
)

func init() {
	go func() {
		for range time.NewTicker(15 * time.Minute).C {
			defaultGuard.mu.Lock()
			cutoff := time.Now().Add(-guardTTL)
			for k, r := range defaultGuard.fail {
				if r.lockedUnti.Before(cutoff) {
					delete(defaultGuard.fail, k)
				}
			}
			defaultGuard.mu.Unlock()
		}
	}()
}

// LoginBruteForceGuard is a gin middleware that only runs for the
// /auth/login route. It peeks at the JSON body to extract the email,
// composes a key with the client IP, and refuses the request if the
// account is currently locked.
func LoginBruteForceGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path != "/auth/login" || c.Request.Method != http.MethodPost {
			c.Next()
			return
		}

		email := peekEmailFromBody(c)
		ip := c.ClientIP()
		key := strings.ToLower(email) + "|" + ip

		if until, locked := checkLocked(c.Request.Context(), key); locked {
			retry := int(time.Until(until).Seconds()) + 1
			c.Header("Retry-After", itoa(retry))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"detail":       "Too many failed login attempts. Try again later.",
				"retry_after":  retry,
				"locked_until": until.UTC().Format(time.RFC3339),
			})
			return
		}

		// Stash the key so the handler can record success/failure.
		c.Set("login_guard_key", key)
		c.Next()
	}
}

// RecordLoginFailure is called by the auth handler when credentials
// don't match. Increments the counter and, past thresholds, sets a lock.
func RecordLoginFailure(ctx context.Context, key string) {
	defaultGuard.mu.Lock()
	defer defaultGuard.mu.Unlock()

	r, ok := defaultGuard.fail[key]
	if !ok {
		r = &failRecord{}
		defaultGuard.fail[key] = r
	}
	r.count++

	switch {
	case r.count >= 10:
		r.lockedUnti = time.Now().Add(30 * time.Minute)
	case r.count >= 5:
		r.lockedUnti = time.Now().Add(5 * time.Minute)
	case r.count >= 3:
		r.lockedUnti = time.Now().Add(30 * time.Second)
	}

	// Mirror to Redis for cross-pod visibility, best-effort.
	if rdb := redisclient.Get(); rdb != nil {
		cx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		rdb.HSet(cx, failRedisNS+key, "c", r.count, "u", r.lockedUnti.Unix())
		rdb.Expire(cx, failRedisNS+key, failRedisTTL)
	}
}

// RecordLoginSuccess clears any recorded failures for the key.
func RecordLoginSuccess(ctx context.Context, key string) {
	defaultGuard.mu.Lock()
	delete(defaultGuard.fail, key)
	defaultGuard.mu.Unlock()

	if rdb := redisclient.Get(); rdb != nil {
		cx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		rdb.Del(cx, failRedisNS+key)
	}
}

func checkLocked(ctx context.Context, key string) (time.Time, bool) {
	// Redis first so a lock set from another pod is respected.
	if rdb := redisclient.Get(); rdb != nil {
		cx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		vals, err := rdb.HMGet(cx, failRedisNS+key, "u").Result()
		if err == nil && len(vals) == 1 {
			if s, ok := vals[0].(string); ok && s != "" {
				var ts int64
				for _, ch := range s {
					if ch < '0' || ch > '9' {
						return time.Time{}, false
					}
					ts = ts*10 + int64(ch-'0')
				}
				until := time.Unix(ts, 0)
				if until.After(time.Now()) {
					return until, true
				}
			}
		}
	}
	defaultGuard.mu.Lock()
	defer defaultGuard.mu.Unlock()
	r, ok := defaultGuard.fail[key]
	if !ok {
		return time.Time{}, false
	}
	if r.lockedUnti.After(time.Now()) {
		return r.lockedUnti, true
	}
	return time.Time{}, false
}

// peekEmailFromBody reads the JSON body, parses just the "email" field,
// and puts the body back so the downstream handler still sees it. Never
// fails — an unreadable body means the key falls back to IP-only.
func peekEmailFromBody(c *gin.Context) string {
	if c.Request.Body == nil {
		return ""
	}
	// Cap the peek so a huge Content-Length can't OOM us.
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, 8192))
	if err != nil {
		return ""
	}
	// Restore the body for the handler.
	c.Request.Body = io.NopCloser(strings.NewReader(string(body)))

	var payload struct {
		Email string `json:"email"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}
	return payload.Email
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 8)
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf = append(buf, byte('0'+n%10))
		n /= 10
	}
	// reverse
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	if neg {
		return "-" + string(buf)
	}
	return string(buf)
}
