package auth

import (
	"time"

	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/gorm"
)

type Session struct {
	database.Base
	UserID    string     `gorm:"type:varchar(36);index;not null" json:"user_id"`
	TokenJTI  string     `gorm:"type:varchar(36);uniqueIndex;not null" json:"-"`
	IPAddress string     `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent string     `gorm:"type:varchar(500)" json:"user_agent"`
	ExpiresAt time.Time  `json:"expires_at"`
	RevokedAt *time.Time `json:"revoked_at"`
}

func (s *Session) IsActive() bool {
	return s.RevokedAt == nil && s.ExpiresAt.After(time.Now())
}

func CreateSession(db *gorm.DB, userID, tokenJTI, ip, ua string, expiresAt time.Time) (*Session, error) {
	s := &Session{
		UserID:    userID,
		TokenJTI:  tokenJTI,
		IPAddress: ip,
		UserAgent: ua,
		ExpiresAt: expiresAt,
	}
	return s, db.Create(s).Error
}

func ListSessions(db *gorm.DB, userID string) ([]Session, error) {
	var out []Session
	err := db.Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Order("created_at DESC").Find(&out).Error
	return out, err
}

func RevokeSession(db *gorm.DB, userID, sessionID string) error {
	now := time.Now()
	return db.Model(&Session{}).
		Where("id = ? AND user_id = ?", sessionID, userID).
		Update("revoked_at", now).Error
}

func RevokeAllSessions(db *gorm.DB, userID, exceptSessionID string) error {
	now := time.Now()
	q := db.Model(&Session{}).Where("user_id = ? AND revoked_at IS NULL", userID)
	if exceptSessionID != "" {
		q = q.Where("id != ?", exceptSessionID)
	}
	return q.Update("revoked_at", now).Error
}
