// Package paylinks — Razorpay Quick Link generation + in-chat send.
//
// A merchant clicks "Send payment link" inside a conversation, picks
// an amount and reason, and TalkEx mints a Razorpay Quick Link, stores
// the row, and posts the URL to the customer on whichever channel the
// conversation is running on. When Razorpay's webhook confirms the
// payment we flip the row to "paid" and post a "thanks!" auto-message.
//
// Parity target: Interakt "Pay link", DoubleTick "Razorpay link".
package paylinks

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// PayLink lifecycle:
//
//   created → sent → paid       (happy path)
//                  ↘ expired    (customer never paid; TTL elapsed)
//                  ↘ cancelled  (merchant cancelled from dashboard)
const (
	StatusCreated   = "created"
	StatusSent      = "sent"
	StatusPaid      = "paid"
	StatusExpired   = "expired"
	StatusCancelled = "cancelled"
)

type PayLink struct {
	database.Base
	OwnerID        string  `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ContactID      string  `gorm:"type:varchar(36);index;not null" json:"contact_id"`
	ConversationID string  `gorm:"type:varchar(36);index" json:"conversation_id,omitempty"`
	AmountPaise    int64   `gorm:"not null" json:"amount_paise"` // Razorpay wants paise
	Currency       string  `gorm:"type:varchar(3);not null;default:'INR'" json:"currency"`
	Description    string  `gorm:"type:varchar(500)" json:"description"`
	Status         string  `gorm:"type:varchar(20);not null;default:'created';index" json:"status"`

	// RazorpayID + URL come back from Razorpay's /payment_links create.
	// In dev/simulation mode we mint a placeholder URL and set
	// Simulated=true so the analytics dashboard can filter them out.
	RazorpayID string `gorm:"type:varchar(64);index" json:"razorpay_id,omitempty"`
	URL        string `gorm:"type:varchar(500);not null" json:"url"`
	Simulated  bool   `gorm:"default:false" json:"simulated"`

	SentAt    *time.Time `json:"sent_at,omitempty"`
	PaidAt    *time.Time `json:"paid_at,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
