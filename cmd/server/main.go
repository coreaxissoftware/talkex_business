// TalkEx Business API — Enterprise CPaaS dashboard.
// Run with: go run ./cmd/server
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/coreaxissoftware/talkex_business/internal/ai"
	"github.com/coreaxissoftware/talkex_business/internal/analytics"
	"github.com/coreaxissoftware/talkex_business/internal/audit"
	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/automation"
	"github.com/coreaxissoftware/talkex_business/internal/billing"
	"github.com/coreaxissoftware/talkex_business/internal/businesshours"
	"github.com/coreaxissoftware/talkex_business/internal/campaigns"
	"github.com/coreaxissoftware/talkex_business/internal/canned"
	"github.com/coreaxissoftware/talkex_business/internal/catalog"
	"github.com/coreaxissoftware/talkex_business/internal/channels"
	"github.com/coreaxissoftware/talkex_business/internal/csat"
	"github.com/coreaxissoftware/talkex_business/internal/compliance"
	"github.com/coreaxissoftware/talkex_business/internal/config"
	"github.com/coreaxissoftware/talkex_business/internal/contactlists"
	"github.com/coreaxissoftware/talkex_business/internal/customers"
	"github.com/coreaxissoftware/talkex_business/internal/customfields"
	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"github.com/coreaxissoftware/talkex_business/internal/conversations"
	"github.com/coreaxissoftware/talkex_business/internal/developers"
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"github.com/coreaxissoftware/talkex_business/internal/deals"
	"github.com/coreaxissoftware/talkex_business/internal/events"
	"github.com/coreaxissoftware/talkex_business/internal/flows"
	"github.com/coreaxissoftware/talkex_business/internal/greentick"
	"github.com/coreaxissoftware/talkex_business/internal/integrations"
	"github.com/coreaxissoftware/talkex_business/internal/media"
	"github.com/coreaxissoftware/talkex_business/internal/messaging"
	"github.com/coreaxissoftware/talkex_business/internal/metrics"
	"github.com/coreaxissoftware/talkex_business/internal/sla"
	"github.com/coreaxissoftware/talkex_business/internal/otp"
	"github.com/coreaxissoftware/talkex_business/internal/paylinks"
	"github.com/coreaxissoftware/talkex_business/internal/payments"
	"github.com/coreaxissoftware/talkex_business/internal/waflows"
	"github.com/coreaxissoftware/talkex_business/internal/redisclient"

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
	"github.com/coreaxissoftware/talkex_business/internal/widget"
)

