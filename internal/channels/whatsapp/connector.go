package whatsapp

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

const cloudAPIBase = "https://graph.facebook.com/v21.0"

var httpClient = &http.Client{Timeout: 30 * time.Second}

type Connector struct {
	PhoneNumberID string
	AccessToken   string
}

func New(phoneNumberID, accessToken string) *Connector {
	return &Connector{PhoneNumberID: phoneNumberID, AccessToken: accessToken}
}

func (c *Connector) Name() string { return "whatsapp" }

func (c *Connector) ValidateConfig(config map[string]string) error {
	if config["phone_number_id"] == "" {
		return errors.New("whatsapp: phone_number_id is required")
	}
	if config["access_token"] == "" {
		return errors.New("whatsapp: access_token is required")
	}
	return nil
}

func (c *Connector) Send(msg *shared.OutboundMessage) (*shared.DeliveryResult, error) {
	if c.PhoneNumberID == "" || c.AccessToken == "" {
		return nil, errors.New("whatsapp: connector not configured")
	}

	var payload map[string]interface{}

	if msg.TemplateID != nil && *msg.TemplateID != "" {
		payload = map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                msg.ContactID,
			"type":              "template",
			"template": map[string]interface{}{
				"name": *msg.TemplateID,
				"language": map[string]string{
					"code": "en",
				},
			},
		}
	} else {
		payload = map[string]interface{}{
			"messaging_product": "whatsapp",
			"to":                msg.ContactID,
			"type":              "text",
			"text": map[string]string{
				"body": msg.Body,
			},
		}
	}

	if msg.MediaURL != nil && msg.Type == shared.TypeImage {
		payload["type"] = "image"
		payload["image"] = map[string]string{
			"link":    *msg.MediaURL,
			"caption": msg.Body,
		}
		delete(payload, "text")
	}

	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("%s/%s/messages", cloudAPIBase, c.PhoneNumberID)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("whatsapp: request creation failed: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("whatsapp: send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
				Code    int    `json:"code"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("whatsapp: API error %d — %s", errResp.Error.Code, errResp.Error.Message)
	}

	var result struct {
		Messages []struct {
			ID string `json:"id"`
		} `json:"messages"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	externalID := ""
	if len(result.Messages) > 0 {
		externalID = result.Messages[0].ID
	}

	return &shared.DeliveryResult{
		ExternalID: externalID,
		Status:     shared.StatusSent,
		Timestamp:  time.Now(),
	}, nil
}

// WebhookVerify handles the Meta webhook verification challenge (GET).
func WebhookVerify(mode, token, challenge, verifyToken string) (string, bool) {
	if mode == "subscribe" && token == verifyToken {
		return challenge, true
	}
	return "", false
}

func init() {
	shared.Register(&Connector{})
}
