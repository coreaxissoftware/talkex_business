package events

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
)

// RegisterRoutes wires the SSE stream endpoint.
//
// The browser's EventSource can't set an Authorization header, so this
// endpoint accepts the JWT via `?token=...` query param. The token
// still has to be a valid access token — no extra privilege exposure
// versus the header path.
func RegisterRoutes(r *gin.Engine) {
	r.GET("/events/stream", handleStream)
}

func handleStream(c *gin.Context) {
	// Auth via query param (EventSource limitation)
	tok := c.Query("token")
	if tok == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Missing token"})
		return
	}
	userID, err := auth.ValidateToken(tok, auth.AccessToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid token"})
		return
	}

	// SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // disable nginx buffering

	sub := Subscribe(userID)
	defer Unsubscribe(sub)

	// Kick off with a hello so the client knows the stream is open.
	fmt.Fprintf(c.Writer, "event: hello\ndata: {\"ok\":true}\n\n")
	c.Writer.Flush()

	// Heartbeat keeps proxies from closing the idle connection.
	heartbeat := time.NewTicker(20 * time.Second)
	defer heartbeat.Stop()

	notify := c.Request.Context().Done()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-notify:
			return false
		case <-heartbeat.C:
			fmt.Fprintf(w, ": ping\n\n")
			return true
		case ev, ok := <-sub.Channel():
			if !ok {
				return false
			}
			payload, err := ev.Encoded()
			if err != nil {
				return true
			}
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Type, payload)
			return true
		}
	})
}
