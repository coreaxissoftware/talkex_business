// Package analytics — cross-cutting aggregations over messages, campaigns,
// and contacts. Read-only; every query is scoped to the caller's owner_id.
//
// Data comes from what other modules already write — messages table, plus
// campaigns table for the campaigns KPI. Nothing here writes to the DB.
package analytics

import (
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/campaigns"
	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"github.com/coreaxissoftware/talkex_business/internal/conversations"
)

// Summary is the top-of-dashboard KPI card set.
type Summary struct {
	TotalMessages     int64                `json:"total_messages"`
	OutboundMessages  int64                `json:"outbound_messages"`
	InboundMessages   int64                `json:"inbound_messages"`
	DeliveryRate      float64              `json:"delivery_rate"`
	OpenConversations int64                `json:"open_conversations"`
	TotalContacts     int64                `json:"total_contacts"`
	ActiveCampaigns   int64                `json:"active_campaigns"`
	ByStatus          map[string]int64     `json:"by_status"`
	ByChannel         []ChannelBreakdown   `json:"by_channel"`
}

type ChannelBreakdown struct {
	Channel string `json:"channel"`
	Count   int64  `json:"count"`
}

// GetSummary computes all KPIs in a handful of queries, scoped to owner_id.
// Every SELECT that touches the messages table joins through conversations
// so we can filter by owner without letting one tenant read another's data.
func GetSummary(db *gorm.DB, ownerID string) (*Summary, error) {
	s := &Summary{ByStatus: map[string]int64{}}

	msgBase := db.Model(&conversations.Message{}).
		Joins("JOIN conversations ON conversations.id = messages.conversation_id").
		Where("conversations.owner_id = ?", ownerID)

	if err := msgBase.Count(&s.TotalMessages).Error; err != nil {
		return nil, err
	}
	if err := msgBase.Session(&gorm.Session{}).Where("messages.direction = ?", conversations.Outbound).Count(&s.OutboundMessages).Error; err != nil {
		return nil, err
	}
	if err := msgBase.Session(&gorm.Session{}).Where("messages.direction = ?", conversations.Inbound).Count(&s.InboundMessages).Error; err != nil {
		return nil, err
	}

	// Delivery rate = delivered+read / outbound (only outbound has delivery
	// semantics; inbound is trivially "delivered" to us).
	if s.OutboundMessages > 0 {
		var delivered int64
		if err := msgBase.Session(&gorm.Session{}).
			Where("messages.direction = ? AND messages.status IN ?", conversations.Outbound, []string{conversations.MsgStatusDelivered, conversations.MsgStatusRead}).
			Count(&delivered).Error; err != nil {
			return nil, err
		}
		s.DeliveryRate = float64(delivered) / float64(s.OutboundMessages) * 100
	}

	// By-status breakdown across all outbound messages.
	type row struct {
		Status string
		N      int64
	}
	var rows []row
	if err := msgBase.Session(&gorm.Session{}).
		Select("messages.status as status, COUNT(*) as n").
		Where("messages.direction = ?", conversations.Outbound).
		Group("messages.status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	for _, r := range rows {
		s.ByStatus[r.Status] = r.N
	}

	// By-channel breakdown across conversations (a conversation's channel is
	// the message's channel — messages don't repeat that column).
	var channels []ChannelBreakdown
	if err := db.Model(&conversations.Conversation{}).
		Select("channel, COUNT(*) as count").
		Where("owner_id = ?", ownerID).
		Group("channel").
		Order("count DESC").
		Scan(&channels).Error; err != nil {
		return nil, err
	}
	s.ByChannel = channels

	// Open conversations = last_inbound_at within the last 24h.
	cutoff := time.Now().Add(-conversations.WindowHours * time.Hour)
	if err := db.Model(&conversations.Conversation{}).
		Where("owner_id = ? AND last_inbound_at IS NOT NULL AND last_inbound_at > ?", ownerID, cutoff).
		Count(&s.OpenConversations).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&contacts.Contact{}).
		Where("owner_id = ?", ownerID).
		Count(&s.TotalContacts).Error; err != nil {
		return nil, err
	}

	if err := db.Model(&campaigns.Campaign{}).
		Where("owner_id = ? AND status IN ?", ownerID, []string{campaigns.StatusScheduled, campaigns.StatusRunning}).
		Count(&s.ActiveCampaigns).Error; err != nil {
		return nil, err
	}

	return s, nil
}

// TimeseriesPoint is one bucket in the daily-volume line chart.
type TimeseriesPoint struct {
	Date     string `json:"date"` // YYYY-MM-DD
	Outbound int64  `json:"outbound"`
	Inbound  int64  `json:"inbound"`
}

// GetTimeseries returns per-day outbound/inbound counts for the last `days`
// days, with any missing days filled in as zeros — so the frontend renders
// a continuous line without having to know which dates the DB actually saw.
func GetTimeseries(db *gorm.DB, ownerID string, days int) ([]TimeseriesPoint, error) {
	if days <= 0 || days > 90 {
		days = 30
	}

	end := time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)
	start := end.Add(-time.Duration(days) * 24 * time.Hour)

	type row struct {
		Day       string
		Direction string
		N         int64
	}
	var raw []row
	if err := db.Model(&conversations.Message{}).
		Joins("JOIN conversations ON conversations.id = messages.conversation_id").
		Select("strftime('%Y-%m-%d', messages.created_at) as day, messages.direction as direction, COUNT(*) as n").
		Where("conversations.owner_id = ? AND messages.created_at >= ?", ownerID, start).
		Group("day, direction").
		Scan(&raw).Error; err != nil {
		return nil, err
	}

	// Bucket into date → {out,in}
	byDay := map[string]*TimeseriesPoint{}
	for _, r := range raw {
		p, ok := byDay[r.Day]
		if !ok {
			p = &TimeseriesPoint{Date: r.Day}
			byDay[r.Day] = p
		}
		if r.Direction == conversations.Outbound {
			p.Outbound = r.N
		} else {
			p.Inbound = r.N
		}
	}

	// Fill missing days.
	out := make([]TimeseriesPoint, 0, days)
	for i := 0; i < days; i++ {
		d := start.Add(time.Duration(i) * 24 * time.Hour).Format("2006-01-02")
		if p, ok := byDay[d]; ok {
			out = append(out, *p)
		} else {
			out = append(out, TimeseriesPoint{Date: d})
		}
	}
	return out, nil
}
