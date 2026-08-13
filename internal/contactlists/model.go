package contactlists

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type ContactList struct {
	database.Base
	OwnerID     string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	Description string `gorm:"type:text" json:"description"`
}

type ContactListMember struct {
	database.Base
	ListID    string `gorm:"type:varchar(36);index;not null" json:"list_id"`
	ContactID string `gorm:"type:varchar(36);index;not null" json:"contact_id"`
}
