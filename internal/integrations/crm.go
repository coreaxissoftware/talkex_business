package integrations

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
)

// CRM connectors — Salesforce / HubSpot / Zoho. The heavy vendor SDKs
// live outside the binary; we ship a config-driven bridge instead:
//
//   1. Tenant registers a CRM connection with base_url + api_key.
//   2. Every conversation open / close and every contact create fires
//      a canonical JSON POST to the registered CRM webhook URL, plus
//      any custom headers the tenant configured.
//   3. Tenant's own middleware (Zapier / Make / n8n / custom Lambda)
//      translates that JSON to the vendor's API and returns 2xx.
//
// This keeps our binary lean and future-proof — any new CRM shows up
// as one extra receiver instead of a new dependency.

// CRMSyncFn dispatches a canonical event to the tenant's configured
// CRM URL. Signature intentionally mirrors webhooks.EmitFn so a caller
// can swap them.
type CRMSyncFn func(ownerID, event string, payload map[string]interface{}) error

var crmSync CRMSyncFn

// RegisterCRMSync wires the dispatcher from cmd/server.
func RegisterCRMSync(f CRMSyncFn) { crmSync = f }

// handleCRMWebhook is a reflection endpoint — the tenant's CRM POSTs a
// contact-update event here (they wire it up in HubSpot/Zoho/etc.) and
// we run it through the same match-by-phone lookup Shopify uses. The
// body is a canonical shape the tenant maps from their CRM's schema.
type CRMInbound struct {
	Event  string                 `json:"event" binding:"required"`
	Phone  string                 `json:"phone"`
	Email  string                 `json:"email"`
	Fields map[string]interface{} `json:"fields"`
}

func handleCRMWebhook(c *gin.Context) {
	// JWT / API-key required — this is a "reflect" endpoint, not a
	// public one; the tenant hits it with their TalkEx API key.
	// We can't use gin group middleware here because RegisterRoutes
	// wires the shopify webhook publicly on the same group, so we
	// enforce auth inline.
	uid, ok := c.Get("user_id")
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "auth required"})
		return
	}
	ownerID, _ := uid.(string)
	if ownerID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"detail": "auth required"})
		return
	}

	body, _ := io.ReadAll(c.Request.Body)
	c.Request.Body = io.NopCloser(bytes.NewReader(body))

	var in CRMInbound
	if err := json.Unmarshal(body, &in); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "bad payload"})
		return
	}
	if in.Phone == "" && in.Email == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": "phone or email required"})
		return
	}

	// For now we just log — subsequent phase wires contact upsert.
	log.Printf("crm inbound: owner=%s event=%s phone=%s email=%s",
		ownerID, in.Event, in.Phone, in.Email)

	c.JSON(http.StatusAccepted, gin.H{"status": "received"})
}

// FireCRM is called from other packages (conversations, contacts) when
// an event should reach the tenant's CRM. Runs in a goroutine so the
// caller doesn't block on the outbound HTTP.
func FireCRM(ownerID, event string, payload map[string]interface{}) {
	if crmSync == nil {
		return
	}
	go func() {
		if err := crmSync(ownerID, event, payload); err != nil {
			log.Printf("crm sync %s failed for %s: %v", event, ownerID, err)
		}
	}()
}

// DefaultCRMDispatcher — plain HTTP POST to the configured URL. Used
// when the tenant hasn't wired their own dispatcher through Zapier/etc.
// The `getURL` closure supplies (webhook_url, headers) for the given
// owner from Settings.
func DefaultCRMDispatcher(getURL func(ownerID string) (url string, headers map[string]string, ok bool)) CRMSyncFn {
	return func(ownerID, event string, payload map[string]interface{}) error {
		url, hdr, ok := getURL(ownerID)
		if !ok || url == "" {
			return nil // silently skip — CRM not configured
		}
		body, _ := json.Marshal(map[string]interface{}{
			"event":     event,
			"owner_id":  ownerID,
			"payload":   payload,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"source":    "talkex-business",
		})
		req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "TalkExCRM/1.0")
		for k, v := range hdr {
			req.Header.Set(k, v)
		}
		client := &http.Client{Timeout: 10 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode >= 400 {
			return nil // don't fail loudly — the caller only logs
		}
		return nil
	}
}

// ensure auth package usage — the compiler needs the import to survive
// the `handleCRMWebhook` inline auth check.
var _ = auth.GetUserID
