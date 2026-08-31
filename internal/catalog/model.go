// Package catalog — WhatsApp Business commerce catalog.
//
// Meta's WhatsApp Cloud API supports interactive.product / product_list
// message types that let a business send a browsable product catalog
// inside a conversation. This package stores the catalog rows locally,
// syncs them to Meta's Commerce Catalog on demand, and exposes CRUD +
// send-product endpoints the messaging engine can call.
//
// Parity target: Wati "Catalog", Interakt "Shop", Zoko "Catalog".
package catalog

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// Availability values match Meta's Commerce Catalog spec verbatim so
// no translation layer is needed on sync.
const (
	AvailabilityInStock    = "in stock"
	AvailabilityOutOfStock = "out of stock"
	AvailabilityPreorder   = "preorder"
	AvailabilityDiscont    = "discontinued"
)

// Product is one row in the tenant's catalog. Metafield-style Extras
// carry Meta-specific fields we don't need first-class columns for
// (brand, google_product_category, gtin, etc.).
type Product struct {
	database.Base

	OwnerID string `gorm:"type:varchar(36);index;not null" json:"owner_id"`

	// RetailerID is the SKU / product-code the merchant uses. Unique
	// per owner and required by Meta (catalog.retailer_id).
	RetailerID string `gorm:"type:varchar(100);index;not null" json:"retailer_id"`

	Name        string  `gorm:"type:varchar(255);not null" json:"name"`
	Description string  `gorm:"type:text" json:"description"`
	ImageURL    string  `gorm:"type:varchar(1024)" json:"image_url"`
	Price       float64 `gorm:"not null" json:"price"`
	Currency    string  `gorm:"type:varchar(3);not null;default:'INR'" json:"currency"`

	Availability string `gorm:"type:varchar(20);not null;default:'in stock'" json:"availability"`

	// URL — the landing page on the merchant's own site.
	URL string `gorm:"type:varchar(1024)" json:"url"`

	// MetaCatalogID / MetaProductID — filled after a successful sync.
	// nil = not yet uploaded to Meta.
	MetaProductID *string `gorm:"type:varchar(64);index" json:"meta_product_id"`

	// Category is a free-text collection ("Sarees", "Kurta Sets"). Used
	// by the WhatsApp product-list message to group items by section.
	Category string `gorm:"type:varchar(100);index" json:"category"`
}

// TableName pins the plural — Gorm would already do this, but the catalog
// package has products.go elsewhere in the repo history and we want no
// ambiguity when reading raw SQL logs.
func (Product) TableName() string { return "catalog_products" }
