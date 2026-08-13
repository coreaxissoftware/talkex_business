package customers

import (
	"errors"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("customer profile not found")

func GetByOwner(db *gorm.DB, ownerID string) (*Customer, error) {
	var c Customer
	err := db.Where("owner_id = ?", ownerID).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	return &c, err
}

type UpsertInput struct {
	BusinessName     string  `json:"business_name" binding:"required"`
	BusinessCategory string  `json:"business_category"`
	GSTIN            *string `json:"gstin"`
	Website          *string `json:"website"`
	Address          *string `json:"address"`
	City             *string `json:"city"`
	State            *string `json:"state"`
	Country          string  `json:"country"`
	Phone            *string `json:"phone"`
	LogoURL          *string `json:"logo_url"`
}

func Upsert(db *gorm.DB, ownerID string, in *UpsertInput) (*Customer, error) {
	c, err := GetByOwner(db, ownerID)
	if err == ErrNotFound {
		c = &Customer{OwnerID: ownerID}
	} else if err != nil {
		return nil, err
	}

	c.BusinessName = in.BusinessName
	c.BusinessCategory = in.BusinessCategory
	c.GSTIN = in.GSTIN
	c.Website = in.Website
	c.Address = in.Address
	c.City = in.City
	c.State = in.State
	if in.Country != "" {
		c.Country = in.Country
	}
	c.Phone = in.Phone
	c.LogoURL = in.LogoURL

	if c.ID == "" {
		return c, db.Create(c).Error
	}
	return c, db.Save(c).Error
}
