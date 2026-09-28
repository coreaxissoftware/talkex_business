package platformenv

import (
	"errors"
	"os"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// Allowed pins the exact keys the UI can write. A superadmin cannot
// invent arbitrary env-var names — that would let one poison PATH,
// LD_PRELOAD, or GORM's internal knobs through the console. Anything
// not on this list is refused with 422.
//
// The set matches internal/config exactly; when a new provider is
// added to config, add its keys here so the operator can wire it
// through the UI without a redeploy.
var Allowed = map[string]struct {
	Category string
	Label    string
	// Sensitive means the value should be masked on read. False for
	// public-safe values like OAuth *client IDs*, MSG91 sender IDs,
	// Mailgun From lines etc.
	Sensitive bool
}{
	// Email
	"MAILGUN_DOMAIN":   {"email", "Mailgun Domain", false},
	"MAILGUN_API_KEY":  {"email", "Mailgun API Key", true},
	"MAILGUN_FROM":     {"email", "From line", false},
	"MAILGUN_BASE_URL": {"email", "Mailgun API base (EU)", false},

	// SMS — India
	"MSG91_AUTH_KEY":       {"sms", "MSG91 auth key", true},
	"MSG91_TEMPLATE_ID":    {"sms", "MSG91 template id", false},
	"MSG91_SENDER_ID":      {"sms", "MSG91 sender id", false},
	"MSG91_ROUTE":          {"sms", "MSG91 route", false},
	"FAST2SMS_API_KEY":     {"sms", "Fast2SMS API key", true},
	"FAST2SMS_SENDER_ID":   {"sms", "Fast2SMS sender id", false},
	"FAST2SMS_TEMPLATE_ID": {"sms", "Fast2SMS template id", false},
	"FAST2SMS_ROUTE":       {"sms", "Fast2SMS route", false},

	// SMS — global
	"TWILIO_ACCOUNT_SID": {"sms", "Twilio account SID", false},
	"TWILIO_AUTH_TOKEN":  {"sms", "Twilio auth token", true},
	"TWILIO_FROM_NUMBER": {"sms", "Twilio from number", false},

	// Payments
	"RAZORPAY_KEY_ID":         {"payments", "Razorpay key id", false},
	"RAZORPAY_SECRET":         {"payments", "Razorpay secret", true},
	"RAZORPAY_WEBHOOK_SECRET": {"payments", "Razorpay webhook secret", true},

	// OAuth
	"OAUTH_GOOGLE_CLIENT_ID":   {"oauth", "Google client id", false},
	"OAUTH_GOOGLE_SECRET":      {"oauth", "Google secret", true},
	"OAUTH_FACEBOOK_CLIENT_ID": {"oauth", "Facebook client id", false},
	"OAUTH_FACEBOOK_SECRET":    {"oauth", "Facebook secret", true},
	"OAUTH_GITHUB_CLIENT_ID":   {"oauth", "GitHub client id", false},
	"OAUTH_GITHUB_SECRET":      {"oauth", "GitHub secret", true},
	"OAUTH_APPLE_CLIENT_ID":    {"oauth", "Apple client id", false},
	"OAUTH_APPLE_TEAM_ID":      {"oauth", "Apple team id", false},
	"OAUTH_APPLE_KEY_ID":       {"oauth", "Apple key id", false},
	"OAUTH_APPLE_PRIVATE_KEY":  {"oauth", "Apple .p8 PEM", true},
}

// cache holds the DB overrides in memory so Get() doesn't hit Postgres
// on every provider call. Rebuilt by Load() at startup and after every
// Set().
var (
	cacheMu sync.RWMutex
	cache   = map[string]string{}
)

// Load reads every row from platform_settings and hydrates the cache.
// Also calls os.Setenv for each so callers that were already read
// through envOr at boot (config.Get, sync.Once-cached) see the DB
// value on the next config.Reload().
//
// Safe to call before the schema exists — AutoMigrate hasn't run yet
// on very first boot. Returns nil in that case; the cache stays empty
// and Get() falls back to env.
func Load(db *gorm.DB) error {
	if !db.Migrator().HasTable(&Setting{}) {
		return nil
	}
	var rows []Setting
	if err := db.Find(&rows).Error; err != nil {
		return err
	}
	next := map[string]string{}
	for _, r := range rows {
		key := strings.ToUpper(r.Key)
		if _, ok := Allowed[key]; !ok {
			continue // silently drop unknown keys; UI can't create them anymore
		}
		next[key] = r.Value
		_ = os.Setenv(key, r.Value)
	}
	cacheMu.Lock()
	cache = next
	cacheMu.Unlock()
	return nil
}

// Get returns the effective value for a platform env key: DB override
// first, then process env. Never returns an error — a missing key is
// just the empty string, matching envOr's semantics.
func Get(key string) string {
	key = strings.ToUpper(key)
	cacheMu.RLock()
	if v, ok := cache[key]; ok {
		cacheMu.RUnlock()
		return v
	}
	cacheMu.RUnlock()
	return os.Getenv(key)
}

// SetInput is the body of a superadmin PUT.
type SetInput struct {
	Value string `json:"value"`
}

// ErrUnknownKey is returned when the caller tries to write a key that
// isn't on the Allowed list.
var ErrUnknownKey = errors.New("platform_settings: key is not on the allow-list")

// Set upserts a row and refreshes the cache. Empty value is treated as
// "clear the override" — the row is deleted so Get() falls back to
// process env.
func Set(db *gorm.DB, key, value, setBy string) error {
	key = strings.ToUpper(key)
	if _, ok := Allowed[key]; !ok {
		return ErrUnknownKey
	}
	if value == "" {
		if err := db.Where("key = ?", key).Delete(&Setting{}).Error; err != nil {
			return err
		}
		cacheMu.Lock()
		delete(cache, key)
		cacheMu.Unlock()
		_ = os.Unsetenv(key)
		return nil
	}
	var existing Setting
	err := db.Where("key = ?", key).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		row := Setting{
			Key:           key,
			Value:         value,
			SetBy:         setBy,
			LastRotatedAt: time.Now(),
		}
		if err := db.Create(&row).Error; err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		updates := map[string]interface{}{
			"value":  value,
			"set_by": setBy,
		}
		if existing.Value != value {
			updates["last_rotated_at"] = time.Now()
		}
		if err := db.Model(&existing).Updates(updates).Error; err != nil {
			return err
		}
	}
	cacheMu.Lock()
	cache[key] = value
	cacheMu.Unlock()
	_ = os.Setenv(key, value)
	return nil
}

