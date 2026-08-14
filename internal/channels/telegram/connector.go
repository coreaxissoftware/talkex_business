// Package telegram implements a Telegram Bot API connector.
// Messages are sent via the Bot API sendMessage / sendPhoto endpoints.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

const apiBase = "https://api.telegram.org"

var httpClient = &http.Client{Timeout: 15 * time.Second}

// Connector sends messages through the Telegram Bot API.
type Connector struct {
	BotToken string
}

func (c *Connector) Name() string { return "telegram" }

func (c *Connector) ValidateConfig(cfg map[string]string) error {
	if cfg["bot_token"] == "" {
		return fmt.Errorf("telegram: bot_token is required")
	}
	return nil
}

func (c *Connector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	token := c.BotToken
	if token == "" {
		return nil, fmt.Errorf("telegram: bot_token not configured")
	}

	// Build sendMessage payload
	payload := map[string]interface{}{
		"chat_id": msg.ContactID,
		"text":    msg.Body,
	}

	// If media is attached, use sendPhoto instead
	endpoint := "sendMessage"
	if msg.MediaURL != nil && *msg.MediaURL != "" {
		endpoint = "sendPhoto"
		payload["photo"] = *msg.MediaURL
		payload["caption"] = msg.Body
		delete(payload, "text")
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/bot%s/%s", apiBase, token, endpoint)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("telegram: request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram: send failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			MessageID int `json:"message_id"`
		} `json:"result"`
		Description string `json:"description"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if !result.OK {
		return nil, fmt.Errorf("telegram: API error: %s", result.Description)
	}

	return &shared.DeliveryResult{
		ExternalID: fmt.Sprintf("tg_%d", result.Result.MessageID),
		Status:     shared.StatusSent,
		Timestamp:  time.Now(),
	}, nil
}

func init() {
	shared.Register(&Connector{})
}
