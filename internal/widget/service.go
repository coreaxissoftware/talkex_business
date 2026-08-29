package widget

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrConfigNotFound = errors.New("widget config not found")
	ErrConfigDisabled = errors.New("widget disabled")
	ErrSessionNotFound = errors.New("widget session not found")
)

// generatePublicKey returns a URL-safe hex string used as the public
// widget identifier. 24 bytes = 48 hex chars, unguessable.
func generatePublicKey() string {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return "tw_" + hex.EncodeToString(b)
}

// GetOrCreateConfig returns the widget config for an owner, creating
// one with a fresh public key on first call.
func GetOrCreateConfig(db *gorm.DB, ownerID string) (*Config, error) {
	var c Config
	err := db.Where("owner_id = ?", ownerID).First(&c).Error
	if err == nil {
		return &c, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	c = Config{
		OwnerID:    ownerID,
		PublicKey:  generatePublicKey(),
		Enabled:    true,
		Title:      "Chat with us",
		Greeting:   "Hi! How can we help you today?",
		ThemeColor: "#2563eb",
	}
	if err := db.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateInput is the PATCH body from the Settings page.
type UpdateInput struct {
	Enabled    *bool   `json:"enabled"`
	Title      *string `json:"title"`
	Greeting   *string `json:"greeting"`
	ThemeColor *string `json:"theme_color"`
}

func UpdateConfig(db *gorm.DB, c *Config, in *UpdateInput) (*Config, error) {
	if in.Enabled != nil {
		c.Enabled = *in.Enabled
	}
	if in.Title != nil {
		c.Title = *in.Title
	}
	if in.Greeting != nil {
		c.Greeting = *in.Greeting
	}
	if in.ThemeColor != nil {
		c.ThemeColor = *in.ThemeColor
	}
	if err := db.Save(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// RotateKey generates a new PublicKey and invalidates the old one so
// the owner can revoke embeds that leaked.
func RotateKey(db *gorm.DB, c *Config) (*Config, error) {
	c.PublicKey = generatePublicKey()
	if err := db.Save(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// FindConfigByKey looks up the public config a widget snippet embeds.
func FindConfigByKey(db *gorm.DB, publicKey string) (*Config, error) {
	var c Config
	err := db.Where("public_key = ?", strings.TrimSpace(publicKey)).First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrConfigNotFound
		}
		return nil, err
	}
	if !c.Enabled {
		return nil, ErrConfigDisabled
	}
	return &c, nil
}

// GetSession returns a session by ID scoped to owner (owner is
// derived from the public key on the request).
func GetSession(db *gorm.DB, ownerID, sessionID string) (*Session, error) {
	var s Session
	err := db.Where("id = ? AND owner_id = ?", sessionID, ownerID).First(&s).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}
	return &s, nil
}
