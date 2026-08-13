package team

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

const (
	RoleAdmin  = "admin"
	RoleAgent  = "agent"
	RoleViewer = "viewer"

	StatusPending = "pending"
	StatusActive  = "active"
)

type Member struct {
	database.Base
	OwnerID string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Email   string `gorm:"type:varchar(255);not null" json:"email"`
	Name    string `gorm:"type:varchar(255)" json:"name"`
	Role    string `gorm:"type:varchar(20);not null;default:'agent'" json:"role"`
	Status  string `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	UserID  string `gorm:"type:varchar(36)" json:"user_id"`
}
