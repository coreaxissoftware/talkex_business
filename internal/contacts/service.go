package contacts

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

var ErrContactNotFound = errors.New("contact not found")

func List(db *gorm.DB, ownerID string) ([]Contact, error) {
	var contacts []Contact
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&contacts).Error
	return contacts, err
}

func GetByID(db *gorm.DB, ownerID, contactID string) (*Contact, error) {
	var c Contact
	err := db.Where("id = ? AND owner_id = ?", contactID, ownerID).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrContactNotFound
	}
	return &c, err
}

type CreateInput struct {
	PhoneNumber  string      `json:"phone_number" binding:"required"`
	Name         *string     `json:"name"`
	Email        *string     `json:"email"`
	Tags         []string    `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*Contact, error) {
	tagsJSON, _ := json.Marshal(in.Tags)
	fieldsJSON, _ := json.Marshal(in.CustomFields)
	if in.Tags == nil {
		tagsJSON = []byte("[]")
	}
	if in.CustomFields == nil {
		fieldsJSON = []byte("{}")
	}

	c := &Contact{
		OwnerID:      ownerID,
		PhoneNumber:  in.PhoneNumber,
		Name:         in.Name,
		Email:        in.Email,
		Tags:         tagsJSON,
		CustomFields: fieldsJSON,
	}
	if err := db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

type UpdateInput struct {
	Name         *string                `json:"name"`
	Email        *string                `json:"email"`
	Tags         *[]string              `json:"tags"`
	CustomFields map[string]interface{} `json:"custom_fields"`
}

func Update(db *gorm.DB, c *Contact, in *UpdateInput) (*Contact, error) {
	if in.Name != nil {
		c.Name = in.Name
	}
	if in.Email != nil {
		c.Email = in.Email
	}
	if in.Tags != nil {
		tagsJSON, _ := json.Marshal(*in.Tags)
		c.Tags = tagsJSON
	}
	if in.CustomFields != nil {
		fieldsJSON, _ := json.Marshal(in.CustomFields)
		c.CustomFields = fieldsJSON
	}
	if err := db.Save(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func Delete(db *gorm.DB, c *Contact) error {
	return db.Delete(c).Error
}
