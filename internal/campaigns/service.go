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

// SendFunc lets main.go inject the conversations.SendOutbound call
// without campaigns importing conversations (which would create a cycle
// via automation → conversations → campaigns).
type SendFunc func(ownerID, contactID, channel, body string, templateID *string) error

// CompletionHook fires when a campaign transitions to completed/failed —
// wired from main.go to fan out webhooks + notifications.
type CompletionHook func(ownerID string, c *Campaign)

var (
	sender          SendFunc
	onComplete      CompletionHook
)

func RegisterSender(f SendFunc)              { sender = f }
func RegisterCompletionHook(h CompletionHook) { onComplete = h }

// Launch flips the state to running and, if a sender is registered,
// fans out the per-recipient sends in a background goroutine so the HTTP
// call returns immediately. Roll-up counters are updated after each send.
// If no sender is wired (dev environment without a connector), the
// campaign still transitions correctly so the UI reflects it.
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

	// Fan out sends off the request path. Recover per-goroutine so a
	// broken sender can't crash the process.
	go func(campaign Campaign) {
		defer func() { _ = recover() }()
		run(db, &campaign)
	}(*c)

	return c, nil
}

// run executes the sends synchronously in a background goroutine. It
// keeps its own db/campaign copy so the caller's pointer isn't mutated
// across goroutines. If the campaign is cancelled mid-run, we stop
// dispatching (any already-sent messages stand).
func run(db *gorm.DB, c *Campaign) {
	// Look up the template body once — cheaper than re-fetching per recipient.
	var tpl templates.MessageTemplate
	if err := db.Where("id = ?", c.TemplateID).First(&tpl).Error; err != nil {
		markCompleted(db, c, StatusFailed)
		return
	}

	var ids []string
	_ = json.Unmarshal(c.ContactIDs, &ids)

	sent, failed := 0, 0
	for _, cid := range ids {
		// Poll cancellation between recipients so a Cancel takes effect
		// without needing an interrupt mechanism.
		var current Campaign
		if err := db.Select("status").Where("id = ?", c.ID).First(&current).Error; err == nil {
			if current.Status == StatusCancelled {
				return
			}
		}

		var err error
		if sender != nil {
			tplID := c.TemplateID
			err = sender(c.OwnerID, cid, c.Channel, tpl.Body, &tplID)
		}
		if err == nil {
			sent++
			_ = db.Model(c).UpdateColumn("sent_count", gorm.Expr("sent_count + 1")).Error
		} else {
			failed++
			_ = db.Model(c).UpdateColumn("failed_count", gorm.Expr("failed_count + 1")).Error
		}
	}

	// Terminal state: completed if we sent anything, failed if nothing went.
	finalStatus := StatusCompleted
	if sent == 0 && failed > 0 {
		finalStatus = StatusFailed
	}
	markCompleted(db, c, finalStatus)
	if onComplete != nil {
		// Re-load so caller sees the latest counters.
		var reloaded Campaign
		if err := db.Where("id = ?", c.ID).First(&reloaded).Error; err == nil {
			onComplete(reloaded.OwnerID, &reloaded)
		}
	}
}

func markCompleted(db *gorm.DB, c *Campaign, status string) {
	now := time.Now()
	_ = db.Model(c).Updates(map[string]interface{}{
		"status":       status,
		"completed_at": now,
	}).Error
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
