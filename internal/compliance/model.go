// Package compliance — DPDP Act 2023 compliance module.
// Handles consent records, Data Subject Access Requests (DSAR),
// right to erasure, and data processing records.
package compliance

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// ConsentRecord tracks explicit consent given by a data principal (contact).
type ConsentRecord struct {
	database.Base
	OwnerID     string     `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ContactID   string     `gorm:"type:varchar(36);index;not null" json:"contact_id"`
	Purpose     string     `gorm:"type:varchar(100);not null" json:"purpose"` // marketing, transactional, analytics
	Channel     string     `gorm:"type:varchar(50);not null" json:"channel"`
	ConsentGiven bool      `gorm:"default:false;not null" json:"consent_given"`
	ConsentedAt  *time.Time `json:"consented_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
	Source       string     `gorm:"type:varchar(100);not null" json:"source"` // opt-in form, import, API
	IPAddress    *string    `gorm:"type:varchar(45)" json:"ip_address"`
}

// DSARRequest — Data Subject Access Request per DPDP Act Section 11.
type DSARRequest struct {
	database.Base
	OwnerID    string     `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ContactID  string     `gorm:"type:varchar(36);index;not null" json:"contact_id"`
	Type       string     `gorm:"type:varchar(20);not null" json:"type"` // access, erasure, correction, portability
	Status     string     `gorm:"type:varchar(20);not null;default:'pending'" json:"status"` // pending, processing, completed, rejected
	Reason     *string    `gorm:"type:varchar(500)" json:"reason"`
	Response   *string    `gorm:"type:text" json:"response"`
	CompletedAt *time.Time `json:"completed_at"`
}

// ProcessingRecord — log of data processing activities per DPDP Act Section 8.
type ProcessingRecord struct {
	database.Base
	OwnerID       string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ContactID     string `gorm:"type:varchar(36);index" json:"contact_id"`
	Activity      string `gorm:"type:varchar(100);not null" json:"activity"` // message_sent, data_export, profile_update, data_deletion
	Purpose       string `gorm:"type:varchar(100);not null" json:"purpose"`  // marketing, transactional, analytics, compliance
	DataCategory  string `gorm:"type:varchar(100);not null" json:"data_category"` // personal, contact, communication, financial
	LegalBasis    string `gorm:"type:varchar(50);not null" json:"legal_basis"` // consent, legitimate_interest, legal_obligation, contract
	Details       string `gorm:"type:text" json:"details"`
}
