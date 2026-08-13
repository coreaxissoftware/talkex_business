package messaging

import (
	"log"
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

type EnqueueInput struct {
	OwnerID    string
	CampaignID *string
	ContactID  string
	Channel    string
	Body       string
	TemplateID *string
	MediaURL   *string
	Priority   int
}

func Enqueue(db *gorm.DB, in *EnqueueInput) (*QueuedMessage, error) {
	if in.Priority == 0 {
		in.Priority = PriorityMarketing
	}
	msg := &QueuedMessage{
		OwnerID:    in.OwnerID,
		CampaignID: in.CampaignID,
		ContactID:  in.ContactID,
		Channel:    in.Channel,
		Body:       in.Body,
		TemplateID: in.TemplateID,
		MediaURL:   in.MediaURL,
		Status:     shared.StatusQueued,
		Priority:   in.Priority,
		MaxRetries: 3,
	}
	if err := db.Create(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

func EnqueueBatch(db *gorm.DB, items []EnqueueInput) (int, error) {
	msgs := make([]QueuedMessage, 0, len(items))
	for _, in := range items {
		pri := in.Priority
		if pri == 0 {
			pri = PriorityMarketing
		}
		msgs = append(msgs, QueuedMessage{
			OwnerID:    in.OwnerID,
			CampaignID: in.CampaignID,
			ContactID:  in.ContactID,
			Channel:    in.Channel,
			Body:       in.Body,
			TemplateID: in.TemplateID,
			MediaURL:   in.MediaURL,
			Status:     shared.StatusQueued,
			Priority:   pri,
			MaxRetries: 3,
		})
	}
	if err := db.CreateInBatches(msgs, 100).Error; err != nil {
		return 0, err
	}
	return len(msgs), nil
}

// ProcessQueue picks up queued/retryable messages and routes them through
// the appropriate channel connector. Called by the background worker.
func ProcessQueue(db *gorm.DB, batchSize int) int {
	var msgs []QueuedMessage
	now := time.Now()

	db.Where(
		"(status = ? OR (status = ? AND attempts < max_retries AND next_retry <= ?))",
		shared.StatusQueued, shared.StatusFailed, now,
	).Order("priority ASC, created_at ASC").Limit(batchSize).Find(&msgs)

	processed := 0
	for i := range msgs {
		msg := &msgs[i]
		processOne(db, msg)
		processed++
	}
	return processed
}

func processOne(db *gorm.DB, msg *QueuedMessage) {
	connector, ok := shared.Get(msg.Channel)
	if !ok {
		errMsg := "no connector registered for channel: " + msg.Channel
		msg.Status = shared.StatusFailed
		msg.Error = &errMsg
		db.Save(msg)
		return
	}

	msg.Attempts++
	result, err := connector.Send(&shared.OutboundMessage{
		ID:         msg.ID,
		OwnerID:    msg.OwnerID,
		ContactID:  msg.ContactID,
		Channel:    msg.Channel,
		Type:       shared.TypeText,
		Body:       msg.Body,
		MediaURL:   msg.MediaURL,
		TemplateID: msg.TemplateID,
	})

	if err != nil {
		errStr := err.Error()
		msg.Error = &errStr
		if msg.Attempts >= msg.MaxRetries {
			msg.Status = shared.StatusFailed
		} else {
			backoff := time.Duration(msg.Attempts*msg.Attempts) * 30 * time.Second
			next := time.Now().Add(backoff)
			msg.NextRetry = &next
		}
		db.Save(msg)
		log.Printf("messaging: send failed for %s (attempt %d): %v", msg.ID, msg.Attempts, err)
		return
	}

	now := time.Now()
	msg.Status = result.Status
	msg.ExternalID = &result.ExternalID
	msg.SentAt = &now
	msg.Error = nil
	db.Save(msg)
}

// GetQueueStats returns counts by status for a given owner.
func GetQueueStats(db *gorm.DB, ownerID string) map[string]int64 {
	stats := map[string]int64{}
	rows, err := db.Model(&QueuedMessage{}).
		Select("status, count(*) as cnt").
		Where("owner_id = ?", ownerID).
		Group("status").Rows()
	if err != nil {
		return stats
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var cnt int64
		rows.Scan(&status, &cnt)
		stats[status] = cnt
	}
	return stats
}

// GetCampaignDeliveryStats returns delivery stats for a specific campaign.
func GetCampaignDeliveryStats(db *gorm.DB, campaignID string) map[string]int64 {
	stats := map[string]int64{}
	rows, err := db.Model(&QueuedMessage{}).
		Select("status, count(*) as cnt").
		Where("campaign_id = ?", campaignID).
		Group("status").Rows()
	if err != nil {
		return stats
	}
	defer rows.Close()
	for rows.Next() {
		var status string
		var cnt int64
		rows.Scan(&status, &cnt)
		stats[status] = cnt
	}
	return stats
}
