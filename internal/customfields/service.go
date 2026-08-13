package customfields

import (
	"errors"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrFieldNotFound = errors.New("custom field not found")
	ErrDuplicateName = errors.New("a field with this name already exists")
	ErrInvalidType   = errors.New("invalid field type")
)

var validTypes = map[string]bool{
	TypeText: true, TypeNumber: true, TypeDate: true,
	TypeBoolean: true, TypeDropdown: true,
}

func List(db *gorm.DB, ownerID string) ([]FieldDefinition, error) {
	var out []FieldDefinition
	err := db.Where("owner_id = ?", ownerID).Order("created_at ASC").Find(&out).Error
	return out, err
}

func GetByID(db *gorm.DB, ownerID, id string) (*FieldDefinition, error) {
	var f FieldDefinition
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&f).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrFieldNotFound
	}
	return &f, err
}

type CreateInput struct {
	Name     string `json:"name" binding:"required"`
	Label    string `json:"label" binding:"required"`
	Type     string `json:"type" binding:"required"`
	Required bool   `json:"required"`
	Options  string `json:"options"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*FieldDefinition, error) {
	if !validTypes[in.Type] {
		return nil, ErrInvalidType
	}

	slug := strings.ToLower(strings.ReplaceAll(strings.TrimSpace(in.Name), " ", "_"))
	var count int64
	db.Model(&FieldDefinition{}).Where("owner_id = ? AND name = ?", ownerID, slug).Count(&count)
	if count > 0 {
		return nil, ErrDuplicateName
	}

	f := &FieldDefinition{
		OwnerID:  ownerID,
		Name:     slug,
		Label:    in.Label,
		Type:     in.Type,
		Required: in.Required,
		Options:  in.Options,
	}
	if err := db.Create(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

type UpdateInput struct {
	Label    *string `json:"label"`
	Required *bool   `json:"required"`
	Options  *string `json:"options"`
}

func Update(db *gorm.DB, f *FieldDefinition, in *UpdateInput) (*FieldDefinition, error) {
	if in.Label != nil {
		f.Label = *in.Label
	}
	if in.Required != nil {
		f.Required = *in.Required
	}
	if in.Options != nil {
		f.Options = *in.Options
	}
	return f, db.Save(f).Error
}

func Delete(db *gorm.DB, f *FieldDefinition) error {
	return db.Delete(f).Error
}
