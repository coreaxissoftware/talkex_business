package paylinks

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/config"
)

var (
	ErrNotFound     = errors.New("pay link not found")
	ErrInvalidInput = errors.New("amount must be > 0 and description required")
)

// CreateInput is what handlers accept from the UI.
type CreateInput struct {
	ContactID      string `json:"contact_id" binding:"required"`
	ConversationID string `json:"conversation_id"`
	AmountPaise    int64  `json:"amount_paise" binding:"required"`
	Description    string `json:"description" binding:"required"`
	ExpireHours    int    `json:"expire_hours"` // 0 → default 24h
}

// Create mints a Razorpay Quick Link. When Razorpay credentials aren't
// configured (dev mode) it fabricates a placeholder URL so the rest of
// the flow (send, mark paid, etc.) can be exercised end-to-end.
func Create(db *gorm.DB, ownerID string, in *CreateInput) (*PayLink, error) {
	if in.AmountPaise <= 0 || in.Description == "" {
		return nil, ErrInvalidInput
	}
	if in.ExpireHours <= 0 {
		in.ExpireHours = 24
	}
	expiry := time.Now().Add(time.Duration(in.ExpireHours) * time.Hour)

	pl := &PayLink{
		OwnerID:        ownerID,
		ContactID:      in.ContactID,
		ConversationID: in.ConversationID,
		AmountPaise:    in.AmountPaise,
		Currency:       "INR",
		Description:    in.Description,
		Status:         StatusCreated,
		ExpiresAt:      &expiry,
	}

	// Try Razorpay for real; fall back to simulation.
	cfg := config.Get()
	if cfg.RazorpayKeyID != "" && cfg.RazorpaySecret != "" {
		id, url, err := mintRazorpayLink(cfg.RazorpayKeyID, cfg.RazorpaySecret, pl)
		if err != nil {
			log.Printf("paylinks: razorpay create failed: %v (falling back to sim)", err)
			pl.Simulated = true
			pl.URL = simulatedURL(pl)
		} else {
			pl.RazorpayID = id
			pl.URL = url
		}
	} else {
		pl.Simulated = true
		pl.URL = simulatedURL(pl)
	}

	if err := db.Create(pl).Error; err != nil {
		return nil, err
	}
	return pl, nil
}

// mintRazorpayLink calls POST /v1/payment_links on Razorpay with basic
// auth. Returns (razorpay_id, short_url) on success. Timeout is 6s so
// a Razorpay hiccup doesn't wedge the request.
func mintRazorpayLink(keyID, secret string, pl *PayLink) (string, string, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"amount":      pl.AmountPaise,
		"currency":    pl.Currency,
		"description": pl.Description,
		"expire_by":   pl.ExpiresAt.Unix(),
		"notes": map[string]string{
			"owner_id":        pl.OwnerID,
			"conversation_id": pl.ConversationID,
			"contact_id":      pl.ContactID,
			"source":          "talkex-business",
		},
	})
	req, err := http.NewRequest(http.MethodPost, "https://api.razorpay.com/v1/payment_links", bytes.NewReader(body))
	if err != nil {
		return "", "", err
	}
	auth := base64.StdEncoding.EncodeToString([]byte(keyID + ":" + secret))
	req.Header.Set("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 6 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		return "", "", fmt.Errorf("razorpay: HTTP %d", res.StatusCode)
	}
	var out struct {
		ID       string `json:"id"`
		ShortURL string `json:"short_url"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return "", "", err
	}
	if out.ID == "" || out.ShortURL == "" {
		return "", "", errors.New("razorpay: missing id or short_url")
	}
	return out.ID, out.ShortURL, nil
}

// simulatedURL — dev-mode placeholder. Deterministic per-record so an
// integration test can assert against a stable string.
func simulatedURL(pl *PayLink) string {
	return fmt.Sprintf("https://rzp.io/sim/%s-%d", pl.ContactID[:8], pl.AmountPaise)
}

// MarkSent records that the URL has been posted to the customer.
// Called by the messaging engine after successful send.
func MarkSent(db *gorm.DB, id string) error {
	now := time.Now()
	return db.Model(&PayLink{}).Where("id = ?", id).
		Updates(map[string]interface{}{"status": StatusSent, "sent_at": &now}).Error
}

// MarkPaid flips the row when the Razorpay webhook confirms payment.
// Idempotent — replaying the webhook is a no-op past the first success.
func MarkPaid(db *gorm.DB, razorpayID string) (*PayLink, error) {
	var pl PayLink
	err := db.Where("razorpay_id = ?", razorpayID).First(&pl).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if pl.Status == StatusPaid {
		return &pl, nil
	}
	now := time.Now()
	pl.Status = StatusPaid
	pl.PaidAt = &now
	return &pl, db.Save(&pl).Error
}

// List returns pay links for the tenant, newest first. Optional status filter.
func List(db *gorm.DB, ownerID, status string) ([]PayLink, error) {
	q := db.Where("owner_id = ?", ownerID)
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var out []PayLink
	err := q.Order("created_at DESC").Limit(200).Find(&out).Error
	return out, err
}
