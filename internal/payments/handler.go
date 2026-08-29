package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/config"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// verifyRazorpaySignature checks the X-Razorpay-Signature header
// (HMAC-SHA256(secret, rawBody), hex-encoded) against a constant-time
// comparison. Returns true when the signature matches.
func verifyRazorpaySignature(rawBody []byte, signature, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	// hmac.Equal is constant-time
	return hmac.Equal([]byte(expected), []byte(signature))
}

// RegisterRoutes wires payment endpoints.
func RegisterRoutes(r *gin.Engine) {
	g := r.Group("/payments")
	g.Use(auth.AuthRequired())
	{
		g.POST("/order", handleCreateOrder)
		g.GET("/orders", handleListOrders)
		g.POST("/dev-simulate", handleDevSimulate)
	}
	// Webhook is public — signed by provider
	r.POST("/payments/webhook", handleWebhook)
}

type createOrderReq struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func handleCreateOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	o, err := CreateOrder(database.DB, auth.GetUserID(c), req.Amount)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Failed to create order"})
		return
	}

	// Client uses these to open Razorpay's checkout modal (in prod).
	// In dev we return a hint that it can call /payments/dev-simulate.
	cfg := config.Get()
	keyID := cfg.RazorpayKeyID
	if keyID == "" && cfg.IsDev() {
		keyID = "rzp_test_DEV_SIMULATION"
	}

	c.JSON(http.StatusCreated, gin.H{
		"order_id": o.ID,
		"amount":   o.Amount,
		"currency": o.Currency,
		"key_id":   keyID,
		"dev_mode": cfg.RazorpayKeyID == "",
	})
}

func handleListOrders(c *gin.Context) {
	items, err := ListOrders(database.DB, auth.GetUserID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, items)
}

// handleDevSimulate is a dev-only endpoint that marks an order as paid
// so the local flow completes without a real Razorpay integration.
// Fatal to expose in production — guarded by IsDev().
func handleDevSimulate(c *gin.Context) {
	if !config.Get().IsDev() {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Not available"})
		return
	}

	var req struct {
		OrderID string `json:"order_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"detail": err.Error()})
		return
	}

	// Verify caller owns the order (defense-in-depth even in dev)
	o, err := GetOrder(database.DB, auth.GetUserID(c), req.OrderID)
	if err == ErrOrderNotFound {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Order not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	paymentID := randomID("pay_")
	log.Printf("payments (dev): simulating payment %s for order %s (%.2f INR)", paymentID, o.ID, o.Amount)

	updated, err := MarkPaid(database.DB, o.ID, paymentID)
	if err != nil && err != ErrAlreadyPaid {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

// handleWebhook receives Razorpay webhook events.
//
// Auth model:
//   - When RAZORPAY_WEBHOOK_SECRET is set, HMAC signature verification
//     is enforced. Missing / mismatched signatures are rejected 401.
//   - When it is NOT set AND the server is in dev mode, unsigned
//     events are accepted (so a local `curl` can exercise the flow).
//   - When it is NOT set outside dev mode, the endpoint refuses to
//     credit anything so a forgotten production config can't be
//     exploited to mint free wallet balance.
func handleWebhook(c *gin.Context) {
	// Read the raw body — signature must be over the exact bytes the
	// provider signed. c.ShouldBindJSON re-serialization would break
	// the HMAC comparison.
	raw, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Cannot read body"})
		return
	}

	cfg := config.Get()
	secret := cfg.RazorpayWebhookSecret
	signature := c.GetHeader("X-Razorpay-Signature")

	if secret != "" {
		if !verifyRazorpaySignature(raw, signature, secret) {
			log.Printf("payments: webhook signature verification failed")
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "Invalid signature"})
			return
		}
	} else if !cfg.IsDev() {
		// Fail closed in production: without a secret we cannot trust
		// any caller, so refuse rather than silently credit wallets.
		log.Printf("payments: refusing webhook, RAZORPAY_WEBHOOK_SECRET not set (non-dev)")
		c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "Webhook not configured"})
		return
	}

	var payload struct {
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					ID      string `json:"id"`
					OrderID string `json:"order_id"`
					Status  string `json:"status"`
				} `json:"entity"`
			} `json:"payment"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid payload"})
		return
	}

	if payload.Event == "payment.captured" && payload.Payload.Payment.Entity.Status == "captured" {
		orderID := payload.Payload.Payment.Entity.OrderID
		paymentID := payload.Payload.Payment.Entity.ID
		if _, err := MarkPaid(database.DB, orderID, paymentID); err != nil && err != ErrAlreadyPaid {
			log.Printf("payments: webhook credit failed for %s: %v", orderID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
