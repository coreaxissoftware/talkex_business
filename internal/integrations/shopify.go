// Package integrations — outbound SaaS connectors.
//
// Shopify cart-abandonment is the canonical CPaaS e-commerce play: when
// Shopify's checkouts/create event fires and no order follows within
// its own configured window, Shopify emits checkouts/abandoned to any
// registered webhook. This handler receives that event, matches the
// customer to a Contact, and enqueues a WhatsApp cart-abandonment
// template message.
//
// The tenant must supply their SHOPIFY_WEBHOOK_SECRET (via Settings)
// so we can verify the X-Shopify-Hmac-Sha256 header; without a match
// we return 401 and log a warning.
package integrations

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// ShopifySendFn is the callback the messaging engine registers; it
// fires the cart-abandonment template on the given channel. Left as a
// callback to keep this package free of a messaging-engine dependency.
type ShopifySendFn func(ownerID, contactID, channel, templateID string, vars map[string]string) error

var shopifySend ShopifySendFn

// ShopifySecretFn resolves the per-tenant HMAC secret given an owner_id.
// Since Shopify's payload doesn't carry the owner_id, we require the
// tenant to include it as a query param `owner=<uuid>` when registering
// the webhook URL in Shopify Admin. Belt-and-braces: without a secret
// on file we reject.
type ShopifySecretFn func(ownerID string) (secret, templateID, channel string, ok bool)

var shopifySecretFn ShopifySecretFn

func RegisterShopifySender(f ShopifySendFn) { shopifySend = f }
func RegisterShopifySecret(f ShopifySecretFn) { shopifySecretFn = f }

// RegisterRoutes wires the public webhook endpoint.
func RegisterRoutes(r *gin.Engine) {
	// Shopify hits this with the raw JSON body — no auth middleware; we
	// verify with the HMAC header instead.
	r.POST("/integrations/shopify/webhook", handleShopifyWebhook)
	r.POST("/integrations/crm/webhook", handleCRMWebhook)
	r.GET("/integrations/zapier/events", handleZapierEvents)
}

func handleShopifyWebhook(c *gin.Context) {
	ownerID := c.Query("owner")
	if ownerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "owner query param required"})
		return
	}
	if shopifySecretFn == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "shopify integration not initialised"})
		return
	}
	secret, templateID, channel, ok := shopifySecretFn(ownerID)
	if !ok || secret == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "shopify not configured for this tenant"})
		return
	}

	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "read body"})
		return
	}
	provided := c.GetHeader("X-Shopify-Hmac-Sha256")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(raw)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(provided)) {
		log.Printf("shopify: HMAC mismatch for owner %s", ownerID)
		c.JSON(http.StatusUnauthorized, gin.H{"detail": "invalid signature"})
		return
	}

	var payload struct {
		AbandonedCheckoutURL string `json:"abandoned_checkout_url"`
		Customer             struct {
			Phone     string `json:"phone"`
			FirstName string `json:"first_name"`
			Email     string `json:"email"`
		} `json:"customer"`
		TotalPrice string `json:"total_price"`
		Currency   string `json:"currency"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "bad payload"})
		return
	}
	if payload.Customer.Phone == "" {
		c.JSON(http.StatusAccepted, gin.H{"detail": "no phone on checkout, skipped"})
		return
	}

	// Look up the contact by phone; opt-in required to send.
	contact, err := findContactByPhone(database.DB, ownerID, payload.Customer.Phone)
	if err != nil || contact == nil || !contact.OptedIn {
		c.JSON(http.StatusAccepted, gin.H{"detail": "no opted-in contact, skipped"})
		return
	}

	if shopifySend != nil {
		vars := map[string]string{
			"1": firstNameOr(payload.Customer.FirstName, "there"),
			"2": payload.TotalPrice,
			"3": payload.AbandonedCheckoutURL,
		}
		if err := shopifySend(ownerID, contact.ID, channel, templateID, vars); err != nil {
			log.Printf("shopify: send failed for %s: %v", contact.ID, err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "queued"})
}

func findContactByPhone(db *gorm.DB, ownerID, phone string) (*contacts.Contact, error) {
	var c contacts.Contact
	err := db.Where("owner_id = ? AND phone_number = ?", ownerID, phone).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func firstNameOr(name, fallback string) string {
	if name == "" {
		return fallback
	}
	return name
}
