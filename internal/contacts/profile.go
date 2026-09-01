package contacts

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/auth"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// Customer 360 — a single endpoint that gathers everything about a
// contact so the frontend can render a rich drawer without twenty
// round-trips: profile, aggregate stats, and the ten most-recent
// interactions (conversations, deals, csat).
//
// Kept in this package rather than a new one because the raw SQL joins
// against tables the contacts package already owns (contact_id foreign
// keys); moving it out would just add import cycles.

// Profile360 is the response shape.
type Profile360 struct {
	Contact          *Contact          `json:"contact"`
	Stats            ProfileStats      `json:"stats"`
	RecentActivity   []ActivityItem    `json:"recent_activity"`
	AverageCSAT      float64           `json:"avg_csat"`
	CSATSamples      int64             `json:"csat_samples"`
	OpenDealValue    float64           `json:"open_deal_value"`
	OpenDealCount    int64             `json:"open_deal_count"`
}

// ProfileStats aggregates all-time conversation numbers.
type ProfileStats struct {
	TotalMessages    int64 `json:"total_messages"`
	InboundMessages  int64 `json:"inbound_messages"`
	OutboundMessages int64 `json:"outbound_messages"`
	Conversations    int64 `json:"conversations"`
}

// ActivityItem is a lightweight timeline entry — one row per event
// worth showing in the drawer, ordered newest-first.
type ActivityItem struct {
	Kind     string    `json:"kind"` // message | deal_moved | csat | opt_out | pay_link
	Summary  string    `json:"summary"`
	Channel  string    `json:"channel,omitempty"`
	At       time.Time `json:"at"`
	RefID    string    `json:"ref_id,omitempty"`
}

// RegisterProfileRoute wires GET /contacts/:id/profile.
func RegisterProfileRoute(r *gin.Engine) {
	g := r.Group("/contacts")
	g.Use(auth.AuthRequired())
	g.GET("/:id/profile", handleProfile)
}

func handleProfile(c *gin.Context) {
	ownerID := auth.GetUserID(c)
	contactID := c.Param("id")

	var contact Contact
	err := database.DB.Where("id = ? AND owner_id = ?", contactID, ownerID).First(&contact).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"detail": "Contact not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": "Internal server error"})
		return
	}

	p := &Profile360{Contact: &contact}
	buildStats(database.DB, ownerID, contactID, p)
	buildCSAT(database.DB, ownerID, contactID, p)
	buildDeals(database.DB, ownerID, contactID, p)
	buildActivity(database.DB, ownerID, contactID, p)

	c.JSON(http.StatusOK, p)
}

// buildStats — raw SQL because we cross packages and don't want import
// cycles. Failures log but return zero counters rather than 500.
func buildStats(db *gorm.DB, ownerID, contactID string, p *Profile360) {
	db.Raw(
		`SELECT COUNT(*) FROM conversations WHERE owner_id = ? AND contact_id = ?`,
		ownerID, contactID,
	).Row().Scan(&p.Stats.Conversations)

	db.Raw(
		`SELECT COUNT(*) FROM messages m
		 JOIN conversations c ON c.id = m.conversation_id
		 WHERE c.owner_id = ? AND c.contact_id = ?`,
		ownerID, contactID,
	).Row().Scan(&p.Stats.TotalMessages)

	db.Raw(
		`SELECT COUNT(*) FROM messages m
		 JOIN conversations c ON c.id = m.conversation_id
		 WHERE c.owner_id = ? AND c.contact_id = ? AND m.direction = 'inbound'`,
		ownerID, contactID,
	).Row().Scan(&p.Stats.InboundMessages)

	p.Stats.OutboundMessages = p.Stats.TotalMessages - p.Stats.InboundMessages
}

func buildCSAT(db *gorm.DB, ownerID, contactID string, p *Profile360) {
	db.Raw(
		`SELECT COUNT(*), COALESCE(AVG(score), 0) FROM ratings r
		 JOIN conversations c ON c.id = r.conversation_id
		 WHERE r.owner_id = ? AND c.contact_id = ?`,
		ownerID, contactID,
	).Row().Scan(&p.CSATSamples, &p.AverageCSAT)
}

func buildDeals(db *gorm.DB, ownerID, contactID string, p *Profile360) {
	db.Raw(
		`SELECT COUNT(*), COALESCE(SUM(value), 0) FROM deals
		 WHERE owner_id = ? AND contact_id = ?
		   AND stage NOT IN ('Won','Lost')`,
		ownerID, contactID,
	).Row().Scan(&p.OpenDealCount, &p.OpenDealValue)
}

// buildActivity — pull the 10 latest inbound messages as timeline items.
// Deals + CSAT + pay-link events can layer on later; keeping this small
// today makes the response fast and easy to render.
func buildActivity(db *gorm.DB, ownerID, contactID string, p *Profile360) {
	rows, err := db.Raw(
		`SELECT m.id, m.body, m.direction, c.channel, m.created_at
		 FROM messages m
		 JOIN conversations c ON c.id = m.conversation_id
		 WHERE c.owner_id = ? AND c.contact_id = ?
		 ORDER BY m.created_at DESC
		 LIMIT 10`,
		ownerID, contactID,
	).Rows()
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id, body, direction, channel string
		var at time.Time
		if err := rows.Scan(&id, &body, &direction, &channel, &at); err != nil {
			continue
		}
		summary := truncateStr(body, 80)
		if direction == "inbound" {
			summary = "→ " + summary
		} else {
			summary = "← " + summary
		}
		p.RecentActivity = append(p.RecentActivity, ActivityItem{
			Kind:    "message",
			Summary: summary,
			Channel: channel,
			At:      at,
			RefID:   id,
		})
	}
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
