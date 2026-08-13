package customers

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type VerificationStatus string

const (
	StatusPending  VerificationStatus = "pending"
	StatusVerified VerificationStatus = "verified"
	StatusRejected VerificationStatus = "rejected"
)

type Customer struct {
	database.Base
	OwnerID            string             `gorm:"type:varchar(36);uniqueIndex;not null" json:"owner_id"`
	BusinessName       string             `gorm:"type:varchar(255);not null" json:"business_name"`
	BusinessCategory   string             `gorm:"type:varchar(100)" json:"business_category"`
	GSTIN              *string            `gorm:"type:varchar(15)" json:"gstin"`
	Website            *string            `gorm:"type:varchar(255)" json:"website"`
	Address            *string            `gorm:"type:text" json:"address"`
	City               *string            `gorm:"type:varchar(100)" json:"city"`
	State              *string            `gorm:"type:varchar(100)" json:"state"`
	Country            string             `gorm:"type:varchar(2);default:'IN'" json:"country"`
	Phone              *string            `gorm:"type:varchar(20)" json:"phone"`
	LogoURL            *string            `gorm:"type:varchar(500)" json:"logo_url"`
	VerificationStatus VerificationStatus `gorm:"type:varchar(20);default:'pending';not null" json:"verification_status"`
	VerificationNote   *string            `gorm:"type:text" json:"verification_note"`
}
