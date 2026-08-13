// TalkEx Business API — Enterprise CPaaS dashboard.
// Run with: go run ./cmd/server
package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/analytics"
	"github.com/coreaxissoftware/talkex_business/internal/audit"
	"github.com/coreaxissoftware/talkex_business/internal/automation"
	"github.com/coreaxissoftware/talkex_business/internal/billing"
	"github.com/coreaxissoftware/talkex_business/internal/campaigns"
	"github.com/coreaxissoftware/talkex_business/internal/config"
	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"github.com/coreaxissoftware/talkex_business/internal/conversations"
	"github.com/coreaxissoftware/talkex_business/internal/developers"
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"github.com/coreaxissoftware/talkex_business/internal/middleware"
	"github.com/coreaxissoftware/talkex_business/internal/support"
	"github.com/coreaxissoftware/talkex_business/internal/templates"
	"github.com/coreaxissoftware/talkex_business/internal/users"
	"github.com/coreaxissoftware/talkex_business/internal/wallet"
)

func main() {
	cfg := config.Get()

	// Database
	if err := database.Connect(cfg.DatabaseURL, cfg.IsDev()); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto-migrate all domain models
	if err := database.AutoMigrate(
		&users.User{},
		&wallet.Wallet{},
		&wallet.WalletTransaction{},
		&contacts.Contact{},
		&templates.MessageTemplate{},
		&audit.LogEntry{},
		&campaigns.Campaign{},
		&conversations.Conversation{},
		&conversations.Message{},
		&developers.ApiKey{},
		&automation.Rule{},
		&billing.Subscription{},
		&billing.Invoice{},
		&support.Ticket{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Router
	if !cfg.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())
	r.Use(gin.Logger())
	r.Use(audit.Middleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"environment": cfg.Environment,
		})
	})

	// Wire domain routes
	users.RegisterRoutes(r)
	wallet.RegisterRoutes(r)
	contacts.RegisterRoutes(r)
	templates.RegisterRoutes(r)
	audit.RegisterRoutes(r)
	campaigns.RegisterRoutes(r)
	conversations.RegisterRoutes(r)
	developers.RegisterRoutes(r)
	analytics.RegisterRoutes(r)
	automation.RegisterRoutes(r)
	billing.RegisterRoutes(r)
	support.RegisterRoutes(r)

	// Auto-reply: match inbound message body against the owner's rules
	// and, on the first match, send an outbound reply. Best-effort — a
	// rule-lookup or send error is logged and the inbound path continues.
	conversations.RegisterInboundHook(func(ownerID string, msg *conversations.Message, conv *conversations.Conversation) {
		rule, err := automation.FindMatching(database.DB, ownerID, msg.Body)
		if err != nil || rule == nil {
			return
		}
		_, _, err = conversations.SendOutbound(database.DB, ownerID, &conversations.SendInput{
			ContactID:  conv.ContactID,
			Channel:    conv.Channel,
			Body:       rule.ResponseBody,
			TemplateID: rule.TemplateID,
		})
		if err != nil {
			log.Printf("automation: rule %s reply failed: %v", rule.ID, err)
			return
		}
		automation.BumpFireCount(database.DB, rule)
	})

	// Start
	addr := ":" + cfg.Port
	log.Printf("TalkEx Business API starting on %s (env=%s)", addr, cfg.Environment)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
