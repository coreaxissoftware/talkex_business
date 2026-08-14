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

// FallbackResolver looks up the fallback channel for a contact.
type FallbackResolver func(ownerID, contactID string) (fallbackChannel string, ok bool)

// SandboxChecker returns true if the owner has sandbox mode enabled.
type SandboxChecker func(ownerID string) bool

// CostResolver returns (costPerMessage, sellPrice) for a given owner+channel.
type CostResolver func(ownerID, channel string) (cost float64, sell float64)

var (
	fallbackResolver FallbackResolver
	sandboxChecker   SandboxChecker
	costResolver     CostResolver
)

func RegisterFallbackResolver(f FallbackResolver) { fallbackResolver = f }
func RegisterSandboxChecker(f SandboxChecker)      { sandboxChecker = f }
func RegisterCostResolver(f CostResolver)          { costResolver = f }

func Enqueue(db *gorm.DB, in *EnqueueInput) (*QueuedMessage, error) {
	if in.Priority == 0 {
		in.Priority = PriorityMarketing
	}
	var cost, sell float64
	if costResolver != nil {
		cost, sell = costResolver(in.OwnerID, in.Channel)
	}
	msg := &QueuedMessage{
		OwnerID:        in.OwnerID,
		CampaignID:     in.CampaignID,
		ContactID:      in.ContactID,
		Channel:        in.Channel,
		Body:           in.Body,
		TemplateID:     in.TemplateID,
		MediaURL:       in.MediaURL,
		Status:         shared.StatusQueued,
		Priority:       in.Priority,
		MaxRetries:     3,
		CostPerMessage: cost,
		SellPrice:      sell,
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
	// Route through sandbox connector if owner has sandbox mode enabled
	channelToUse := msg.Channel
	if sandboxChecker != nil && sandboxChecker(msg.OwnerID) {
		channelToUse = "sandbox"
	}

	connector, ok := shared.Get(channelToUse)
	if !ok {
		errMsg := "no connector registered for channel: " + channelToUse
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
			// Try fallback channel before marking as permanently failed
			if !msg.FallbackTried && fallbackResolver != nil {
				if fbChannel, ok := fallbackResolver(msg.OwnerID, msg.ContactID); ok && fbChannel != "" && fbChannel != msg.Channel {
					log.Printf("messaging: switching %s from %s to fallback channel %s", msg.ID, msg.Channel, fbChannel)
					orig := msg.Channel
					msg.OriginalChannel = &orig
					msg.Channel = fbChannel
					msg.FallbackTried = true
					msg.Attempts = 0
					msg.Status = shared.StatusQueued
					msg.Error = nil
					db.Save(msg)
					return
				}
			}
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

// CostSummary aggregates cost and revenue for an owner.
type CostSummary struct {
	TotalCost     float64            `json:"total_cost"`
	TotalRevenue  float64            `json:"total_revenue"`
	TotalMargin   float64            `json:"total_margin"`
	MarginPercent float64            `json:"margin_percent"`
	ByChannel     []ChannelCostBreak `json:"by_channel"`
}

type ChannelCostBreak struct {
	Channel  string  `json:"channel"`
	Messages int64   `json:"messages"`
	Cost     float64 `json:"cost"`
	Revenue  float64 `json:"revenue"`
	Margin   float64 `json:"margin"`
}

// GetCostSummary computes cost vs revenue for all sent messages.
func GetCostSummary(db *gorm.DB, ownerID string) *CostSummary {
	type row struct {
		Channel  string
		Messages int64
		Cost     float64
		Revenue  float64
	}
	var rows []row
	db.Model(&QueuedMessage{}).
		Select("channel, COUNT(*) as messages, COALESCE(SUM(cost_per_message),0) as cost, COALESCE(SUM(sell_price),0) as revenue").
		Where("owner_id = ? AND status IN ?", ownerID, []string{string(shared.StatusSent), string(shared.StatusDelivered)}).
		Group("channel").
		Scan(&rows)

	s := &CostSummary{}
	for _, r := range rows {
		margin := r.Revenue - r.Cost
		s.TotalCost += r.Cost
		s.TotalRevenue += r.Revenue
		s.ByChannel = append(s.ByChannel, ChannelCostBreak{
			Channel:  r.Channel,
			Messages: r.Messages,
			Cost:     r.Cost,
			Revenue:  r.Revenue,
			Margin:   margin,
		})
	}
	s.TotalMargin = s.TotalRevenue - s.TotalCost
	if s.TotalRevenue > 0 {
		s.MarginPercent = (s.TotalMargin / s.TotalRevenue) * 100
	}
	return s
}

// GetCampaignCost returns total cost and revenue for a specific campaign.
func GetCampaignCost(db *gorm.DB, campaignID string) (cost float64, revenue float64) {
	type result struct {
		Cost    float64
		Revenue float64
	}
	var r result
	db.Model(&QueuedMessage{}).
		Select("COALESCE(SUM(cost_per_message),0) as cost, COALESCE(SUM(sell_price),0) as revenue").
		Where("campaign_id = ?", campaignID).
		Scan(&r)
	return r.Cost, r.Revenue
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
