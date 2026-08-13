package messaging

import (
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

type DeadLetter struct {
	database.Base
	OwnerID    string  `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	CampaignID *string `gorm:"type:varchar(36)" json:"campaign_id"`
	ContactID  string  `gorm:"type:varchar(36);not null" json:"contact_id"`
	Channel    string  `gorm:"type:varchar(20);not null" json:"channel"`
	Body       string  `gorm:"type:text;not null" json:"body"`
	TemplateID *string `gorm:"type:varchar(36)" json:"template_id"`
	Error      string  `gorm:"type:text;not null" json:"error"`
	Attempts   int     `gorm:"not null" json:"attempts"`
	Resolved   bool    `gorm:"default:false;not null" json:"resolved"`
}

// MoveToDLQ takes a failed message that exhausted retries and creates a
// dead-letter entry for admin review. Called by the messaging worker.
func MoveToDLQ(db *gorm.DB, msg *QueuedMessage) error {
	errStr := ""
	if msg.Error != nil {
		errStr = *msg.Error
	}
	dl := &DeadLetter{
		OwnerID:    msg.OwnerID,
		CampaignID: msg.CampaignID,
		ContactID:  msg.ContactID,
		Channel:    msg.Channel,
		Body:       msg.Body,
		TemplateID: msg.TemplateID,
		Error:      errStr,
		Attempts:   msg.Attempts,
	}
	return db.Create(dl).Error
}

func ListDeadLetters(db *gorm.DB, ownerID string, includeResolved bool) ([]DeadLetter, error) {
	q := db.Where("owner_id = ?", ownerID)
	if !includeResolved {
		q = q.Where("resolved = ?", false)
	}
	var out []DeadLetter
	err := q.Order("created_at DESC").Limit(100).Find(&out).Error
	return out, err
}

// RetryDeadLetter re-enqueues a dead letter back into the message queue.
func RetryDeadLetter(db *gorm.DB, ownerID, dlID string) error {
	var dl DeadLetter
	if err := db.Where("id = ? AND owner_id = ?", dlID, ownerID).First(&dl).Error; err != nil {
		return err
	}
	_, err := Enqueue(db, &EnqueueInput{
		OwnerID:    dl.OwnerID,
		CampaignID: dl.CampaignID,
		ContactID:  dl.ContactID,
		Channel:    dl.Channel,
		Body:       dl.Body,
		TemplateID: dl.TemplateID,
		Priority:   PriorityMarketing,
	})
	if err != nil {
		return err
	}
	dl.Resolved = true
	return db.Save(&dl).Error
}

// DiscardDeadLetter marks a dead letter as resolved without retrying.
func DiscardDeadLetter(db *gorm.DB, ownerID, dlID string) error {
	return db.Model(&DeadLetter{}).
		Where("id = ? AND owner_id = ?", dlID, ownerID).
		Update("resolved", true).Error
}

// SweepToDLQ finds messages that exhausted retries and moves them to DLQ.
// Called periodically by the worker.
func SweepToDLQ(db *gorm.DB) int {
	var failed []QueuedMessage
	db.Where("status = ? AND attempts >= max_retries", shared.StatusFailed).Limit(100).Find(&failed)

	moved := 0
	for i := range failed {
		if err := MoveToDLQ(db, &failed[i]); err == nil {
			db.Delete(&failed[i])
			moved++
		}
	}
	return moved
}
