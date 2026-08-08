package audit

import (
	"gorm.io/gorm"
)

// Record persists one log entry. Called from the logging middleware after
// each request completes — failures here are swallowed by the caller so a
// logging problem never breaks the actual API response.
func Record(db *gorm.DB, entry *LogEntry) error {
	return db.Create(entry).Error
}

// ListFilter narrows the audit log query. Zero values are "no filter".
type ListFilter struct {
	UserID     string
	OnlyFailed bool
	Method     string
	Search     string // matches Path (LIKE)
	Limit      int
	Offset     int
}

// List returns log entries newest-first, scoped to the given owner unless
// the caller is listing platform-wide (UserID == "" means "all users",
// used by future admin views; today every caller passes their own ID).
func List(db *gorm.DB, f ListFilter) ([]LogEntry, int64, error) {
	q := db.Model(&LogEntry{})

	if f.UserID != "" {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.OnlyFailed {
		q = q.Where("success = ?", false)
	}
	if f.Method != "" {
		q = q.Where("method = ?", f.Method)
	}
	if f.Search != "" {
		q = q.Where("path LIKE ?", "%"+f.Search+"%")
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}

	var entries []LogEntry
	err := q.Order("created_at DESC").Limit(limit).Offset(f.Offset).Find(&entries).Error
	return entries, total, err
}

// Stats summarizes the last N log entries for a user — used by the
// dashboard's "API health" card.
type Stats struct {
	Total       int64 `json:"total"`
	Failed      int64 `json:"failed"`
	SuccessRate float64 `json:"success_rate"`
}

func GetStats(db *gorm.DB, userID string) (*Stats, error) {
	var total, failed int64
	base := db.Model(&LogEntry{}).Where("user_id = ?", userID)
	if err := base.Count(&total).Error; err != nil {
		return nil, err
	}
	if err := db.Model(&LogEntry{}).Where("user_id = ? AND success = ?", userID, false).Count(&failed).Error; err != nil {
		return nil, err
	}

	rate := 100.0
	if total > 0 {
		rate = float64(total-failed) / float64(total) * 100
	}
	return &Stats{Total: total, Failed: failed, SuccessRate: rate}, nil
}
