package templates

import (
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

var ErrTemplateNotFound = errors.New("template not found")

func List(db *gorm.DB, ownerID string) ([]MessageTemplate, error) {
	var tpls []MessageTemplate
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&tpls).Error
	return tpls, err
}

func GetByID(db *gorm.DB, ownerID, templateID string) (*MessageTemplate, error) {
	var t MessageTemplate
	err := db.Where("id = ? AND owner_id = ?", templateID, ownerID).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTemplateNotFound
	}
	return &t, err
}

type CreateInput struct {
	Name      string           `json:"name" binding:"required"`
	Category  TemplateCategory `json:"category" binding:"required"`
	Channel   string           `json:"channel" binding:"required"`
	Body      string           `json:"body" binding:"required"`
	Variables []string         `json:"variables"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*MessageTemplate, error) {
	varsJSON, _ := json.Marshal(in.Variables)
	if in.Variables == nil {
		varsJSON = []byte("[]")
	}

	t := &MessageTemplate{
		OwnerID:   ownerID,
		Name:      in.Name,
		Category:  in.Category,
		Channel:   in.Channel,
		Body:      in.Body,
		Variables: varsJSON,
	}
	if err := db.Create(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

type UpdateInput struct {
	Name      *string         `json:"name"`
	Body      *string         `json:"body"`
	Variables *[]string       `json:"variables"`
	Status    *TemplateStatus `json:"status"`
}

func Update(db *gorm.DB, t *MessageTemplate, in *UpdateInput) (*MessageTemplate, error) {
	if in.Name != nil {
		t.Name = *in.Name
	}
	if in.Body != nil {
		t.Body = *in.Body
	}
	if in.Variables != nil {
		varsJSON, _ := json.Marshal(*in.Variables)
		t.Variables = varsJSON
	}
	if in.Status != nil {
		t.Status = *in.Status
	}
	if err := db.Save(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func Delete(db *gorm.DB, t *MessageTemplate) error {
	return db.Delete(t).Error
}
