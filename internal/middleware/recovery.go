package middleware

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery catches panics and returns a clean 500 JSON response instead
// of leaking stack traces to the client.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: %v", err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"detail": "Internal server error",
				})
			}
		}()
		c.Next()
	}
}
