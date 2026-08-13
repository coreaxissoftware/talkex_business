package media

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type Media struct {
	database.Base
	OwnerID      string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Filename     string `gorm:"type:varchar(255);not null" json:"filename"`
	OriginalName string `gorm:"type:varchar(255);not null" json:"original_name"`
	MimeType     string `gorm:"type:varchar(100);not null" json:"mime_type"`
	Size         int64  `gorm:"not null" json:"size"`
	URL          string `gorm:"type:text;not null" json:"url"`
}
