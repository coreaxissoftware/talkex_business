package channels

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrUnknownKind = errors.New("unknown channel kind")
)

// CatalogItem is the marketing-side description that combines with the
// per-owner Config row to give the frontend everything it needs to draw
// a channel tile — implemented flag, display name, blurb.
type CatalogItem struct {
	Kind        Kind   `json:"kind"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Implemented bool   `json:"implemented"`
	Icon        string `json:"icon"`
}

var Catalog = []CatalogItem{
	{
		Kind:        KindTalkEx,
		DisplayName: "TalkEx",
		Description: "First-party channel powered by the TalkEx consumer app.",
		Implemented: false, // scaffolded; real send/receive comes later.
		Icon:        "radio",
	},
	{
		Kind:        KindWhatsApp,
		DisplayName: "WhatsApp Business",
		Description: "Cloud API integration for 2B+ WhatsApp users worldwide.",
		Implemented: false,
		Icon:        "message-circle",
	},
	{
		Kind:        KindTelegram,
		DisplayName: "Telegram",
		Description: "Bot API for one-to-many broadcasts and 2-way chat.",
		Implemented: false,
		Icon:        "send",
	},
	{
		Kind:        KindEmail,
		DisplayName: "Email",
		Description: "Transactional and campaign email via SMTP or provider.",
		Implemented: false,
		Icon:        "mail",
	},
	{
		Kind:        KindSMS,
		DisplayName: "SMS",
		Description: "Reach any mobile number via aggregator SMPP or REST.",
		Implemented: false,
		Icon:        "smartphone",
	},
	{
		Kind:        KindRCS,
		DisplayName: "RCS Business Messaging",
		Description: "Rich, branded messaging for Android users.",
		Implemented: false,
		Icon:        "message-square",
	},
}

func isKnown(k Kind) bool {
	for _, x := range AllKinds {
		if x == k {
			return true
		}
	}
	return false
}

// ListConfigs returns every Config row for the owner. The frontend joins
// this with the static Catalog to render its grid.
func ListConfigs(db *gorm.DB, ownerID string) ([]Config, error) {
	var out []Config
	err := db.Where("owner_id = ?", ownerID).Find(&out).Error
	return out, err
}

// SetEnabled is the toggle backing the /channels page switches. Creates
// the row if it doesn't yet exist.
type SetEnabledInput struct {
	Enabled bool                   `json:"enabled"`
	Config  map[string]interface{} `json:"config"`
}

func SetEnabled(db *gorm.DB, ownerID string, kind Kind, in *SetEnabledInput) (*Config, error) {
	if !isKnown(kind) {
		return nil, ErrUnknownKind
	}
	var c Config
	err := db.Where("owner_id = ? AND kind = ?", ownerID, kind).First(&c).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	created := errors.Is(err, gorm.ErrRecordNotFound)

	c.OwnerID = ownerID
	c.Kind = kind
	c.Enabled = in.Enabled
	if in.Config != nil {
		cfg, _ := json.Marshal(in.Config)
		c.Config = cfg
	}
	if in.Enabled {
		now := time.Now()
		c.VerifiedAt = &now
	} else {
		c.VerifiedAt = nil
	}

	if created {
		return &c, db.Create(&c).Error
	}
	return &c, db.Save(&c).Error
}
