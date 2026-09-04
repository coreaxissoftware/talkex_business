// Package apihelpers — shared HTTP response helpers.
//
// Exists solely to plug the silent-500 hole: an audit of the codebase
// found ~200 handlers doing
//
//     c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
//
// with zero surrounding log line. When one of these fires in prod,
// nobody — merchant or on-call — sees what actually failed.
// ServerError() logs the error with the request context (method, path,
// request_id, user_id) and Sentry-captures it before returning the
// same clean JSON we always did, so the fix is one function call at
// each site rather than 200 hand-added log lines.
package apihelpers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/observability"
)

// ServerError logs the (usually swallowed) error with request context
// and captures it to Sentry, then writes the standard 500 body.
// Passes through the merchant-facing string unchanged so nothing about
// the outward surface shifts.
//
// Typical use replaces:
//
//	if err != nil {
//	    c.JSON(500, gin.H{"detail": "Internal server error"})
//	    return
//	}
//
// with:
//
//	if err != nil {
//	    apihelpers.ServerError(c, err, "list deals")
//	    return
//	}
func ServerError(c *gin.Context, err error, context string) {
	if err == nil {
		return
	}
	requestID := c.GetString("request_id")
	extras := map[string]interface{}{
		"context":    context,
		"method":     c.Request.Method,
		"path":       c.Request.URL.Path,
		"request_id": requestID,
		"user_id":    c.GetString("user_id"),
	}
	// Sentry gets the full context; every request also drops a log line
	// so a merchant support ticket with just the request_id is enough
	// for on-call to grep.
	observability.CaptureError(err, extras)

	body := gin.H{"detail": "Internal server error"}
	if requestID != "" {
		body["request_id"] = requestID
	}
	c.JSON(http.StatusInternalServerError, body)
}

// BadRequest is the same pattern for 400s with a merchant-facing message.
// Not logged to Sentry — a 400 is the merchant's fault, not ours.
func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{"detail": message})
}

// NotFound is the same pattern for 404s.
func NotFound(c *gin.Context, message string) {
	if message == "" {
		message = "Not found"
	}
	c.JSON(http.StatusNotFound, gin.H{"detail": message})
}
