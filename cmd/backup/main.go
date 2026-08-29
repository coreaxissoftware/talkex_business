// Command backup exports a single tenant's data to a JSON file.
// Useful for support handoffs, DPDP DSAR fulfillment, and disaster
// recovery. Reads the same DATABASE_URL the server does.
//
// Usage:
//   go run ./cmd/backup -owner <user-id> [-out backup.json]
//   go run ./cmd/backup -owner <user-id> -restore < backup.json  (planned)
//
// Output shape is a single JSON object keyed by table name; each
// value is an array of rows (owner-scoped). Fields marked json:"-"
// on the models (e.g. hashed_password) are correctly omitted.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/audit"
	"github.com/coreaxissoftware/talkex_business/internal/automation"
	"github.com/coreaxissoftware/talkex_business/internal/campaigns"
	"github.com/coreaxissoftware/talkex_business/internal/canned"
	"github.com/coreaxissoftware/talkex_business/internal/channels"
	"github.com/coreaxissoftware/talkex_business/internal/compliance"
	"github.com/coreaxissoftware/talkex_business/internal/config"
	"github.com/coreaxissoftware/talkex_business/internal/contactlists"
	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"github.com/coreaxissoftware/talkex_business/internal/conversations"
	"github.com/coreaxissoftware/talkex_business/internal/csat"
	"github.com/coreaxissoftware/talkex_business/internal/customers"
	"github.com/coreaxissoftware/talkex_business/internal/customfields"
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"github.com/coreaxissoftware/talkex_business/internal/developers"
	"github.com/coreaxissoftware/talkex_business/internal/flows"
	"github.com/coreaxissoftware/talkex_business/internal/media"
	"github.com/coreaxissoftware/talkex_business/internal/notifications"
	"github.com/coreaxissoftware/talkex_business/internal/payments"
	"github.com/coreaxissoftware/talkex_business/internal/settings"
	"github.com/coreaxissoftware/talkex_business/internal/team"
	"github.com/coreaxissoftware/talkex_business/internal/templates"
	"github.com/coreaxissoftware/talkex_business/internal/users"
	"github.com/coreaxissoftware/talkex_business/internal/wallet"
	"github.com/coreaxissoftware/talkex_business/internal/webhooks"
	"github.com/coreaxissoftware/talkex_business/internal/widget"
)

func main() {
	owner := flag.String("owner", "", "owner (user) ID to back up")
	out := flag.String("out", "", "output file path (default: talkex-backup-<owner>-<ts>.json)")
	flag.Parse()

	if *owner == "" {
		fmt.Fprintln(os.Stderr, "usage: backup -owner <user-id> [-out FILE]")
		os.Exit(2)
	}

	cfg := config.Get()
	if err := database.Connect(cfg.DatabaseURL, cfg.IsDev()); err != nil {
		log.Fatalf("db connect: %v", err)
	}

	dump := map[string]interface{}{
		"exported_at": time.Now().Format(time.RFC3339),
		"owner_id":    *owner,
	}

	// Any query error just logs — we prefer a partial backup to no
	// backup, and the empty slice fields signal missing data.
	scan(database.DB, dump, "users", *owner, &[]users.User{}, "id = ?")
	scan(database.DB, dump, "contacts", *owner, &[]contacts.Contact{}, "owner_id = ?")
	scan(database.DB, dump, "contact_lists", *owner, &[]contactlists.ContactList{}, "owner_id = ?")
	scan(database.DB, dump, "contact_list_members", *owner, &[]contactlists.ContactListMember{}, "owner_id = ?")
	scan(database.DB, dump, "custom_fields", *owner, &[]customfields.FieldDefinition{}, "owner_id = ?")
	scan(database.DB, dump, "templates", *owner, &[]templates.MessageTemplate{}, "owner_id = ?")
	scan(database.DB, dump, "media", *owner, &[]media.Media{}, "owner_id = ?")
	scan(database.DB, dump, "campaigns", *owner, &[]campaigns.Campaign{}, "owner_id = ?")
	scan(database.DB, dump, "conversations", *owner, &[]conversations.Conversation{}, "owner_id = ?")
	// Messages have no owner_id column — filter through the conversations table.
	var convIDs []string
	database.DB.Model(&conversations.Conversation{}).
		Where("owner_id = ?", *owner).Pluck("id", &convIDs)
	if len(convIDs) > 0 {
		var msgs []conversations.Message
		database.DB.Where("conversation_id IN ?", convIDs).Find(&msgs)
		dump["messages"] = msgs
	} else {
		dump["messages"] = []conversations.Message{}
	}
	scan(database.DB, dump, "automation_rules", *owner, &[]automation.Rule{}, "owner_id = ?")
	scan(database.DB, dump, "flows", *owner, &[]flows.Flow{}, "owner_id = ?")
	scan(database.DB, dump, "flow_runs", *owner, &[]flows.RunState{}, "owner_id = ?")
	scan(database.DB, dump, "webhook_endpoints", *owner, &[]webhooks.Endpoint{}, "owner_id = ?")
	scan(database.DB, dump, "api_keys", *owner, &[]developers.ApiKey{}, "owner_id = ?")
	scan(database.DB, dump, "team_members", *owner, &[]team.Member{}, "owner_id = ?")
	scan(database.DB, dump, "channels", *owner, &[]channels.Config{}, "owner_id = ?")
	scan(database.DB, dump, "settings", *owner, &[]settings.UserSettings{}, "owner_id = ?")
	scan(database.DB, dump, "customer_profile", *owner, &[]customers.Customer{}, "owner_id = ?")
	scan(database.DB, dump, "wallet", *owner, &[]wallet.Wallet{}, "owner_id = ?")
	scan(database.DB, dump, "wallet_transactions", *owner, &[]wallet.WalletTransaction{}, "owner_id = ?")
	scan(database.DB, dump, "payments", *owner, &[]payments.Order{}, "owner_id = ?")
	// Tags aren't a first-class row — they live inside contacts.Tags JSON,
	// so the contacts dump above already includes them.
	scan(database.DB, dump, "canned_responses", *owner, &[]canned.Response{}, "owner_id = ?")
	scan(database.DB, dump, "csat_ratings", *owner, &[]csat.Rating{}, "owner_id = ?")
	scan(database.DB, dump, "notifications", *owner, &[]notifications.Notification{}, "owner_id = ?")
	scan(database.DB, dump, "audit_logs", *owner, &[]audit.LogEntry{}, "owner_id = ?")
	scan(database.DB, dump, "compliance_consents", *owner, &[]compliance.ConsentRecord{}, "owner_id = ?")
	scan(database.DB, dump, "compliance_dsars", *owner, &[]compliance.DSARRequest{}, "owner_id = ?")
	scan(database.DB, dump, "widget_config", *owner, &[]widget.Config{}, "owner_id = ?")
	scan(database.DB, dump, "widget_sessions", *owner, &[]widget.Session{}, "owner_id = ?")

	path := *out
	if path == "" {
		path = fmt.Sprintf("talkex-backup-%s-%s.json", *owner, time.Now().Format("20060102-150405"))
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatalf("create output: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(dump); err != nil {
		log.Fatalf("encode: %v", err)
	}
	fmt.Printf("wrote %s\n", path)
}

func scan[T any](db *gorm.DB, dst map[string]interface{}, name, owner string, into *[]T, whereClause string) {
	if err := db.Where(whereClause, owner).Find(into).Error; err != nil {
		log.Printf("%s: %v", name, err)
		dst[name] = []T{}
		return
	}
	dst[name] = *into
}
