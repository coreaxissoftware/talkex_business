// TalkEx Business API — Enterprise CPaaS dashboard.
// Run with: go run ./cmd/server
package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/coreaxissoftware/talkex_business/internal/analytics"
	"github.com/coreaxissoftware/talkex_business/internal/audit"
	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/automation"
	"github.com/coreaxissoftware/talkex_business/internal/billing"
	"github.com/coreaxissoftware/talkex_business/internal/campaigns"
	"github.com/coreaxissoftware/talkex_business/internal/channels"
	"github.com/coreaxissoftware/talkex_business/internal/compliance"
	"github.com/coreaxissoftware/talkex_business/internal/config"
	"github.com/coreaxissoftware/talkex_business/internal/contactlists"
	"github.com/coreaxissoftware/talkex_business/internal/customers"
	"github.com/coreaxissoftware/talkex_business/internal/customfields"
	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"github.com/coreaxissoftware/talkex_business/internal/conversations"
	"github.com/coreaxissoftware/talkex_business/internal/developers"
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"github.com/coreaxissoftware/talkex_business/internal/media"
	"github.com/coreaxissoftware/talkex_business/internal/messaging"
	"github.com/coreaxissoftware/talkex_business/internal/otp"

	// Channel connectors — imported for side-effect init() registration
	_ "github.com/coreaxissoftware/talkex_business/internal/channels/email"
	_ "github.com/coreaxissoftware/talkex_business/internal/channels/instagram"
	_ "github.com/coreaxissoftware/talkex_business/internal/channels/messenger"
	_ "github.com/coreaxissoftware/talkex_business/internal/channels/rcs"
	_ "github.com/coreaxissoftware/talkex_business/internal/channels/sandbox"
	_ "github.com/coreaxissoftware/talkex_business/internal/channels/telegram"
	_ "github.com/coreaxissoftware/talkex_business/internal/channels/talkex"
	waOnboarding "github.com/coreaxissoftware/talkex_business/internal/channels/whatsapp"
	"github.com/coreaxissoftware/talkex_business/internal/middleware"
	"github.com/coreaxissoftware/talkex_business/internal/notifications"
	"github.com/coreaxissoftware/talkex_business/internal/organizations"
	"github.com/coreaxissoftware/talkex_business/internal/quality"
	"github.com/coreaxissoftware/talkex_business/internal/settings"
	"github.com/coreaxissoftware/talkex_business/internal/support"
	"github.com/coreaxissoftware/talkex_business/internal/tags"
	"github.com/coreaxissoftware/talkex_business/internal/team"
	"github.com/coreaxissoftware/talkex_business/internal/templates"
	"github.com/coreaxissoftware/talkex_business/internal/users"
	"github.com/coreaxissoftware/talkex_business/internal/wallet"
	"github.com/coreaxissoftware/talkex_business/internal/webhooks"
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
		&notifications.Notification{},
		&webhooks.Endpoint{},
		&webhooks.Delivery{},
		&channels.Config{},
		&contactlists.ContactList{},
		&contactlists.ContactListMember{},
		&media.Media{},
		&team.Member{},
		&customfields.FieldDefinition{},
		&customers.Customer{},
		&messaging.QueuedMessage{},
		&messaging.DeadLetter{},
		&quality.Event{},
		&settings.UserSettings{},
		&auth.Session{},
		&waOnboarding.Onboarding{},
		&organizations.Organization{},
		&organizations.OrgMember{},
		&compliance.ConsentRecord{},
		&compliance.DSARRequest{},
		&compliance.ProcessingRecord{},
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
	r.Use(middleware.RateLimit(middleware.DefaultRateLimiterConfig()))
	r.Use(middleware.Idempotency())

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
	notifications.RegisterRoutes(r)
	webhooks.RegisterRoutes(r)
	channels.RegisterRoutes(r)
	contactlists.RegisterRoutes(r)
	media.RegisterRoutes(r)
	team.RegisterRoutes(r)
	customfields.RegisterRoutes(r)
	customers.RegisterRoutes(r)
	messaging.RegisterRoutes(r)
	quality.RegisterRoutes(r)
	settings.RegisterRoutes(r)
	tags.RegisterRoutes(r)
	organizations.RegisterRoutes(r)
	waOnboarding.RegisterRoutes(r)
	compliance.RegisterRoutes(r)
	otp.RegisterRoutes(r)
	auth.RegisterOAuthRoutes(r)
	channels.RegisterWebhookRoutes(r)

	// Wire OAuth user creator — lets auth.handleOAuthCallback find/create
	// users without importing the users package directly.
	auth.RegisterOAuthUserCreator(func(email, fullName, provider string) (string, error) {
		user, err := users.FindOrCreateOAuth(database.DB, email, fullName, provider)
		if err != nil {
			return "", err
		}
		return user.ID, nil
	})

	// Wire RBAC team role resolver — lets auth.RoleRequired() check
	// team membership for role-based access control.
	auth.RegisterTeamRoleResolver(func(userID string) (string, bool, string) {
		m, err := team.GetByUserID(database.DB, userID)
		if err != nil || m == nil {
			return "", false, ""
		}
		return m.Role, true, m.OwnerID
	})

	// Register API-key resolver with the auth package — lets any endpoint
	// guarded by auth.AuthRequired accept a plaintext API key in place of
	// a JWT access token (for server-to-server callers).
	auth.RegisterApiKeyResolver(func(raw string) (string, bool) {
		k, err := developers.ResolveKey(database.DB, raw)
		if err != nil || k == nil {
			return "", false
		}
		return k.OwnerID, true
	})

	// Campaigns need to reach conversations.SendOutbound without a package
	// cycle, so we inject the sender via the campaigns' registered SendFunc.
	campaigns.RegisterSender(func(ownerID, contactID, channel, body string, templateID *string) error {
		_, _, err := conversations.SendOutbound(database.DB, ownerID, &conversations.SendInput{
			ContactID:  contactID,
			Channel:    channel,
			Body:       body,
			TemplateID: templateID,
		})
		return err
	})

	campaigns.RegisterCompletionHook(func(ownerID string, c *campaigns.Campaign) {
		notifications.Emit(database.DB, notifications.EmitInput{
			OwnerID: ownerID,
			Type:    notifications.TypeSuccess,
			Title:   "Campaign completed: " + c.Name,
			Body:    fmt.Sprintf("Sent %d, failed %d of %d recipients.", c.SentCount, c.FailedCount, c.TotalCount),
			Link:    "/campaigns",
		})
		webhooks.Deliver(database.DB, ownerID, webhooks.EventCampaignCompleted, c)
	})

	// Contacts webhook fan-out on create — hook via a callback registered
	// with the contacts package (added in parallel with the runner work).
	contacts.RegisterCreateHook(func(ownerID string, c *contacts.Contact) {
		webhooks.Deliver(database.DB, ownerID, webhooks.EventContactCreated, c)
	})

	// Message-status changes (outbound flip to delivered/read) fire the
	// message.status webhook. Currently only 'sent' is stamped by our own
	// send path, so this fires with the initial status; real delivery
	// receipts will call the same conversations hook once wired.
	conversations.RegisterOutboundHook(func(ownerID string, msg *conversations.Message, conv *conversations.Conversation) {
		webhooks.Deliver(database.DB, ownerID, webhooks.EventMessageStatus, map[string]interface{}{
			"message":      msg,
			"conversation": conv,
		})
	})

	// On every inbound message we (a) fire matching automation rules,
	// (b) drop an in-app notification, and (c) deliver the event to any
	// subscribed outbound webhooks. Each is best-effort.
	conversations.RegisterInboundHook(func(ownerID string, msg *conversations.Message, conv *conversations.Conversation) {
		// (a) Automation — match a rule + reply. When the rule points to a
		// template, use the template body so the auto-reply respects the
		// author's approved copy; ResponseBody remains the free-form
		// fallback (and is also what shows when the 24h window is open).
		if rule, err := automation.FindMatching(database.DB, ownerID, msg.Body); err == nil && rule != nil {
			body := rule.ResponseBody
			if rule.TemplateID != nil {
				var tpl templates.MessageTemplate
				if err := database.DB.Where("id = ? AND owner_id = ?", *rule.TemplateID, ownerID).First(&tpl).Error; err == nil {
					body = tpl.Body
				}
			}
			_, _, err := conversations.SendOutbound(database.DB, ownerID, &conversations.SendInput{
				ContactID:  conv.ContactID,
				Channel:    conv.Channel,
				Body:       body,
				TemplateID: rule.TemplateID,
			})
			if err != nil {
				log.Printf("automation: rule %s reply failed: %v", rule.ID, err)
			} else {
				automation.BumpFireCount(database.DB, rule)
			}
		}

		// (b) In-app notification for the owner
		notifications.Emit(database.DB, notifications.EmitInput{
			OwnerID: ownerID,
			Type:    notifications.TypeInfo,
			Title:   "New message received",
			Body:    msg.Body,
			Link:    "/conversations",
		})

		// (c) Outbound webhook delivery
		webhooks.Deliver(database.DB, ownerID, webhooks.EventInboundMessage, map[string]interface{}{
			"message":      msg,
			"conversation": conv,
		})
	})

	// Wire campaign enqueuer through the messaging engine
	campaigns.RegisterEnqueuer(func(ownerID, campaignID, contactID, channel, body string, templateID *string) error {
		_, err := messaging.Enqueue(database.DB, &messaging.EnqueueInput{
			OwnerID:    ownerID,
			CampaignID: &campaignID,
			ContactID:  contactID,
			Channel:    channel,
			Body:       body,
			TemplateID: templateID,
			Priority:   messaging.PriorityMarketing,
		})
		return err
	})

	// Auto-resume paused campaigns when wallet is recharged above threshold
	wallet.RegisterCreditHook(func(ownerID string, newBalance float64) {
		_, prefs, err := settings.Get(database.DB, ownerID)
		if err != nil || !prefs.AutoPauseEnabled || prefs.MinBalance <= 0 {
			return
		}
		if newBalance >= prefs.MinBalance {
			resumed := campaigns.ResumeAllPaused(database.DB, ownerID)
			if resumed > 0 {
				notifications.Emit(database.DB, notifications.EmitInput{
					OwnerID: ownerID,
					Type:    notifications.TypeSuccess,
					Title:   "Campaigns auto-resumed",
					Body:    fmt.Sprintf("Wallet recharged to %.2f. %d paused campaign(s) resumed.", newBalance, resumed),
					Link:    "/campaigns",
				})
			}
		}
	})

	// Wire sandbox mode checker for messaging engine
	messaging.RegisterSandboxChecker(func(ownerID string) bool {
		_, prefs, err := settings.Get(database.DB, ownerID)
		if err != nil {
			return false
		}
		return prefs.SandboxMode
	})

	// Wire fallback channel resolver for messaging retry
	messaging.RegisterFallbackResolver(func(ownerID, contactID string) (string, bool) {
		var c contacts.Contact
		if err := database.DB.Where("id = ? AND owner_id = ?", contactID, ownerID).First(&c).Error; err != nil {
			return "", false
		}
		if c.FallbackChannel == nil || *c.FallbackChannel == "" {
			return "", false
		}
		return *c.FallbackChannel, true
	})

	// Wire wallet balance checker for messaging auto-pause
	messaging.RegisterWalletChecker(func(ownerID string) (float64, float64, bool) {
		w, err := wallet.GetOrCreateWallet(database.DB, ownerID)
		if err != nil {
			return 0, 0, false
		}
		_, prefs, err := settings.Get(database.DB, ownerID)
		if err != nil {
			return w.Balance, 0, false
		}
		return w.Balance, prefs.MinBalance, prefs.AutoPauseEnabled
	})

	messaging.RegisterPauseCallback(func(ownerID string, balance float64) {
		paused := campaigns.PauseAllRunning(database.DB, ownerID)
		if paused > 0 {
			notifications.Emit(database.DB, notifications.EmitInput{
				OwnerID: ownerID,
				Type:    notifications.TypeWarning,
				Title:   "Campaigns auto-paused",
				Body:    fmt.Sprintf("Wallet balance (%.2f) is below minimum threshold. %d campaign(s) paused.", balance, paused),
				Link:    "/wallet",
			})
		}
	})

	// Quality health alert hook — fires notifications + webhooks on status changes
	quality.RegisterAlertHook(func(ownerID, status string, count int64) {
		if status == "red" {
			notifications.Emit(database.DB, notifications.EmitInput{
				OwnerID: ownerID,
				Type:    notifications.TypeError,
				Title:   "⚠️ Critical: Number at risk of ban",
				Body:    fmt.Sprintf("Quality rating is RED with %d blocks/reports in 7 days. Messaging may be restricted.", count),
				Link:    "/analytics",
			})
			webhooks.Deliver(database.DB, ownerID, "quality.critical", map[string]interface{}{
				"status": status, "events_7d": count,
			})
		} else if status == "yellow" {
			notifications.Emit(database.DB, notifications.EmitInput{
				OwnerID: ownerID,
				Type:    notifications.TypeWarning,
				Title:   "Quality rating declining",
				Body:    fmt.Sprintf("You have %d blocks/reports in 7 days. Review your messaging to avoid restrictions.", count),
				Link:    "/analytics",
			})
		}
	})

	// Wire per-channel cost/sell price resolver for messaging engine
	messaging.RegisterCostResolver(func(ownerID, channel string) (float64, float64) {
		_, prefs, err := settings.Get(database.DB, ownerID)
		if err != nil {
			return 0, 0
		}
		switch channel {
		case "whatsapp":
			return prefs.CostWhatsapp, prefs.SellWhatsapp
		case "sms":
			return prefs.CostSMS, prefs.SellSMS
		case "talkex":
			return prefs.CostTalkex, prefs.SellTalkex
		case "telegram":
			return prefs.CostTelegram, prefs.SellTelegram
		case "email":
			return prefs.CostEmail, prefs.SellEmail
		case "rcs":
			return prefs.CostRCS, prefs.SellRCS
		case "instagram":
			return prefs.CostInstagram, prefs.SellInstagram
		case "messenger":
			return prefs.CostMessenger, prefs.SellMessenger
		default:
			return 0, 0
		}
	})

	// Wire campaign cost rollup from messaging engine
	campaigns.RegisterCostLookup(func(campaignID string) (float64, float64) {
		return messaging.GetCampaignCost(database.DB, campaignID)
	})

	// Wire contact list member getter for list-based campaign targeting
	campaigns.RegisterListMemberGetter(func(listID string) ([]string, error) {
		return contactlists.GetMembers(database.DB, listID)
	})

	// Wire maker-checker approval threshold
	campaigns.RegisterApprovalChecker(func(ownerID string) int {
		_, prefs, err := settings.Get(database.DB, ownerID)
		if err != nil {
			return 0
		}
		return prefs.ApprovalThreshold
	})

	// Background campaign scheduler — auto-launches campaigns when scheduled_at arrives
	campaigns.StartScheduler(database.DB)

	// Background messaging worker — processes queued messages every 5 seconds
	messaging.StartWorker(database.DB, 5*time.Second, 50)

	// Start
	addr := ":" + cfg.Port
	log.Printf("TalkEx Business API starting on %s (env=%s)", addr, cfg.Environment)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
