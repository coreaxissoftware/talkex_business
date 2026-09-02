package middleware

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/observability"
)

// Recovery catches panics and returns a clean 500 JSON response instead
// of leaking stack traces to the client. Every panic is also captured
// to Sentry (when SENTRY_DSN is set) with the request URL + method as
// extra context so an on-call engineer sees the route that broke.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				log.Printf("PANIC on %s %s: %v\n%s", c.Request.Method, c.Request.URL.Path, r, stack)

				var err error
				switch v := r.(type) {
				case error:
					err = v
				default:
					err = fmt.Errorf("%v", v)
				}
				observability.CaptureError(err, map[string]interface{}{
					"method":     c.Request.Method,
					"path":       c.Request.URL.Path,
					"request_id": c.GetString("request_id"),
					"user_id":    c.GetString("user_id"),
				})

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"detail":     "Internal server error",
					"request_id": c.GetString("request_id"),
				})
			}
		}()
		c.Next()
	}
}
