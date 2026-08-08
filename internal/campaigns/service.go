package campaigns

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"github.com/coreaxissoftware/talkex_business/internal/templates"
)

var (
	ErrCampaignNotFound = errors.New("campaign not found")
	ErrTemplateNotFound = errors.New("template not found or not owned")
	ErrNoRecipients     = errors.New("campaign must have at least one contact")
	ErrInvalidStatus    = errors.New("campaign is not in a launchable state")
)

func List(db *gorm.DB, ownerID string) ([]Campaign, error) {
	var out []Campaign
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}

func GetByID(db *gorm.DB, ownerID, id string) (*Campaign, error) {
	var c Campaign
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrCampaignNotFound
	}
	return &c, err
}

type CreateInput struct {
	Name        string     `json:"name" binding:"required"`
	TemplateID  string     `json:"template_id" binding:"required"`
	ContactIDs  []string   `json:"contact_ids" binding:"required"`
	ScheduledAt *time.Time `json:"scheduled_at"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*Campaign, error) {
	if len(in.ContactIDs) == 0 {
		return nil, ErrNoRecipients
	}

	// Verify template belongs to this owner and grab its channel — the
	// campaign inherits the template's channel so we don't have to look
	// it up on every message send.
	var tpl templates.MessageTemplate
	if err := db.Where("id = ? AND owner_id = ?", in.TemplateID, ownerID).First(&tpl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, err
	}

	// Filter contact_ids to only those actually owned — an incoming id list
	// shouldn't be able to leak or target another tenant's contacts.
	var ownedIDs []string
	if err := db.Model(&contacts.Contact{}).
		Where("owner_id = ? AND id IN ?", ownerID, in.ContactIDs).
		Pluck("id", &ownedIDs).Error; err != nil {
		return nil, err
	}
	if len(ownedIDs) == 0 {
		return nil, ErrNoRecipients
	}

	ids, _ := json.Marshal(ownedIDs)
	status := StatusDraft
	if in.ScheduledAt != nil && in.ScheduledAt.After(time.Now()) {
		status = StatusScheduled
	}

	c := &Campaign{
		OwnerID:     ownerID,
		Name:        in.Name,
		TemplateID:  in.TemplateID,
		Channel:     tpl.Channel,
		Status:      status,
		ScheduledAt: in.ScheduledAt,
		ContactIDs:  datatypes.JSON(ids),
		TotalCount:  len(ownedIDs),
	}
	if err := db.Create(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// Launch marks a draft/scheduled campaign as running and stamps StartedAt.
// The actual per-message enqueue happens elsewhere (worker/channel connector);
// this is just the state transition + audit anchor.
func Launch(db *gorm.DB, c *Campaign) (*Campaign, error) {
	if c.Status != StatusDraft && c.Status != StatusScheduled {
		return nil, ErrInvalidStatus
	}
	now := time.Now()
	c.Status = StatusRunning
	c.StartedAt = &now
	if err := db.Save(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

// Cancel moves a non-terminal campaign into cancelled state.
func Cancel(db *gorm.DB, c *Campaign) (*Campaign, error) {
	if c.Status == StatusCompleted || c.Status == StatusFailed || c.Status == StatusCancelled {
		return nil, ErrInvalidStatus
	}
	c.Status = StatusCancelled
	if err := db.Save(c).Error; err != nil {
		return nil, err
	}
	return c, nil
}

func Delete(db *gorm.DB, c *Campaign) error {
	return db.Delete(c).Error
}
