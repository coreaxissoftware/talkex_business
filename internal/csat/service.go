package csat

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidScore    = errors.New("score must be between 1 and 5")
	ErrConversationNotOwned = errors.New("conversation does not belong to caller")
)

// ConversationOwnerCheck resolves a conversation to its owner without
// importing conversations (avoids an import cycle). Wired from main.go.
type ConversationOwnerCheck func(conversationID string) (ownerID, contactID, channel string, err error)

var convOwnerCheck ConversationOwnerCheck

// RegisterConversationOwnerCheck wires the ownership resolver.
func RegisterConversationOwnerCheck(f ConversationOwnerCheck) { convOwnerCheck = f }

type SubmitInput struct {
	ConversationID string  `json:"conversation_id" binding:"required"`
	ContactID      string  `json:"contact_id"`
	AgentUserID    *string `json:"agent_user_id,omitempty"`
	Score          int     `json:"score" binding:"required,min=1,max=5"`
	Comment        string  `json:"comment"`
	Channel        string  `json:"channel"`
}

func Submit(db *gorm.DB, ownerID string, in *SubmitInput) (*Rating, error) {
	if in.Score < 1 || in.Score > 5 {
		return nil, ErrInvalidScore
	}

	// Verify the conversation actually belongs to the caller and
	// canonicalize contact_id / channel from the row rather than
	// trusting client-supplied values (prevents cross-tenant
	// dashboard pollution and stale-field mismatches).
	contactID := in.ContactID
	channel := in.Channel
	if convOwnerCheck != nil {
		convOwner, convContactID, convChannel, err := convOwnerCheck(in.ConversationID)
		if err != nil || convOwner != ownerID {
			return nil, ErrConversationNotOwned
		}
		contactID = convContactID
		channel = convChannel
	}

	r := &Rating{
		OwnerID:        ownerID,
		ConversationID: in.ConversationID,
		ContactID:      contactID,
		AgentUserID:    in.AgentUserID,
		Score:          in.Score,
		Comment:        in.Comment,
		Channel:        channel,
	}
	if err := db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

func List(db *gorm.DB, ownerID string, limit int) ([]Rating, error) {
	var out []Rating
	q := db.Where("owner_id = ?", ownerID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&out).Error
	return out, err
}

// Summary aggregates CSAT ratings across a rolling 30-day window.
type Summary struct {
	Total       int64             `json:"total"`
	Average     float64           `json:"average"`
	Distribution map[int]int64    `json:"distribution"` // score → count
	PerAgent    []AgentAggregate  `json:"per_agent"`
	PerChannel  []ChannelAggregate `json:"per_channel"`
}

type AgentAggregate struct {
	AgentUserID string  `json:"agent_user_id"`
	Count       int64   `json:"count"`
	Average     float64 `json:"average"`
}

type ChannelAggregate struct {
	Channel string  `json:"channel"`
	Count   int64   `json:"count"`
	Average float64 `json:"average"`
}

func GetSummary(db *gorm.DB, ownerID string) (*Summary, error) {
	since := time.Now().AddDate(0, 0, -30)

	var all []Rating
	if err := db.Where("owner_id = ? AND created_at >= ?", ownerID, since).Find(&all).Error; err != nil {
		return nil, err
	}

	s := &Summary{
		Distribution: map[int]int64{1: 0, 2: 0, 3: 0, 4: 0, 5: 0},
	}

	sum := 0
	agents := map[string]*AgentAggregate{}
	channels := map[string]*ChannelAggregate{}

	for _, r := range all {
		s.Total++
		sum += r.Score
		s.Distribution[r.Score]++

		if r.AgentUserID != nil && *r.AgentUserID != "" {
			a, ok := agents[*r.AgentUserID]
			if !ok {
				a = &AgentAggregate{AgentUserID: *r.AgentUserID}
				agents[*r.AgentUserID] = a
			}
			a.Count++
			a.Average = ((a.Average * float64(a.Count-1)) + float64(r.Score)) / float64(a.Count)
		}

		if r.Channel != "" {
			ch, ok := channels[r.Channel]
			if !ok {
				ch = &ChannelAggregate{Channel: r.Channel}
				channels[r.Channel] = ch
			}
			ch.Count++
			ch.Average = ((ch.Average * float64(ch.Count-1)) + float64(r.Score)) / float64(ch.Count)
		}
	}

	if s.Total > 0 {
		s.Average = float64(sum) / float64(s.Total)
	}

	for _, a := range agents {
		s.PerAgent = append(s.PerAgent, *a)
	}
	for _, ch := range channels {
		s.PerChannel = append(s.PerChannel, *ch)
	}

	return s, nil
}
