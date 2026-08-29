// Package events provides an in-process pub/sub hub plus a Server-Sent
// Events endpoint so the frontend can react to inbound messages,
// notifications, and campaign changes without polling.
//
// Simple in-memory implementation — one hub per process. Horizontal
// scaling would swap this for Redis pub/sub or NATS; every publisher
// already goes through Publish() so the wire only needs a re-wiring.
package events

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/redisclient"
)

// redisChannel is the single Pub/Sub channel every pod publishes to
// and subscribes from. Payload is the JSON-encoded envelope below
// (with ownerID kept so subscribers can route per-tenant).
const redisChannel = "talkex.events"

// envelope is what we put on the wire across pods.
type envelope struct {
	OwnerID string      `json:"o"`
	Type    string      `json:"t"`
	Data    interface{} `json:"d"`
}

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
//
// When Redis is configured, the event is also published to a shared
// Pub/Sub channel so subscribers on OTHER pods see it. The local
// dispatch below happens either way so the request that triggered the
// event still gets an immediate notification.
func Publish(ownerID, eventType string, data interface{}) {
	dispatchLocal(ownerID, eventType, data)

	if rdb := redisclient.Get(); rdb != nil {
		payload, err := json.Marshal(envelope{OwnerID: ownerID, Type: eventType, Data: data})
		if err == nil {
			// Fire-and-forget — Pub/Sub is best-effort by design.
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			rdb.Publish(ctx, redisChannel, payload)
			cancel()
		}
	}
}

// dispatchLocal fans out to in-process subscribers only. Used by both
// Publish() (this pod) and StartRedisFanout() (events from other pods).
func dispatchLocal(ownerID, eventType string, data interface{}) {
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

// StartRedisFanout listens on the shared Pub/Sub channel and dispatches
// every event to local subscribers. Safe no-op when Redis isn't
// configured. Call once from main.go after wiring the hub.
func StartRedisFanout() {
	rdb := redisclient.Get()
	if rdb == nil {
		return
	}
	go func() {
		ctx := context.Background()
		ps := rdb.Subscribe(ctx, redisChannel)
		defer ps.Close()
		ch := ps.Channel()
		for msg := range ch {
			var env envelope
			if err := json.Unmarshal([]byte(msg.Payload), &env); err != nil {
				log.Printf("events: bad Redis payload: %v", err)
				continue
			}
			dispatchLocal(env.OwnerID, env.Type, env.Data)
		}
	}()
}

// SubscriberCount is used by health / debug endpoints.
func SubscriberCount() int {
	subsMu.RLock()
	defer subsMu.RUnlock()
	return len(subs)
}
