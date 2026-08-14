package quality

import (
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/users"
)

const (
	flagThreshold = 5
	flagWindow    = 7 * 24 * time.Hour
)

// AlertHook is called when quality status changes (Yellow or Red).
type AlertHook func(ownerID, status string, blocksReports int64)

var alertHook AlertHook

func RegisterAlertHook(h AlertHook) { alertHook = h }

type Stats struct {
	Status         string     `json:"status"`
	BlocksLast7d   int64     `json:"blocks_last_7d"`
	ReportsLast7d  int64     `json:"reports_last_7d"`
	TotalBlocks    int64     `json:"total_blocks"`
	TotalReports   int64     `json:"total_reports"`
	FlaggedAt      *time.Time `json:"flagged_at"`
	Threshold      int       `json:"threshold"`
	HealthScore    int       `json:"health_score"`
}

func GetStats(db *gorm.DB, ownerID string) (*Stats, error) {
	user, err := users.GetByID(db, ownerID)
	if err != nil {
		return nil, err
	}

	cutoff := time.Now().Add(-flagWindow)

	var blocksRecent, reportsRecent, blocksTotal, reportsTotal int64
	db.Model(&Event{}).Where("owner_id = ? AND type = ? AND created_at >= ?", ownerID, EventBlock, cutoff).Count(&blocksRecent)
	db.Model(&Event{}).Where("owner_id = ? AND type = ? AND created_at >= ?", ownerID, EventReport, cutoff).Count(&reportsRecent)
	db.Model(&Event{}).Where("owner_id = ? AND type = ?", ownerID, EventBlock).Count(&blocksTotal)
	db.Model(&Event{}).Where("owner_id = ? AND type = ?", ownerID, EventReport).Count(&reportsTotal)

	status := users.QualityStatus(user)
	recentEvents := blocksRecent + reportsRecent

	// Health score: 100 = perfect, decreases with blocks/reports
	// Score = max(0, 100 - (recentEvents * 15))
	score := 100 - int(recentEvents)*15
	if score < 0 {
		score = 0
	}

	return &Stats{
		Status:        status,
		BlocksLast7d:  blocksRecent,
		ReportsLast7d: reportsRecent,
		TotalBlocks:   blocksTotal,
		TotalReports:  reportsTotal,
		FlaggedAt:     user.QualityFlaggedAt,
		Threshold:     flagThreshold,
		HealthScore:   score,
	}, nil
}

func RecordEvent(db *gorm.DB, ownerID, contactID, channel string, eventType EventType, reason *string) error {
	ev := &Event{
		OwnerID:   ownerID,
		ContactID: contactID,
		Channel:   channel,
		Type:      eventType,
		Reason:    reason,
	}
	if err := db.Create(ev).Error; err != nil {
		return err
	}

	if eventType == EventBlock || eventType == EventReport {
		refreshFlag(db, ownerID)
	}
	return nil
}

func refreshFlag(db *gorm.DB, ownerID string) {
	cutoff := time.Now().Add(-flagWindow)
	var count int64
	db.Model(&Event{}).
		Where("owner_id = ? AND type IN (?, ?) AND created_at >= ?", ownerID, EventBlock, EventReport, cutoff).
		Count(&count)

	if count >= flagThreshold {
		now := time.Now()
		db.Model(&users.User{}).Where("id = ?", ownerID).Update("quality_flagged_at", now)
		// Fire critical alert — Red status
		if alertHook != nil {
			alertHook(ownerID, "red", count)
		}
	} else if count >= 3 {
		// Warning alert at Yellow threshold (3+ events approaching limit)
		if alertHook != nil {
			alertHook(ownerID, "yellow", count)
		}
	}
}

func ListEvents(db *gorm.DB, ownerID string, limit int) ([]Event, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var out []Event
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}
