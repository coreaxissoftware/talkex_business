package customfields

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

const (
	TypeText     = "text"
	TypeNumber   = "number"
	TypeDate     = "date"
	TypeBoolean  = "boolean"
	TypeDropdown = "dropdown"
)

type FieldDefinition struct {
	database.Base
	OwnerID  string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name     string `gorm:"type:varchar(100);not null" json:"name"`
	Label    string `gorm:"type:varchar(255);not null" json:"label"`
	Type     string `gorm:"type:varchar(20);not null;default:'text'" json:"type"`
	Required bool   `gorm:"default:false" json:"required"`
	Options  string `gorm:"type:text" json:"options"`
}
