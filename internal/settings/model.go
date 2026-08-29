package settings

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

type UserSettings struct {
	database.Base
	OwnerID string         `gorm:"type:varchar(36);uniqueIndex;not null" json:"owner_id"`
	Prefs   datatypes.JSON `gorm:"type:json;default:'{}'" json:"prefs"`
}

type PrefsData struct {
	NotifCampaigns   bool    `json:"notif_campaigns"`
	NotifMessages    bool    `json:"notif_messages"`
	NotifSystem      bool    `json:"notif_system"`
	EmailDigest      bool    `json:"email_digest"`
	Timezone         string  `json:"timezone"`
	Language         string  `json:"language"`
	AutoPauseEnabled  bool    `json:"auto_pause_enabled"`
	MinBalance        float64 `json:"min_balance"`
	SandboxMode       bool    `json:"sandbox_mode"`
	ApprovalThreshold int     `json:"approval_threshold"`

	// Per-channel cost configuration (cost to business per message)
	CostWhatsapp  float64 `json:"cost_whatsapp"`
	CostSMS       float64 `json:"cost_sms"`
	CostTalkex    float64 `json:"cost_talkex"`
	CostTelegram  float64 `json:"cost_telegram"`
	CostEmail     float64 `json:"cost_email"`
	CostRCS       float64 `json:"cost_rcs"`
	CostInstagram float64 `json:"cost_instagram"`
	CostMessenger float64 `json:"cost_messenger"`
	// Sell price per message (what you charge the client)
	SellWhatsapp  float64 `json:"sell_whatsapp"`
	SellSMS       float64 `json:"sell_sms"`
	SellTalkex    float64 `json:"sell_talkex"`
	SellTelegram  float64 `json:"sell_telegram"`
	SellEmail     float64 `json:"sell_email"`
	SellRCS       float64 `json:"sell_rcs"`
	SellInstagram float64 `json:"sell_instagram"`
	SellMessenger float64 `json:"sell_messenger"`

	// Business hours + away message. When BusinessHoursEnabled is true
	// and an inbound message arrives outside the day/hour window, an
	// AwayMessage is auto-sent (once per contact per 24h). Days are
	// 1-indexed monday..sunday; times are 24h "HH:MM" strings in the
	// owner's Timezone (falls back to IST).
	BusinessHoursEnabled bool     `json:"business_hours_enabled"`
	BusinessDays         []int    `json:"business_days"`   // e.g. [1,2,3,4,5]
	BusinessOpenTime     string   `json:"business_open_time"`  // "09:00"
	BusinessCloseTime    string   `json:"business_close_time"` // "18:00"
	AwayMessage          string   `json:"away_message"`

	// SLA policy: alert when a conversation goes unanswered past the
	// FirstResponseMins threshold. 0 = disabled.
	SLAFirstResponseMins int `json:"sla_first_response_mins"`

	// AI auto-tag on inbound: adds a sentiment label (positive/neutral/
	// negative) to the conversation. Uses the shared ai package —
	// requires ANTHROPIC_API_KEY for real classification, dev heuristic
	// otherwise.
	AIAutoTagEnabled bool `json:"ai_auto_tag_enabled"`
}
