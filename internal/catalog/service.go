package catalog

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound      = errors.New("product not found")
	ErrDuplicateSKU  = errors.New("retailer_id already exists for this owner")
)

// Create inserts a new product. Enforces (owner_id, retailer_id) uniqueness.
func Create(db *gorm.DB, ownerID string, p *Product) error {
	if p.RetailerID == "" || p.Name == "" || p.Price <= 0 {
		return errors.New("retailer_id, name, and price are required")
	}
	p.OwnerID = ownerID
	if p.Currency == "" {
		p.Currency = "INR"
	}
	if p.Availability == "" {
		p.Availability = AvailabilityInStock
	}
	// Duplicate check.
	var existing Product
	err := db.Where("owner_id = ? AND retailer_id = ?", ownerID, p.RetailerID).First(&existing).Error
	if err == nil {
		return ErrDuplicateSKU
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(p).Error
}

// List returns every product for the tenant, newest first. Optional
// category filter for the WhatsApp product-list picker.
func List(db *gorm.DB, ownerID, category string) ([]Product, error) {
	q := db.Where("owner_id = ?", ownerID)
	if category != "" {
		q = q.Where("category = ?", category)
	}
	var out []Product
	err := q.Order("created_at DESC").Find(&out).Error
	return out, err
}

// GetByID scoped to the caller's tenant.
func GetByID(db *gorm.DB, ownerID, id string) (*Product, error) {
	var p Product
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &p, err
}

// Update patches allowed fields. RetailerID is immutable — changing an
// SKU while Meta already has it mapped would silently orphan the row
// on Meta's side.
type UpdateInput struct {
	Name         *string  `json:"name"`
	Description  *string  `json:"description"`
	ImageURL     *string  `json:"image_url"`
	Price        *float64 `json:"price"`
	Currency     *string  `json:"currency"`
	Availability *string  `json:"availability"`
	URL          *string  `json:"url"`
	Category     *string  `json:"category"`
}

func Update(db *gorm.DB, p *Product, in *UpdateInput) error {
	if in.Name != nil {
		p.Name = *in.Name
	}
	if in.Description != nil {
		p.Description = *in.Description
	}
	if in.ImageURL != nil {
		p.ImageURL = *in.ImageURL
	}
	if in.Price != nil && *in.Price > 0 {
		p.Price = *in.Price
	}
	if in.Currency != nil && *in.Currency != "" {
		p.Currency = *in.Currency
	}
	if in.Availability != nil {
		p.Availability = *in.Availability
	}
	if in.URL != nil {
		p.URL = *in.URL
	}
	if in.Category != nil {
		p.Category = *in.Category
	}
	return db.Save(p).Error
}

func Delete(db *gorm.DB, p *Product) error {
	return db.Delete(p).Error
}