// ListItem is one entry as returned by the admin list endpoint —
// masked value only, never the raw secret.
type ListItem struct {
	Key           string    `json:"key"`
	Category      string    `json:"category"`
	Label         string    `json:"label"`
	Sensitive     bool      `json:"sensitive"`
	Set           bool      `json:"set"`             // true if the DB row exists
	FromEnv       bool      `json:"from_env"`        // true if the effective value comes from env not DB
	Masked        string    `json:"masked"`          // "••••1234" for sensitive, real value for public
	LastRotatedAt time.Time `json:"last_rotated_at"` // zero if never
}

// List returns every allowed key with its current mask + source.
// Values are masked so a compromised support session cannot exfiltrate
// raw secrets. Ordering: category, then key — stable for the UI.
func List(db *gorm.DB) ([]ListItem, error) {
	dbRows := map[string]Setting{}
	if db.Migrator().HasTable(&Setting{}) {
		var rows []Setting
		if err := db.Find(&rows).Error; err != nil {
			return nil, err
		}
		for _, r := range rows {
			dbRows[strings.ToUpper(r.Key)] = r
		}
	}
	out := make([]ListItem, 0, len(Allowed))
	for k, meta := range Allowed {
		item := ListItem{Key: k, Category: meta.Category, Label: meta.Label, Sensitive: meta.Sensitive}
		row, hasRow := dbRows[k]
		effective := ""
		if hasRow && row.Value != "" {
			effective = row.Value
			item.Set = true
			item.FromEnv = false
			item.LastRotatedAt = row.LastRotatedAt
		} else if envVal := os.Getenv(k); envVal != "" {
			effective = envVal
			item.Set = true
			item.FromEnv = true
		}
		item.Masked = mask(effective, meta.Sensitive)
		out = append(out, item)
	}
	// Stable sort: category, then key.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0; j-- {
			a, b := out[j-1], out[j]
			if a.Category > b.Category || (a.Category == b.Category && a.Key > b.Key) {
				out[j-1], out[j] = b, a
			}
		}
	}
	return out, nil
}

// mask returns a UI-safe rendering. Public values (client IDs, sender
// IDs) show verbatim so an operator can eyeball them; sensitive
// values show last-4 only, or bullets when shorter.
func mask(v string, sensitive bool) string {
	if v == "" {
		return ""
	}
	if !sensitive {
		return v
	}
	if len(v) <= 4 {
		return "••••"
	}
	return "••••" + v[len(v)-4:]
}

// EnsureSchema runs the AutoMigrate for platform_settings. Called from
// server bootstrap alongside the other module migrations.
func EnsureSchema() error {
	return database.DB.AutoMigrate(&Setting{})
}
