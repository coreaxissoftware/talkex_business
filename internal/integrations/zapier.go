package integrations

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Zapier trigger catalogue — the read-only list of events a Zapier
// (or Make / n8n / Pabbly) app can subscribe to. Tenants point their
// Zapier "webhook" trigger at /webhooks (existing endpoint) after
// picking one of these event names in the Zap editor.
//
// This endpoint is what Zapier calls to render the "trigger event"
// dropdown when a user builds a Zap against TalkEx.

// zapierEvent describes one triggerable event for the Zapier UI.
type zapierEvent struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	SampleKeys  []string `json:"sample_keys"`
}

var zapierCatalog = []zapierEvent{
	{
		Key:         "message.received",
		Name:        "New inbound message",
		Description: "Fires when a customer replies on any channel.",
		SampleKeys:  []string{"contact_id", "channel", "body", "conversation_id", "received_at"},
	},
	{
		Key:         "message.delivered",
		Name:        "Outbound message delivered",
		Description: "Fires when a campaign or agent send is confirmed by the channel.",
		SampleKeys:  []string{"message_id", "contact_id", "channel", "delivered_at"},
	},
	{
		Key:         "conversation.assigned",
		Name:        "Conversation assigned to agent",
		Description: "Fires when a conversation is (re-)assigned.",
		SampleKeys:  []string{"conversation_id", "agent_user_id", "assigned_at"},
	},
	{
		Key:         "conversation.closed",
		Name:        "Conversation closed",
		Description: "Fires when an agent marks a conversation resolved.",
		SampleKeys:  []string{"conversation_id", "closed_by", "closed_at", "duration_seconds"},
	},
	{
		Key:         "campaign.completed",
		Name:        "Campaign completed",
		Description: "Fires when a campaign's last message is dispatched.",
		SampleKeys:  []string{"campaign_id", "sent_count", "delivered_count", "failed_count", "total_cost"},
	},
	{
		Key:         "contact.created",
		Name:        "New contact",
		Description: "Fires when a contact is added (manual, CSV import, or first inbound).",
		SampleKeys:  []string{"contact_id", "phone_number", "name", "source"},
	},
	{
		Key:         "contact.opted_out",
		Name:        "Contact opted out",
		Description: "Fires when a customer sends STOP / UNSUBSCRIBE.",
		SampleKeys:  []string{"contact_id", "phone_number", "matched_keyword", "at"},
	},
	{
		Key:         "csat.submitted",
		Name:        "CSAT rating submitted",
		Description: "Fires when a customer submits a CSAT score.",
		SampleKeys:  []string{"conversation_id", "score", "comment", "submitted_at"},
	},
	{
		Key:         "wallet.low",
		Name:        "Wallet balance low",
		Description: "Fires when wallet balance falls below the tenant's alert threshold.",
		SampleKeys:  []string{"balance", "threshold", "checked_at"},
	},
	{
		Key:         "sla.breached",
		Name:        "SLA breached",
		Description: "Fires when a conversation stays open past the SLA threshold.",
		SampleKeys:  []string{"conversation_id", "waited_minutes", "breach_at"},
	},
}

func handleZapierEvents(c *gin.Context) {
	c.JSON(http.StatusOK, zapierCatalog)
}
