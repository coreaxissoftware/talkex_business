// Package conversations — 2-way inbox per (owner, contact, channel).
// A conversation is the container; messages are the individual events.
//
// The 24-hour customer-service window (CONTEXT.md) is derived here:
// LastInboundAt drives IsWindowOpen — outbound sends can be free-form
// while the window is open; once closed they must use an approved
// template and are billed as marketing/utility/authentication.
package conversations

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

const WindowHours = 24

// Direction of a Message.
const (
	Inbound  = "inbound"
	Outbound = "outbound"
)

// Message delivery status.
const (
	MsgStatusQueued    = "queued"
	MsgStatusSent      = "sent"
	MsgStatusDelivered = "delivered"
	MsgStatusRead      = "read"
	MsgStatusFailed    = "failed"
)

// Conversation is one running thread with a contact on one channel.
// OwnerID + ContactID + Channel is a natural unique key; a separate id
// exists so the API/frontend can address a conversation without the
// caller having to know all three parts.
type Conversation struct {
	database.Base
	OwnerID        string     `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ContactID      string     `gorm:"type:varchar(36);index;not null" json:"contact_id"`
	Channel        string     `gorm:"type:varchar(50);not null" json:"channel"`
	LastInboundAt  *time.Time `json:"last_inbound_at"`
	LastOutboundAt *time.Time `json:"last_outbound_at"`
	LastMessageAt  *time.Time `json:"last_message_at"`
	UnreadCount    int        `gorm:"default:0" json:"unread_count"`
	Labels         string     `gorm:"type:text;default:'[]'" json:"labels"`
	AssignedTo     *string    `gorm:"type:varchar(36)" json:"assigned_to"`
	AssignedName   *string    `gorm:"type:varchar(255)" json:"assigned_name"`
}

// IsWindowOpen — a free-form outbound send is allowed only inside the
// 24-hour window that starts on the most recent inbound message.
func (c *Conversation) IsWindowOpen() bool {
	if c.LastInboundAt == nil {
		return false
	}
	return time.Since(*c.LastInboundAt) < WindowHours*time.Hour
}

// Message is one event in a conversation.
type Message struct {
	database.Base
	ConversationID string     `gorm:"type:varchar(36);index;not null" json:"conversation_id"`
	Direction      string     `gorm:"type:varchar(10);not null" json:"direction"`
	Body           string     `gorm:"type:text;not null" json:"body"`
	Status         string     `gorm:"type:varchar(15);not null;default:'queued'" json:"status"`
	TemplateID     *string    `gorm:"type:varchar(36)" json:"template_id,omitempty"`
	DeliveredAt    *time.Time `json:"delivered_at,omitempty"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	ErrorReason    *string    `gorm:"type:varchar(500)" json:"error_reason,omitempty"`
}
