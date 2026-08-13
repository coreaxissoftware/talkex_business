package shared

import "time"

type DeliveryStatus string

const (
	StatusQueued    DeliveryStatus = "queued"
	StatusSent      DeliveryStatus = "sent"
	StatusDelivered DeliveryStatus = "delivered"
	StatusRead      DeliveryStatus = "read"
	StatusFailed    DeliveryStatus = "failed"
)

type MessageType string

const (
	TypeText     MessageType = "text"
	TypeImage    MessageType = "image"
	TypeVideo    MessageType = "video"
	TypeDocument MessageType = "document"
	TypeAudio    MessageType = "audio"
	TypeTemplate MessageType = "template"
)

type OutboundMessage struct {
	ID          string      `json:"id"`
	OwnerID     string      `json:"owner_id"`
	ContactID   string      `json:"contact_id"`
	Channel     string      `json:"channel"`
	Type        MessageType `json:"type"`
	Body        string      `json:"body"`
	MediaURL    *string     `json:"media_url,omitempty"`
	TemplateID  *string     `json:"template_id,omitempty"`
	Variables   map[string]string `json:"variables,omitempty"`
}

type DeliveryResult struct {
	ExternalID string         `json:"external_id"`
	Status     DeliveryStatus `json:"status"`
	Error      string         `json:"error,omitempty"`
	Timestamp  time.Time      `json:"timestamp"`
}

type InboundMessage struct {
	ExternalID  string      `json:"external_id"`
	From        string      `json:"from"`
	Channel     string      `json:"channel"`
	Type        MessageType `json:"type"`
	Body        string      `json:"body"`
	MediaURL    *string     `json:"media_url,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// Connector is the interface every channel connector must implement.
type Connector interface {
	// Name returns the channel identifier (e.g. "talkex", "whatsapp").
	Name() string

	// Send delivers an outbound message through this channel.
	Send(msg *OutboundMessage) (*DeliveryResult, error)

	// ValidateConfig checks whether the channel's configuration is sufficient to send.
	ValidateConfig(config map[string]string) error
}

// Registry holds all registered channel connectors keyed by channel name.
var registry = map[string]Connector{}

func Register(c Connector) {
	registry[c.Name()] = c
}

func Get(name string) (Connector, bool) {
	c, ok := registry[name]
	return c, ok
}

func All() map[string]Connector {
	return registry
}