func main() {
	cfg := config.Get()

	// Database
	if err := database.Connect(cfg.DatabaseURL, cfg.IsDev()); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Redis — optional. When REDIS_URL is set the rate limiter,
	// OTP store, and SSE hub switch to shared backends so multiple
	// API pods see the same state. Nil client = in-memory fallback.
	redisclient.Init(cfg.RedisURL)

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
		&canned.Response{},
		&catalog.Product{},
		&conversations.Reaction{},
		&deals.Pipeline{},
		&deals.Deal{},
		&greentick.Application{},
		&paylinks.PayLink{},
		&waflows.WAFlow{},
		&waflows.FlowResponse{},
		&csat.Rating{},
		&payments.Order{},
		&flows.Flow{},
		&flows.RunState{},
		&widget.Config{},
		&widget.Session{},
	); err != nil {
		log.Fatalf("Failed to auto-migrate: %v", err)
	}

	// Router
	if !cfg.IsDev() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	// Only trust Fly.io / Vercel edge; refuse to honor X-Forwarded-For
	// from arbitrary upstream hops otherwise.
	_ = r.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	// Cap multipart uploads at 32 MiB (media library enforces its own
	// tighter per-file limit).
	r.MaxMultipartMemory = 32 << 20

	r.Use(middleware.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	r.Use(middleware.CORS())
	r.Use(middleware.BodyLimit(10 << 20)) // 10 MiB JSON/form cap
	r.Use(middleware.ContentTypeGuard())
	r.Use(middleware.LoginBruteForceGuard())
	r.Use(metrics.Middleware())
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
	canned.RegisterRoutes(r)
	catalog.RegisterRoutes(r)
	deals.RegisterRoutes(r)
	greentick.RegisterRoutes(r)
	paylinks.RegisterRoutes(r)
	waflows.RegisterRoutes(r)
	contacts.RegisterProfileRoute(r)
	integrations.RegisterRoutes(r)
	integrations.RegisterSheetsRoutes(r)
	analytics.RegisterPDFRoute(r)
	conversations.RegisterReactionRoutes(r)
	csat.RegisterRoutes(r)
	payments.RegisterRoutes(r)
	flows.RegisterRoutes(r)
	ai.RegisterRoutes(r)
	events.RegisterRoutes(r)
	events.StartRedisFanout() // no-op when Redis is not configured
	widget.RegisterRoutes(r)
	metrics.RegisterRoutes(r)
	channels.RegisterWebhookRoutes(r)

	// Wire payments → wallet credit. Uses paymentID as idempotency key
	// so a duplicate webhook can't double-credit.
	payments.RegisterCreditFn(func(ownerID string, amount float64, reference, idempotencyKey string) error {
		w, err := wallet.GetOrCreateWallet(database.DB, ownerID)
		if err != nil {
			return err
		}
		refPtr := &reference
		_, err = wallet.ApplyTransaction(database.DB, w, wallet.Credit, amount, idempotencyKey, refPtr)
		return err
	})

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
		// (SSE) update any client that has this thread open on another device
		events.Publish(ownerID, events.TypeMessageOutbound, map[string]interface{}{
			"message":      msg,
			"conversation": conv,
		})
		webhooks.Deliver(database.DB, ownerID, webhooks.EventMessageStatus, map[string]interface{}{
			"message":      msg,
			"conversation": conv,
		})
	})

	// awaySentAt tracks when we last sent an away-message per contact
	// so an out-of-hours user pinging every minute doesn't get one
	// auto-reply per message.
	awaySentAt := struct {
		sync.Mutex
		m map[string]time.Time
	}{m: make(map[string]time.Time)}

	// On every inbound message we (a) fire matching automation rules,
	// (b) drop an in-app notification, and (c) deliver the event to any
	// subscribed outbound webhooks. Each is best-effort.
	conversations.RegisterInboundHook(func(ownerID string, msg *conversations.Message, conv *conversations.Conversation) {
		// (SSE) live update to the inbox / open thread
		events.Publish(ownerID, events.TypeMessageInbound, map[string]interface{}{
			"message":      msg,
			"conversation": conv,
		})

		// (0) Flow engine — first advance any active RunState for this
		// contact (branch steps waiting on a reply), then try to start a
		// new flow if a keyword-triggered flow matches the body.
		flows.HandleInbound(database.DB, ownerID, conv.ContactID, msg.Body)
		if f, err := flows.FindMatchingByKeyword(database.DB, ownerID, msg.Body); err == nil && f != nil {
			if _, err := flows.StartRun(database.DB, f, conv.ContactID, conv.Channel); err != nil {
				log.Printf("flow %s: start failed: %v", f.ID, err)
			}
		}

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

		// (d) Business-hours away message — outside configured window
		// send a one-off auto-reply and dedupe per contact for 24h so
		// a chatty visitor isn't spammed. Load owner prefs on demand.
		if _, prefs, err := settings.Get(database.DB, ownerID); err == nil && prefs.BusinessHoursEnabled && prefs.AwayMessage != "" {
			cfg := businesshours.Config{
				Enabled:   prefs.BusinessHoursEnabled,
				Days:      prefs.BusinessDays,
				OpenTime:  prefs.BusinessOpenTime,
				CloseTime: prefs.BusinessCloseTime,
				Timezone:  prefs.Timezone,
			}
			if !businesshours.IsOpen(cfg, time.Now()) {
				awaySentAt.Lock()
				last, seen := awaySentAt.m[conv.ContactID]
				should := !seen || time.Since(last) > 24*time.Hour
				if should {
					awaySentAt.m[conv.ContactID] = time.Now()
				}
				awaySentAt.Unlock()
				if should {
					_, _, err := conversations.SendOutbound(database.DB, ownerID, &conversations.SendInput{
						ContactID: conv.ContactID,
						Channel:   conv.Channel,
						Body:      prefs.AwayMessage,
					})
					if err != nil {
						log.Printf("away-message send failed: %v", err)
					}
				}
			}
		}

		// (e) AI auto-tag on inbound — classify sentiment, add label.
		// Best-effort: never blocks the inbound path; async via goroutine
		// so the visitor's ack isn't delayed by the Claude call.
		if _, prefs, err := settings.Get(database.DB, ownerID); err == nil && prefs.AIAutoTagEnabled {
			go func(oid, cid, cid2 string) {
				res, err := ai.Sentiment(context.Background(), []ai.Message{{Direction: "inbound", Body: msg.Body}})
				if err != nil || res == nil || res.Score == "" {
					return
				}
				// Merge into the conversation's labels JSON array.
				var conv2 conversations.Conversation
				if err := database.DB.Where("id = ?", cid).First(&conv2).Error; err != nil {
					return
				}
				var labels []string
				_ = json.Unmarshal([]byte(conv2.Labels), &labels)
				label := "sentiment:" + res.Score
				// Drop any prior sentiment: label so the tag reflects
				// only the latest inbound message.
				out := []string{label}
				for _, l := range labels {
					if !strings.HasPrefix(l, "sentiment:") {
						out = append(out, l)
					}
				}
				_, _ = conversations.UpdateConversation(database.DB, &conv2, &conversations.UpdateInput{Labels: &out})
			}(ownerID, conv.ID, "")
		}
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

	// Wire flow engine — send/assign/tag callbacks
	flows.RegisterSender(func(ownerID, contactID, channel, body string, templateID *string) error {
		_, _, err := conversations.SendOutbound(database.DB, ownerID, &conversations.SendInput{
			ContactID:  contactID,
			Channel:    channel,
			Body:       body,
			TemplateID: templateID,
		})
		return err
	})
	flows.RegisterAssigner(func(ownerID, contactID, agentUserID string) error {
		// Find the most recent conversation for this contact (any channel)
		var conv conversations.Conversation
		if err := database.DB.Where("owner_id = ? AND contact_id = ?", ownerID, contactID).
			Order("last_message_at DESC").First(&conv).Error; err != nil {
			return err
		}
		name := agentUserID
		if m, err := team.GetByUserID(database.DB, agentUserID); err == nil && m != nil && m.Name != "" {
			name = m.Name
		}
		_, err := conversations.UpdateConversation(database.DB, &conv, &conversations.UpdateInput{
			AssignedTo:   &agentUserID,
			AssignedName: &name,
		})
		return err
	})
	flows.RegisterTagger(func(ownerID, contactID, tag string) error {
		ct, err := contacts.GetByID(database.DB, ownerID, contactID)
		if err != nil {
			return err
		}
		var tags []string
		_ = json.Unmarshal(ct.Tags, &tags)
		for _, t := range tags {
			if t == tag {
				return nil // already present
			}
		}
		tags = append(tags, tag)
		_, err = contacts.Update(database.DB, ct, &contacts.UpdateInput{Tags: &tags})
		return err
	})

	// SSE fan-out for notifications — live bell without polling
	notifications.RegisterEmitHook(func(ownerID string, n *notifications.Notification) {
		events.Publish(ownerID, events.TypeNotificationNew, n)
	})

	// Background flow sweeper — advances waiting steps every 30s
	flows.StartSweeper(database.DB, 30*time.Second)

	// SLA breach alerts — check every minute, fires notification +
	// webhook per breach with 12h re-alert cooldown.
	sla.RegisterPrefsFetcher(func(ownerID string) int {
		_, prefs, err := settings.Get(database.DB, ownerID)
		if err != nil {
			return 0
		}
		return prefs.SLAFirstResponseMins
	})
	sla.RegisterNotifier(func(ownerID, title, body, link string) {
		notifications.Emit(database.DB, notifications.EmitInput{
			OwnerID: ownerID, Type: notifications.TypeWarning,
			Title: title, Body: body, Link: link,
		})
	})
	sla.RegisterWebhook(func(ownerID, event string, payload interface{}) {
		webhooks.Deliver(database.DB, ownerID, event, payload)
	})
	sla.Start(database.DB, 1*time.Minute)

	// OTP reaper — expires abandoned in-memory entries so the store
	// doesn't grow unbounded across dropped signups.
	otp.StartCleanup(5 * time.Minute)

	// -------------------------------------------------------------
	// Widget wires — a live-chat visitor becomes a Contact of channel
	// "webchat", the same conversations.RecordInbound path fires the
	// existing hooks, and outbound agent replies stream back via the
	// events hub filtered on conversation ID.
	// -------------------------------------------------------------
	widget.RegisterContactCreator(func(ownerID, name, email string) (string, error) {
		// A widget contact needs a unique "phone number" — reuse the
		// Contact model by giving each session a synthetic web:<uuid>
		// identifier scoped by our unique index on (owner_id, phone).
		phone := "web:" + uuid.New().String()
		in := &contacts.CreateInput{
			PhoneNumber: phone,
			Name:        &name,
			Tags:        []string{"webchat"},
		}
		if email != "" {
			in.Email = &email
		}
		c, err := contacts.Create(database.DB, ownerID, in)
		if err != nil {
			return "", err
		}
		return c.ID, nil
	})
	widget.RegisterInboundRecorder(func(ownerID, contactID, body string) error {
		_, _, err := conversations.RecordInbound(database.DB, ownerID, &conversations.InboundInput{
			ContactID: contactID,
			Channel:   "webchat",
			Body:      body,
		})
		return err
	})
	widget.RegisterConvLookup(func(ownerID, contactID string) (string, error) {
		var conv conversations.Conversation
		err := database.DB.Where("owner_id = ? AND contact_id = ? AND channel = ?",
			ownerID, contactID, "webchat").
			Order("last_message_at DESC").First(&conv).Error
		if err != nil {
			return "", err
		}
		return conv.ID, nil
	})
	widget.RegisterMessageLister(func(conversationID string) ([]widget.OutboundMsg, error) {
		msgs, err := conversations.ListMessages(database.DB, conversationID)
		if err != nil {
			return nil, err
		}
		out := make([]widget.OutboundMsg, len(msgs))
		for i, m := range msgs {
			out[i] = widget.OutboundMsg{
				ID:        m.ID,
				Direction: m.Direction,
				Body:      m.Body,
				CreatedAt: m.CreatedAt,
			}
		}
		return out, nil
	})
	// Streamer bridges the owner-scoped events hub to a per-conversation
	// channel that the widget's SSE handler drains.
	widget.RegisterOutboundStreamer(func(conversationID string) (<-chan widget.OutboundMsg, func()) {
		// Resolve owner ID once — we need it to subscribe.
		var conv conversations.Conversation
		if err := database.DB.Where("id = ?", conversationID).First(&conv).Error; err != nil {
			ch := make(chan widget.OutboundMsg)
			close(ch)
			return ch, func() {}
		}
		out := make(chan widget.OutboundMsg, 8)
		sub := events.Subscribe(conv.OwnerID)
		done := make(chan struct{})
		go func() {
			defer close(out)
			for {
				select {
				case <-done:
					return
				case ev, ok := <-sub.Channel():
					if !ok {
						return
					}
					if ev.Type != events.TypeMessageOutbound && ev.Type != events.TypeMessageInbound {
						continue
					}
					payload, _ := ev.Data.(map[string]interface{})
					if payload == nil {
						continue
					}
					msg, _ := payload["message"].(*conversations.Message)
					c2, _ := payload["conversation"].(*conversations.Conversation)
					if msg == nil || c2 == nil || c2.ID != conversationID {
						continue
					}
					select {
					case out <- widget.OutboundMsg{
						ID:        msg.ID,
						Direction: msg.Direction,
						Body:      msg.Body,
						CreatedAt: msg.CreatedAt,
					}:
					default:
					}
				}
			}
		}()
		cancel := func() {
			close(done)
			events.Unsubscribe(sub)
		}
		return out, cancel
	})

	// CSAT ownership resolver — canonicalizes contact_id + channel
	// and rejects submissions for conversations the caller doesn't own.
	csat.RegisterConversationOwnerCheck(func(conversationID string) (string, string, string, error) {
		var conv conversations.Conversation
		if err := database.DB.Where("id = ?", conversationID).First(&conv).Error; err != nil {
			return "", "", "", err
		}
		return conv.OwnerID, conv.ContactID, conv.Channel, nil
	})

	// Wire AI conversation fetcher — pulls last 30 messages + contact name
	ai.RegisterConversationFetcher(func(ownerID, conversationID string) (string, []ai.Message, error) {
		conv, err := conversations.GetByID(database.DB, ownerID, conversationID)
		if err != nil {
			return "", nil, err
		}
		msgs, err := conversations.ListMessages(database.DB, conv.ID)
		if err != nil {
			return "", nil, err
		}
		// Keep at most last 30 turns for context — enough for continuity, under our token cap
		if len(msgs) > 30 {
			msgs = msgs[len(msgs)-30:]
		}
		out := make([]ai.Message, len(msgs))
		for i, m := range msgs {
			out[i] = ai.Message{Direction: m.Direction, Body: m.Body, CreatedAt: m.CreatedAt}
		}
		// Resolve contact name
		name := "customer"
		if ct, err := contacts.GetByID(database.DB, ownerID, conv.ContactID); err == nil && ct != nil {
			if ct.Name != nil && *ct.Name != "" {
				name = *ct.Name
			} else {
				name = ct.PhoneNumber
			}
		}
		return name, out, nil
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
