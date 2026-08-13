package users

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type UserRole string

const (
	RoleOwner     UserRole = "owner"
	RoleAdmin     UserRole = "admin"
	RoleAgent     UserRole = "agent"
	RoleDeveloper UserRole = "developer"
)

type User struct {
	database.Base
	Email              string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	HashedPassword     string    `gorm:"type:varchar(255);not null" json:"-"`
	FullName           string    `gorm:"type:varchar(255);not null" json:"full_name"`
	Role               UserRole  `gorm:"type:varchar(20);default:owner;not null" json:"role"`
	IsActive           bool      `gorm:"default:true;not null" json:"is_active"`
	IsBusinessVerified bool      `gorm:"default:false;not null" json:"is_business_verified"`
	BusinessCategory   *string   `gorm:"type:varchar(100)" json:"business_category"`
	QualityFlaggedAt   *time.Time `json:"quality_flagged_at"`
	TwoFactorSecret    *string    `gorm:"type:varchar(64)" json:"-"`
	TwoFactorEnabled   bool       `gorm:"default:false;not null" json:"two_factor_enabled"`
}
