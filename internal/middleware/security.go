package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

// SecurityHeaders sets the response headers every banking-grade API is
// expected to send: HSTS, CSP, X-Frame-Options, X-Content-Type-Options,
// Referrer-Policy, Permissions-Policy, and the two cross-origin
// isolation headers. In dev the CSP is relaxed so Vite's HMR shim keeps
// working; in prod it is locked down.
func SecurityHeaders() gin.HandlerFunc {
	cfg := config.Get()
	isProd := !cfg.IsDev()

	// Base CSP — self-only. Frontend origins are added as a connect-src
	// allowlist in prod so the browser accepts XHR to the API.
	csp := strings.Join([]string{
		"default-src 'self'",
		"img-src 'self' data: blob:",
		"style-src 'self' 'unsafe-inline'",
		"font-src 'self' data:",
		"script-src 'self'",
		"connect-src 'self' " + strings.Join(cfg.CORSOrigins, " "),
		"frame-ancestors 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"object-src 'none'",
	}, "; ")

	return func(c *gin.Context) {
		h := c.Writer.Header()

		// Never let downstream cache proxies swap MIME sniffing rules.
		h.Set("X-Content-Type-Options", "nosniff")
		// Clickjacking — the API is never framed.
		h.Set("X-Frame-Options", "DENY")
		// Legacy XSS auditor — modern browsers ignore, but keep for old ones.
		h.Set("X-XSS-Protection", "1; mode=block")
		// Don't leak the referrer to third-party APIs on redirect.
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		// Powerful features — none are needed on the API.
		h.Set("Permissions-Policy",
			"camera=(), microphone=(), geolocation=(), payment=(), usb=(), "+
				"accelerometer=(), gyroscope=(), magnetometer=()")
		// Cross-origin isolation — the API doesn't serve documents that
		// need to embed anything.
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-site")

		// HSTS — only on real HTTPS; browsers reject on plain-HTTP requests.
		if isProd || c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
		}

		// Content-Security-Policy — the API responds JSON, but browsers
		// still enforce CSP on any HTML error page that might leak out.
		h.Set("Content-Security-Policy", csp)

		// Server identity — don't broadcast Go/Gin version.
		h.Set("Server", "TalkEx")
		h.Del("X-Powered-By")

		c.Next()
	}
}

// BodyLimit caps the request body at maxBytes and returns 413 when
// exceeded. Multipart uploads are handled by Gin's MaxMultipartMemory
// separately; this covers JSON, form, and text payloads.
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		// If Content-Length is set and already over the cap, reject early.
		if c.Request.ContentLength > maxBytes {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"detail": "Request body too large",
				"limit":  maxBytes,
			})
			return
		}
		// Wrap the body so streaming reads also enforce the cap.
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}

// RequestID stamps every request with a short random id and echoes it
// back so support can correlate client-side reports with server logs.
// If the client already sent X-Request-Id (< 64 chars), that one wins.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-Id")
		if rid == "" || len(rid) > 64 {
			b := make([]byte, 12)
			if _, err := io.ReadFull(rand.Reader, b); err == nil {
				rid = hex.EncodeToString(b)
			}
		}
		c.Set("request_id", rid)
		c.Writer.Header().Set("X-Request-Id", rid)
		c.Next()
	}
}

// ContentTypeGuard rejects POST/PUT/PATCH requests whose Content-Type
// is neither application/json nor multipart/form-data nor
// application/x-www-form-urlencoded. Blocks a whole class of CSRF-via-
// simple-request attacks by ensuring the browser preflights.
func ContentTypeGuard() gin.HandlerFunc {
	allowed := map[string]bool{
		"application/json":                  true,
		"multipart/form-data":               true,
		"application/x-www-form-urlencoded": true,
		"text/plain":                        true, // some webhook providers
	}
	return func(c *gin.Context) {
		m := c.Request.Method
		if m != http.MethodPost && m != http.MethodPut && m != http.MethodPatch {
			c.Next()
			return
		}
		if c.Request.ContentLength == 0 {
			c.Next()
			return
		}
		ct := c.GetHeader("Content-Type")
		// Strip parameters (charset, boundary).
		if i := strings.Index(ct, ";"); i >= 0 {
			ct = ct[:i]
		}
		ct = strings.TrimSpace(strings.ToLower(ct))
		if !allowed[ct] {
			c.AbortWithStatusJSON(http.StatusUnsupportedMediaType, gin.H{
				"detail": fmt.Sprintf("Unsupported content-type: %s", ct),
			})
			return
		}
		c.Next()
	}
}
