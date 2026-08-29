// Package events provides an in-process pub/sub hub plus a Server-Sent
// Events endpoint so the frontend can react to inbound messages,
// notifications, and campaign changes without polling.
//
// Simple in-memory implementation — one hub per process. Horizontal
// scaling would swap this for Redis pub/sub or NATS; every publisher
// already goes through Publish() so the wire only needs a re-wiring.
package events

import (
	"encoding/json"
	"sync"
)

// Event types — keep in sync with frontend/src/services/events.ts
const (
	TypeMessageInbound    = "message.inbound"
	TypeMessageOutbound   = "message.outbound"
	TypeMessageStatus     = "message.status"
	TypeNotificationNew   = "notification.new"
	TypeCampaignChanged   = "campaign.changed"
	TypeConversationUpdate = "conversation.update"
)

// Event is a per-owner message on the bus.
type Event struct {
	Type    string      `json:"type"`
	OwnerID string      `json:"-"` // routing key, never sent to clients
	Data    interface{} `json:"data"`
}

// Encoded returns the event payload as JSON minus the owner (which is
// used only for routing, never leaked to a browser).
func (e Event) Encoded() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"type": e.Type,
		"data": e.Data,
	})
}

// subscriber holds one open SSE connection's send channel.
type subscriber struct {
	ownerID string
	ch      chan Event
}

var (
	subs   = map[*subscriber]struct{}{}
	subsMu sync.RWMutex
)

// Subscribe registers a new listener for one owner. The returned
// channel is closed only when Unsubscribe is called.
func Subscribe(ownerID string) *subscriber {
	s := &subscriber{
		ownerID: ownerID,
		ch:      make(chan Event, 16), // small buffer — drop old on backpressure
	}
	subsMu.Lock()
	subs[s] = struct{}{}
	subsMu.Unlock()
	return s
}

// Unsubscribe removes a listener and closes its channel.
func Unsubscribe(s *subscriber) {
	subsMu.Lock()
	if _, ok := subs[s]; ok {
		delete(subs, s)
		close(s.ch)
	}
	subsMu.Unlock()
}

// Channel exposes the subscriber's receive-only channel.
func (s *subscriber) Channel() <-chan Event { return s.ch }

// Publish fans an event out to every subscriber for that owner.
// Non-blocking: if a subscriber's buffer is full the event is dropped
// for THAT client so a slow browser doesn't back up the whole hub.
func Publish(ownerID, eventType string, data interface{}) {
	e := Event{Type: eventType, OwnerID: ownerID, Data: data}
	subsMu.RLock()
	defer subsMu.RUnlock()
	for s := range subs {
		if s.ownerID != ownerID {
			continue
		}
		select {
		case s.ch <- e:
		default:
			// drop: subscriber is falling behind
		}
	}
}

// SubscriberCount is used by health / debug endpoints.
func SubscriberCount() int {
	subsMu.RLock()
	defer subsMu.RUnlock()
	return len(subs)
}
