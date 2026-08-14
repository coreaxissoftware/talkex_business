package channels

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InboundHandler processes incoming messages from a channel provider.
// Injected from main.go to avoid circular imports.
type InboundHandler func(ownerID, contactPhone, channel, body string) error

var inboundHandler InboundHandler

func RegisterInboundHandler(f InboundHandler) { inboundHandler = f }

// StatusUpdateHandler processes delivery status updates from a channel provider.
type StatusUpdateHandler func(messageID, status string) error

var statusUpdateHandler StatusUpdateHandler

func RegisterStatusUpdateHandler(f StatusUpdateHandler) { statusUpdateHandler = f }

// RegisterWebhookRoutes sets up the public (unauthenticated) endpoints that
// channel providers POST to when they have an incoming message or a delivery
// status update. These are separate from the dashboard API routes.
func RegisterWebhookRoutes(r *gin.Engine) {
	wh := r.Group("/webhooks/channels")
	{
		// Verification endpoint for WhatsApp/Meta webhook setup
		wh.GET("/:channel", handleWebhookVerify)

		// Inbound messages + status updates
		wh.POST("/:channel", handleWebhookInbound)
	}
}

// handleWebhookVerify handles the GET verification challenge used by
// WhatsApp, Instagram, and Messenger webhook setup. The provider sends
// hub.mode=subscribe, hub.challenge=<token>, hub.verify_token=<secret>.
func handleWebhookVerify(c *gin.Context) {
	mode := c.Query("hub.mode")
	challenge := c.Query("hub.challenge")
	// In production the verify_token should be checked against a stored
	// secret per-owner. For now we accept any verification request so the
	// webhook can be registered during development.
	if mode == "subscribe" && challenge != "" {
		c.String(http.StatusOK, challenge)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// handleWebhookInbound is the POST handler that receives events from
// channel providers. It accepts a generic JSON payload and routes it
// based on the :channel path parameter.
func handleWebhookInbound(c *gin.Context) {
	channel := c.Param("channel")

	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "invalid payload"})
		return
	}

	// Route based on event type — common patterns across providers:
	// 1. Inbound message: has "from", "body"/"text"/"message"
	// 2. Status update: has "message_id"/"mid", "status"

	switch channel {
	case "whatsapp":
		handleWhatsAppWebhook(c, payload)
	case "telegram":
		handleTelegramWebhook(c, payload)
	case "instagram", "messenger":
		handleMetaWebhook(c, channel, payload)
	default:
		// Generic handler — attempt to parse as a standard inbound
		handleGenericWebhook(c, channel, payload)
	}
}

func handleWhatsAppWebhook(c *gin.Context, payload map[string]interface{}) {
	// WhatsApp Cloud API webhook format:
	// { "object": "whatsapp_business_account", "entry": [...] }
	// Each entry has changes[].value.messages[] for inbound
	// and changes[].value.statuses[] for delivery updates
	if entries, ok := payload["entry"].([]interface{}); ok {
		for _, entry := range entries {
			e, _ := entry.(map[string]interface{})
			changes, _ := e["changes"].([]interface{})
			for _, change := range changes {
				ch, _ := change.(map[string]interface{})
				value, _ := ch["value"].(map[string]interface{})

				// Process inbound messages
				if messages, ok := value["messages"].([]interface{}); ok {
					for _, msg := range messages {
						m, _ := msg.(map[string]interface{})
						from, _ := m["from"].(string)
						text, _ := m["text"].(map[string]interface{})
						body, _ := text["body"].(string)
						if from != "" && body != "" && inboundHandler != nil {
							// Owner ID would need to be resolved from the WABA ID
							// in production. For now we log and accept.
							_ = inboundHandler("", from, "whatsapp", body)
						}
					}
				}

				// Process status updates
				if statuses, ok := value["statuses"].([]interface{}); ok {
					for _, stat := range statuses {
						s, _ := stat.(map[string]interface{})
						msgID, _ := s["id"].(string)
						status, _ := s["status"].(string)
						if msgID != "" && status != "" && statusUpdateHandler != nil {
							_ = statusUpdateHandler(msgID, status)
						}
					}
				}
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleTelegramWebhook(c *gin.Context, payload map[string]interface{}) {
	// Telegram Bot API webhook: { "update_id": ..., "message": { "from": ..., "text": ... } }
	if message, ok := payload["message"].(map[string]interface{}); ok {
		from, _ := message["from"].(map[string]interface{})
		chatID := ""
		if id, ok := from["id"].(float64); ok {
			chatID = fmt.Sprintf("%.0f", id)
		}
		text, _ := message["text"].(string)

		if chatID != "" && text != "" && inboundHandler != nil {
			_ = inboundHandler("", chatID, "telegram", text)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleMetaWebhook(c *gin.Context, channel string, payload map[string]interface{}) {
	// Instagram/Messenger webhook: same Meta format as WhatsApp but
	// with object=instagram or object=page
	if entries, ok := payload["entry"].([]interface{}); ok {
		for _, entry := range entries {
			e, _ := entry.(map[string]interface{})
			messaging, _ := e["messaging"].([]interface{})
			for _, event := range messaging {
				evt, _ := event.(map[string]interface{})
				sender, _ := evt["sender"].(map[string]interface{})
				senderID, _ := sender["id"].(string)
				message, _ := evt["message"].(map[string]interface{})
				text, _ := message["text"].(string)

				if senderID != "" && text != "" && inboundHandler != nil {
					_ = inboundHandler("", senderID, channel, text)
				}
			}
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func handleGenericWebhook(c *gin.Context, channel string, payload map[string]interface{}) {
	// Best-effort parse: look for common field names
	from, _ := payload["from"].(string)
	body, _ := payload["body"].(string)
	if body == "" {
		body, _ = payload["text"].(string)
	}
	if body == "" {
		body, _ = payload["message"].(string)
	}

	if from != "" && body != "" && inboundHandler != nil {
		_ = inboundHandler("", from, channel, body)
	}

	// Status updates
	msgID, _ := payload["message_id"].(string)
	status, _ := payload["status"].(string)
	if msgID != "" && status != "" && statusUpdateHandler != nil {
		_ = statusUpdateHandler(msgID, status)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
