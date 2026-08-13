package webhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/gorm"
)

var (
	ErrEndpointNotFound = errors.New("webhook endpoint not found")
)

// httpClient — shared, short-timeout client for outbound deliveries. The
// dashboard-visible delivery log gives operators the tools to spot
// misbehaving endpoints; we don't need to be forgiving here.
var httpClient = &http.Client{Timeout: 10 * time.Second}

// generateSecret returns 32 random bytes hex-encoded, used as the shared
// signing secret with the endpoint's operator.
func generateSecret() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// Sign returns hex(HMAC-SHA256(secret, body)) — the X-TalkEx-Signature
// header value receivers must recompute to verify authenticity.
func Sign(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

// ListEndpoints returns the owner's registered endpoints newest first.
func ListEndpoints(db *gorm.DB, ownerID string) ([]Endpoint, error) {
	var out []Endpoint
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}

type CreateEndpointInput struct {
	Name   string   `json:"name" binding:"required"`
	URL    string   `json:"url" binding:"required"`
	Events []string `json:"events" binding:"required"`
	Active bool     `json:"active"`
}

// CreateEndpointResult returns the row plus the plaintext secret — the
// secret is stored on the row too, but this response is the caller's cue
// to save it somewhere secure.
type CreateEndpointResult struct {
	Endpoint  Endpoint `json:"endpoint"`
	Plaintext string   `json:"plaintext_secret"`
}

func CreateEndpoint(db *gorm.DB, ownerID string, in *CreateEndpointInput) (*CreateEndpointResult, error) {
	secret, err := generateSecret()
	if err != nil {
		return nil, err
	}
	eventsJSON, _ := json.Marshal(in.Events)
	e := Endpoint{
		OwnerID: ownerID,
		Name:    in.Name,
		URL:     in.URL,
		Secret:  secret,
		Events:  eventsJSON,
		Active:  in.Active,
	}
	if err := db.Create(&e).Error; err != nil {
		return nil, err
	}
	return &CreateEndpointResult{Endpoint: e, Plaintext: secret}, nil
}

func GetEndpoint(db *gorm.DB, ownerID, id string) (*Endpoint, error) {
	var e Endpoint
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrEndpointNotFound
	}
	return &e, err
}

func DeleteEndpoint(db *gorm.DB, e *Endpoint) error {
	return db.Delete(e).Error
}

// Deliver ships a signed event body to every endpoint of `ownerID` whose
// events list contains `event`. Runs synchronously — callers who don't
// want to wait wrap it in a goroutine.
func Deliver(db *gorm.DB, ownerID, event string, payload interface{}) {
	var endpoints []Endpoint
	if err := db.Where("owner_id = ? AND active = ?", ownerID, true).Find(&endpoints).Error; err != nil {
		return
	}
	body, err := json.Marshal(map[string]interface{}{
		"event":     event,
		"data":      payload,
		"delivered": time.Now().UTC(),
	})
	if err != nil {
		return
	}

	for i := range endpoints {
		ep := &endpoints[i]
		// Filter by subscribed events.
		var events []string
		_ = json.Unmarshal(ep.Events, &events)
		matches := false
		for _, ev := range events {
			if ev == event {
				matches = true
				break
			}
		}
		if !matches {
			continue
		}
		go post(db, ep, event, body)
	}
}

// ListDeliveries returns delivery log rows for a given endpoint, newest
// first — capped so a chatty endpoint doesn't dump megabytes to the UI.
func ListDeliveries(db *gorm.DB, ownerID, endpointID string) ([]Delivery, error) {
	// Verify ownership first so we don't leak another tenant's history.
	if _, err := GetEndpoint(db, ownerID, endpointID); err != nil {
		return nil, err
	}
	var out []Delivery
	err := db.Where("endpoint_id = ?", endpointID).Order("created_at DESC").Limit(50).Find(&out).Error
	return out, err
}

func RetryDelivery(db *gorm.DB, ownerID, deliveryID string) error {
	var d Delivery
	if err := db.Where("id = ?", deliveryID).First(&d).Error; err != nil {
		return ErrEndpointNotFound
	}
	ep, err := GetEndpoint(db, ownerID, d.EndpointID)
	if err != nil {
		return err
	}
	go post(db, ep, d.Event, []byte(d.Payload))
	return nil
}

func post(db *gorm.DB, ep *Endpoint, event string, body []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, ep.URL, bytes.NewReader(body))
	if err != nil {
		recordFailure(db, ep, event, body, 0, err.Error())
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-TalkEx-Event", event)
	req.Header.Set("X-TalkEx-Signature", Sign(ep.Secret, body))

	resp, err := httpClient.Do(req)
	if err != nil {
		recordFailure(db, ep, event, body, 0, err.Error())
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	now := time.Now()
	d := &Delivery{
		EndpointID: ep.ID,
		Event:      event,
		Payload:    string(body),
		StatusCode: resp.StatusCode,
		Success:    success,
		Attempts:   1,
	}
	if success {
		d.DeliveredAt = &now
	} else {
		msg := fmt.Sprintf("HTTP %d", resp.StatusCode)
		d.ErrorMessage = &msg
	}
	_ = db.Create(d).Error

	_ = db.Model(ep).Update("last_fired", now).Error
}

func recordFailure(db *gorm.DB, ep *Endpoint, event string, body []byte, status int, msg string) {
	d := &Delivery{
		EndpointID:   ep.ID,
		Event:        event,
		Payload:      string(body),
		StatusCode:   status,
		Success:      false,
		Attempts:     1,
		ErrorMessage: &msg,
	}
	_ = db.Create(d).Error
}
