// Package observability — Sentry error tracking wiring.
//
// Optional: when SENTRY_DSN is unset, InitSentry is a no-op and every
// helper degrades to a plain log line. This lets dev + CI runs skip
// the dependency entirely while prod gets full stack captures.
//
// Rather than pulling the Sentry Go SDK (which drags ~2 MB into the
// binary), this file speaks Sentry's tiny envelope protocol directly.
// We only send events (no transactions, no profiling), which needs
// nothing more than a single POST per capture.
package observability

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

var (
	sentryOnce  sync.Once
	sentryDSN   string
	sentryURL   string
	sentryAuth  string
	sentryReady bool
	sentryEnv   string
)

// InitSentry parses SENTRY_DSN once and caches the URL + auth header
// used by every capture. Safe to call multiple times.
func InitSentry() {
	sentryOnce.Do(func() {
		dsn := os.Getenv("SENTRY_DSN")
		if dsn == "" {
			return
		}
		// A Sentry DSN looks like:
		//   https://<publicKey>@o<org>.ingest.sentry.io/<projectId>
		// We rewrite it to the envelope endpoint used by /store/.
		parts := strings.SplitN(dsn, "@", 2)
		if len(parts) != 2 {
			log.Printf("sentry: invalid DSN")
			return
		}
		scheme := parts[0][:strings.Index(parts[0], "//")+2]
		publicKey := parts[0][strings.Index(parts[0], "//")+2:]
		host, projectID, ok := splitLast(parts[1], "/")
		if !ok {
			log.Printf("sentry: invalid DSN (no project id)")
			return
		}
		sentryDSN = dsn
		sentryURL = fmt.Sprintf("%s%s/api/%s/store/", scheme, host, projectID)
		sentryAuth = fmt.Sprintf(
			"Sentry sentry_version=7, sentry_client=talkex-go/1.0, sentry_key=%s",
			publicKey,
		)
		sentryEnv = config.Get().Environment
		sentryReady = true
		log.Printf("sentry: initialised (project %s, env %s)", projectID, sentryEnv)
	})
}

// CaptureError sends an error event to Sentry with the current stack
// trace. Fire-and-forget: never blocks the caller for more than 4s.
func CaptureError(err error, extra map[string]interface{}) {
	if err == nil {
		return
	}
	if !sentryReady {
		log.Printf("[error] %v", err)
		return
	}
	stack := string(debug.Stack())
	payload := map[string]interface{}{
		"event_id":    strings.ReplaceAll(uuid.New().String(), "-", ""),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"level":       "error",
		"logger":      "talkex",
		"platform":    "go",
		"environment": sentryEnv,
		"server_name": hostname(),
		"exception": map[string]interface{}{
			"values": []map[string]interface{}{{
				"type":  fmt.Sprintf("%T", err),
				"value": err.Error(),
				"stacktrace": map[string]interface{}{
					"frames": parseStackFrames(stack),
				},
			}},
		},
		"contexts": map[string]interface{}{
			"runtime": map[string]interface{}{
				"name":    "go",
				"version": runtime.Version(),
			},
		},
	}
	if extra != nil {
		payload["extra"] = extra
	}
	go sendEnvelope(payload)
}

// CaptureMessage sends a plain string message. Useful for warn-level
// events that aren't errors (rate-limit hit, degraded mode, etc.).
func CaptureMessage(msg, level string, extra map[string]interface{}) {
	if !sentryReady {
		log.Printf("[%s] %s", level, msg)
		return
	}
	payload := map[string]interface{}{
		"event_id":    strings.ReplaceAll(uuid.New().String(), "-", ""),
		"timestamp":   time.Now().UTC().Format(time.RFC3339),
		"level":       level,
		"logger":      "talkex",
		"platform":    "go",
		"environment": sentryEnv,
		"server_name": hostname(),
		"message":     map[string]string{"formatted": msg},
	}
	if extra != nil {
		payload["extra"] = extra
	}
	go sendEnvelope(payload)
}

func sendEnvelope(payload map[string]interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		return
	}
	req, err := http.NewRequest(http.MethodPost, sentryURL, bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Sentry-Auth", sentryAuth)

	client := &http.Client{Timeout: 4 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return
	}
	res.Body.Close()
}

func parseStackFrames(stack string) []map[string]interface{} {
	lines := strings.Split(stack, "\n")
	frames := make([]map[string]interface{}, 0, len(lines)/2)
	// runtime/debug.Stack lines come in pairs: function name, then file:line.
	for i := 0; i < len(lines)-1; i += 2 {
		fn := strings.TrimSpace(lines[i])
		loc := strings.TrimSpace(lines[i+1])
		if fn == "" || !strings.Contains(loc, ":") {
			continue
		}
		file, lineStr, _ := splitLast(loc, ":")
		// strip trailing " +0x..." offset if present
		if sp := strings.Index(lineStr, " "); sp >= 0 {
			lineStr = lineStr[:sp]
		}
		frames = append(frames, map[string]interface{}{
			"function": fn,
			"filename": file,
			"lineno":   lineStr,
		})
	}
	// Sentry expects most-recent-first, but debug.Stack returns top-down;
	// reverse.
	for i, j := 0, len(frames)-1; i < j; i, j = i+1, j-1 {
		frames[i], frames[j] = frames[j], frames[i]
	}
	return frames
}

func splitLast(s, sep string) (string, string, bool) {
	i := strings.LastIndex(s, sep)
	if i < 0 {
		return s, "", false
	}
	return s[:i], s[i+1:], true
}

func hostname() string {
	h, _ := os.Hostname()
	if h == "" {
		return "unknown"
	}
	return h
}
