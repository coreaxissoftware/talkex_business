// Package whitelabel — per-tenant / per-org branding overrides so a
// reseller can ship the dashboard as their own product without a code
// change: swap the logo, colours, product name, from-email, and (at
// the reverse-proxy layer) the custom domain.
//
// Model is scoped to owner_id so a solo tenant can rebrand their own
// login page too — a reseller inherits the same rows, keyed by their
// tenants' owner_ids.
package whitelabel

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// Branding holds everything the frontend needs to render a login /
// signup / dashboard shell under the tenant's own identity.
type Branding struct {
	database.Base
	OwnerID string `gorm:"type:varchar(36);uniqueIndex;not null" json:"owner_id"`

	// Product identity
	BrandName   string `gorm:"type:varchar(80)" json:"brand_name"`
	Tagline     string `gorm:"type:varchar(200)" json:"tagline"`

	// Visual — hex colour, e.g. "#0EA5A0". Falls back to jade default.
	PrimaryColor string `gorm:"type:varchar(9)" json:"primary_color"`
	AccentColor  string `gorm:"type:varchar(9)" json:"accent_color"`

	// Asset URLs — reseller uploads to their own CDN and pastes here.
	LogoURL     string `gorm:"type:varchar(500)" json:"logo_url"`
	LogoDarkURL string `gorm:"type:varchar(500)" json:"logo_dark_url"`
	FaviconURL  string `gorm:"type:varchar(500)" json:"favicon_url"`

	// Domain — the CNAME the reseller has pointed at our load balancer,
	// e.g. "app.reseller.co.in". Used by the public /branding endpoint
	// to look up which brand to serve when Host matches.
	CustomDomain string `gorm:"type:varchar(200);index" json:"custom_domain"`

	// Transactional email identity
	FromEmail   string `gorm:"type:varchar(255)" json:"from_email"`
	SupportURL  string `gorm:"type:varchar(500)" json:"support_url"`

	// Legal — override the footer legal links the marketing site links to.
	PrivacyURL string `gorm:"type:varchar(500)" json:"privacy_url"`
	TermsURL   string `gorm:"type:varchar(500)" json:"terms_url"`

	// PoweredBy — when false, the "Powered by TalkEx" chip is hidden.
	// Enforce plan-tier gating in the handler.
	HidePoweredBy bool `gorm:"default:false" json:"hide_powered_by"`
}
