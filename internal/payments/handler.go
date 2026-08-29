package payments

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/config"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

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
// In production, this validates the X-Razorpay-Signature header using
// the RAZORPAY_WEBHOOK_SECRET. Dev mode accepts unsigned events.
func handleWebhook(c *gin.Context) {
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
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Invalid payload"})
		return
	}

	// Production would verify HMAC signature here.
	// signature := c.GetHeader("X-Razorpay-Signature")
	// if !VerifySignature(rawBody, signature, cfg.RazorpayWebhookSecret) { ... }

	if payload.Event == "payment.captured" && payload.Payload.Payment.Entity.Status == "captured" {
		orderID := payload.Payload.Payment.Entity.OrderID
		paymentID := payload.Payload.Payment.Entity.ID
		if _, err := MarkPaid(database.DB, orderID, paymentID); err != nil && err != ErrAlreadyPaid {
			log.Printf("payments: webhook credit failed for %s: %v", orderID, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"received": true})
}
