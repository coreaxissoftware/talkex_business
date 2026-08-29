package canned

import (
	"errors"

	"gorm.io/gorm"
)

var (
	ErrNotFound    = errors.New("canned response not found")
	ErrDuplicate   = errors.New("shortcut already exists")
)

type CreateInput struct {
	Shortcut string `json:"shortcut" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Category string `json:"category"`
}

type UpdateInput struct {
	Shortcut *string `json:"shortcut"`
	Title    *string `json:"title"`
	Body     *string `json:"body"`
	Category *string `json:"category"`
}

func List(db *gorm.DB, ownerID string) ([]Response, error) {
	var out []Response
	err := db.Where("owner_id = ?", ownerID).Order("usage_count DESC, shortcut ASC").Find(&out).Error
	return out, err
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*Response, error) {
	var existing Response
	if err := db.Where("owner_id = ? AND shortcut = ?", ownerID, in.Shortcut).First(&existing).Error; err == nil {
		return nil, ErrDuplicate
	}
	cat := in.Category
	if cat == "" {
		cat = "general"
	}
	r := &Response{
		OwnerID:  ownerID,
		Shortcut: in.Shortcut,
		Title:    in.Title,
		Body:     in.Body,
		Category: cat,
	}
	if err := db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func GetByID(db *gorm.DB, ownerID, id string) (*Response, error) {
	var r Response
	if err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&r).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &r, nil
}

func Update(db *gorm.DB, r *Response, in *UpdateInput) (*Response, error) {
	if in.Shortcut != nil {
		r.Shortcut = *in.Shortcut
	}
	if in.Title != nil {
		r.Title = *in.Title
	}
	if in.Body != nil {
		r.Body = *in.Body
	}
	if in.Category != nil {
		r.Category = *in.Category
	}
	if err := db.Save(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func Delete(db *gorm.DB, r *Response) error {
	return db.Delete(r).Error
}

// BumpUsage increments the usage counter — called when an agent inserts
// the response into a reply. Best-effort; a failure here shouldn't block
// the send.
func BumpUsage(db *gorm.DB, id string) {
	db.Model(&Response{}).Where("id = ?", id).UpdateColumn("usage_count", gorm.Expr("usage_count + 1"))
}
