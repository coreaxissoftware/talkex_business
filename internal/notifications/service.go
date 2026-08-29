package notifications

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

var ErrNotFound = errors.New("notification not found")

// EmitInput is what other packages pass to Emit — kept small so callers
// don't have to know the DB row shape.
type EmitInput struct {
	OwnerID string
	Type    string
	Title   string
	Body    string
	Link    string
}

// EmitHook fires after a notification is persisted — used by the SSE
// hub in main.go to push a live update to the bell without polling.
type EmitHook func(ownerID string, n *Notification)

var emitHook EmitHook

// RegisterEmitHook wires a post-persist callback. Nil-safe.
func RegisterEmitHook(h EmitHook) { emitHook = h }

// Emit writes one notification for the given owner. Fire-and-forget on
// the caller's side — the caller shouldn't fail because logging a
// notification failed, so we swallow errors here after logging.
func Emit(db *gorm.DB, in EmitInput) {
	if in.Type == "" {
		in.Type = TypeInfo
	}
	n := &Notification{
		OwnerID: in.OwnerID,
		Type:    in.Type,
		Title:   in.Title,
		Body:    in.Body,
		Link:    in.Link,
	}
	if err := db.Create(n).Error; err != nil {
		return
	}
	if emitHook != nil {
		emitHook(in.OwnerID, n)
	}
}

// ListOptions narrows the notifications query.
type ListOptions struct {
	OwnerID    string
	UnreadOnly bool
	Limit      int
}

func List(db *gorm.DB, opts ListOptions) ([]Notification, error) {
	q := db.Where("owner_id = ?", opts.OwnerID)
	if opts.UnreadOnly {
		q = q.Where("read_at IS NULL")
	}
	limit := opts.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var out []Notification
	err := q.Order("created_at DESC").Limit(limit).Find(&out).Error
	return out, err
}

// UnreadCount is exposed so the bell badge can render without pulling
// the full list on every page.
func UnreadCount(db *gorm.DB, ownerID string) (int64, error) {
	var n int64
	err := db.Model(&Notification{}).
		Where("owner_id = ? AND read_at IS NULL", ownerID).
		Count(&n).Error
	return n, err
}

func MarkRead(db *gorm.DB, ownerID, id string) error {
	now := time.Now()
	res := db.Model(&Notification{}).
		Where("id = ? AND owner_id = ? AND read_at IS NULL", id, ownerID).
		Update("read_at", now)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		// Either the row doesn't exist, doesn't belong to this owner,
		// or was already read — either way, a no-op is fine.
		var count int64
		db.Model(&Notification{}).Where("id = ? AND owner_id = ?", id, ownerID).Count(&count)
		if count == 0 {
			return ErrNotFound
		}
	}
	return nil
}

func MarkAllRead(db *gorm.DB, ownerID string) error {
	now := time.Now()
	return db.Model(&Notification{}).
		Where("owner_id = ? AND read_at IS NULL", ownerID).
		Update("read_at", now).Error
}
