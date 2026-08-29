// Package canned provides quick-reply message snippets that agents can
// insert into conversation replies with a shortcut (e.g. "/greeting").
// Reduces typing for common responses like greetings, shipping updates,
// office-hours notices — every support team needs these.
package canned

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type Response struct {
	database.Base
	OwnerID   string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Shortcut  string `gorm:"type:varchar(64);index;not null" json:"shortcut"`
	Title     string `gorm:"type:varchar(255);not null" json:"title"`
	Body      string `gorm:"type:text;not null" json:"body"`
	Category  string `gorm:"type:varchar(64);default:'general'" json:"category"`
	UsageCount int64 `gorm:"default:0" json:"usage_count"`
}
