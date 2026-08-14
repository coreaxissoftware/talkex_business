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
}
