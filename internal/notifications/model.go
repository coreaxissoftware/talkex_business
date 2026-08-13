// Package notifications — in-app bell-icon center for platform users.
// Other modules call Emit(...) to publish an event; the dashboard reads
// them via GET /notifications.
package notifications

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

const (
	TypeInfo    = "info"
	TypeSuccess = "success"
	TypeWarning = "warning"
	TypeError   = "error"
)

type Notification struct {
	database.Base
	OwnerID string     `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Type    string     `gorm:"type:varchar(20);not null;default:'info'" json:"type"`
	Title   string     `gorm:"type:varchar(255);not null" json:"title"`
	Body    string     `gorm:"type:text" json:"body"`
	Link    string     `gorm:"type:varchar(500)" json:"link"`
	ReadAt  *time.Time `json:"read_at"`
}
