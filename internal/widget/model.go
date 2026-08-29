// Package widget implements the public live-chat widget: an embeddable
// JS snippet that lets a business's website visitors start a chat that
// flows into the same Conversations inbox as WhatsApp/Telegram/etc.
//
// A visitor session is bound to a per-owner WidgetKey (public, safe to
// embed in HTML) and creates a Contact + Conversation on the fly. All
// downstream hooks (SSE, automation, flows, notifications) fire the
// same way they do for inbound WhatsApp because we route through
// conversations.RecordInbound.
package widget

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// Config is the per-owner widget configuration — public key, greeting,
// theme color. One row per owner (enforced by unique index on OwnerID).
type Config struct {
	database.Base
	OwnerID     string `gorm:"type:varchar(36);uniqueIndex;not null" json:"owner_id"`
	PublicKey   string `gorm:"type:varchar(64);uniqueIndex;not null" json:"public_key"`
	Enabled     bool   `gorm:"not null;default:true" json:"enabled"`
	Title       string `gorm:"type:varchar(120);default:'Chat with us'" json:"title"`
	Greeting    string `gorm:"type:text;default:'Hi! How can we help you today?'" json:"greeting"`
	ThemeColor  string `gorm:"type:varchar(20);default:'#2563eb'" json:"theme_color"`
}

// Session — one visitor's chat session, referenced from the widget by
// its ID; kept small so it can survive as a signed cookie/localStorage
// entry in the browser without server-side session store.
//
// The wire "session token" is the SessionID plus the ContactID, both
// unguessable UUIDs, so the widget doesn't need JWT infrastructure.
type Session struct {
	database.Base
	OwnerID        string  `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	ContactID      string  `gorm:"type:varchar(36);not null" json:"contact_id"`
	ConversationID string  `gorm:"type:varchar(36);not null" json:"conversation_id"`
	VisitorName    *string `gorm:"type:varchar(255)" json:"visitor_name"`
	VisitorEmail   *string `gorm:"type:varchar(255)" json:"visitor_email"`
	PageURL        *string `gorm:"type:varchar(512)" json:"page_url"`
	UserAgent      *string `gorm:"type:varchar(255)" json:"user_agent"`
}
