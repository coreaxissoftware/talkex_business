package quality

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type EventType string

const (
	EventBlock  EventType = "block"
	EventReport EventType = "report"
	EventUnblock EventType = "unblock"
)

type Event struct {
	database.Base
	OwnerID   string    `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ContactID string    `gorm:"type:varchar(36);not null" json:"contact_id"`
	Channel   string    `gorm:"type:varchar(20);not null" json:"channel"`
	Type      EventType `gorm:"type:varchar(20);not null" json:"type"`
	Reason    *string   `gorm:"type:varchar(255)" json:"reason"`
}
