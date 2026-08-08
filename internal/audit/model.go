// Package audit records every API request as a structured log entry —
// method, path, status code, latency, actor, and (for failures) the
// response body — so failed requests and user activity can be inspected
// from the dashboard instead of grepping server stdout.
package audit

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// LogEntry is one recorded HTTP request/response pair.
type LogEntry struct {
	database.Base
	UserID     *string `gorm:"type:varchar(36);index" json:"user_id"`
	UserEmail  *string `gorm:"type:varchar(255)" json:"user_email"`
	Method     string  `gorm:"type:varchar(10);not null" json:"method"`
	Path       string  `gorm:"type:varchar(500);not null;index" json:"path"`
	StatusCode int     `gorm:"not null;index" json:"status_code"`
	Success    bool    `gorm:"not null;index" json:"success"`
	LatencyMs  int64   `gorm:"not null" json:"latency_ms"`
	ClientIP   string  `gorm:"type:varchar(45)" json:"client_ip"`
	ErrorBody  *string `gorm:"type:text" json:"error_body,omitempty"`
}
