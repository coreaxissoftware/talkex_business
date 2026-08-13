package support

import (
	"errors"

	"gorm.io/gorm"
)

var ErrTicketNotFound = errors.New("ticket not found")

func List(db *gorm.DB, ownerID string) ([]Ticket, error) {
	var out []Ticket
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}

type CreateInput struct {
	Subject  string `json:"subject" binding:"required"`
	Body     string `json:"body" binding:"required"`
	Priority string `json:"priority"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*Ticket, error) {
	pri := in.Priority
	if pri == "" {
		pri = PriorityNormal
	}
	t := &Ticket{
		OwnerID:  ownerID,
		Subject:  in.Subject,
		Body:     in.Body,
		Priority: pri,
		Status:   TicketStatusOpen,
	}
	return t, db.Create(t).Error
}
