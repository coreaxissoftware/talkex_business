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
	NotifCampaigns bool `json:"notif_campaigns"`
	NotifMessages  bool `json:"notif_messages"`
	NotifSystem    bool `json:"notif_system"`
	EmailDigest    bool `json:"email_digest"`
	Timezone       string `json:"timezone"`
	Language       string `json:"language"`
}
