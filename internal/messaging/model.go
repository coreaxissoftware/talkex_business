package messaging

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type QueuedMessage struct {
	database.Base
	OwnerID    string               `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	CampaignID *string              `gorm:"type:varchar(36);index" json:"campaign_id"`
	ContactID  string               `gorm:"type:varchar(36);not null" json:"contact_id"`
	Channel    string               `gorm:"type:varchar(20);not null" json:"channel"`
	Body       string               `gorm:"type:text;not null" json:"body"`
	TemplateID *string              `gorm:"type:varchar(36)" json:"template_id"`
	MediaURL   *string              `gorm:"type:varchar(500)" json:"media_url"`
	Status     shared.DeliveryStatus `gorm:"type:varchar(20);default:'queued';index;not null" json:"status"`
	ExternalID *string              `gorm:"type:varchar(255)" json:"external_id"`
	Error      *string              `gorm:"type:text" json:"error"`
	Attempts   int                  `gorm:"default:0;not null" json:"attempts"`
	MaxRetries int                  `gorm:"default:3;not null" json:"max_retries"`
	NextRetry  *time.Time           `gorm:"index" json:"next_retry"`
	SentAt          *time.Time           `json:"sent_at"`
	Priority        int                  `gorm:"default:10;not null;index" json:"priority"`
	FallbackTried   bool                 `gorm:"default:false;not null" json:"fallback_tried"`
	OriginalChannel *string              `gorm:"type:varchar(50)" json:"original_channel"`
	CostPerMessage  float64              `gorm:"default:0" json:"cost_per_message"`
	SellPrice       float64              `gorm:"default:0" json:"sell_price"`
}

const (
	PriorityOTP           = 1
	PriorityTransactional = 5
	PriorityMarketing     = 10
)
