package audit

import (
	"bytes"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"github.com/coreaxissoftware/talkex_business/internal/users"
)

// bodyCapture wraps gin.ResponseWriter to also buffer the response body so
// failed requests can be inspected later without re-running them.
type bodyCapture struct {
	gin.ResponseWriter
	buf *bytes.Buffer
}

func (w bodyCapture) Write(b []byte) (int, error) {
	w.buf.Write(b)
	return w.ResponseWriter.Write(b)
}

// Middleware records every request as a LogEntry once the handler chain
// completes. It skips /health and the audit-log routes themselves to avoid
// self-noise, and never blocks the response on the DB write.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		start := time.Now()
		buf := &bytes.Buffer{}
		writer := bodyCapture{ResponseWriter: c.Writer, buf: buf}
		c.Writer = writer

		c.Next()

		status := c.Writer.Status()
		success := status < 400

		entry := &LogEntry{
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			StatusCode: status,
			Success:    success,
			LatencyMs:  time.Since(start).Milliseconds(),
			ClientIP:   c.ClientIP(),
		}

		if uid, ok := c.Get("user_id"); ok {
			id := uid.(string)
			entry.UserID = &id
			if u, err := users.GetByID(database.DB, id); err == nil {
				entry.UserEmail = &u.Email
			}
		}

		if !success && buf.Len() > 0 {
			body := buf.String()
			if len(body) > 2000 {
				body = body[:2000]
			}
			entry.ErrorBody = &body
		}

		// Written synchronously (not fire-and-forget) — SQLite's single
		// writer serializes concurrent connections poorly, and this keeps
		// audit writes ordered. A logging failure never affects the
		// response, which has already been sent to the client at this point.
		_ = Record(database.DB, entry)
	}
}
