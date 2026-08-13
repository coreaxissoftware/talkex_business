package talkex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/channels/shared"
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

type Connector struct {
	BaseURL string
	APIKey  string
}

func New(baseURL, apiKey string) *Connector {
	return &Connector{BaseURL: baseURL, APIKey: apiKey}
}

func (c *Connector) Name() string { return "talkex" }

func (c *Connector) ValidateConfig(config map[string]string) error {
	if config["base_url"] == "" {
		return errors.New("talkex: base_url is required")
	}
	if config["api_key"] == "" {
		return errors.New("talkex: api_key is required")
	}
	return nil
}

func (c *Connector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	payload := map[string]interface{}{
		"recipient_id": msg.ContactID,
		"body":         msg.Body,
		"type":         string(msg.Type),
	}
	if msg.MediaURL != nil {
		payload["media_url"] = *msg.MediaURL
	}

	body, _ := json.Marshal(payload)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("talkex: request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("talkex: send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("talkex: API returned HTTP %d", resp.StatusCode)
	}

	var result struct {
		MessageID string `json:"message_id"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	return &shared.DeliveryResult{
		ExternalID: result.MessageID,
		Status:     shared.StatusSent,
		Timestamp:  time.Now(),
	}, nil
}

func init() {
	// Auto-register with empty config — real config loaded at runtime from channels.Config
	shared.Register(&Connector{})
}
