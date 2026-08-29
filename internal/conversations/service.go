package conversations

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/contacts"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrContactNotFound      = errors.New("contact not found or not owned")
	ErrWindowClosed         = errors.New("24-hour window is closed; use a template")
)

// InboundHook fires after an inbound message is persisted. Other packages
// (automation) register handlers here to react without introducing an
// import cycle back into automation.
type InboundHook func(ownerID string, msg *Message, conv *Conversation)

// OutboundHook fires after an outbound message is persisted, used by the
// webhook layer to emit `message.status`.
type OutboundHook func(ownerID string, msg *Message, conv *Conversation)

var (
	inboundHooks  []InboundHook
	outboundHooks []OutboundHook
)

func RegisterInboundHook(h InboundHook) {
	inboundHooks = append(inboundHooks, h)
}

func RegisterOutboundHook(h OutboundHook) {
	outboundHooks = append(outboundHooks, h)
}

// ConversationWithContact is the shape the inbox actually wants: the
// conversation row alongside the contact's display fields, so the
// frontend doesn't have to fan out N contact lookups per render.
type ConversationWithContact struct {
	Conversation
	ContactName        *string `json:"contact_name"`
	ContactPhoneNumber string  `json:"contact_phone_number"`
	AssignedMemberName *string `json:"assigned_member_name,omitempty"`
}

// List returns every conversation for the owner, newest-message first,
// joined with the contact's name/phone for the inbox display.
func List(db *gorm.DB, ownerID string) ([]ConversationWithContact, error) {
	var out []ConversationWithContact
	err := db.
		Table("conversations").
		Select("conversations.*, contacts.name as contact_name, contacts.phone_number as contact_phone_number").
		Joins("LEFT JOIN contacts ON contacts.id = conversations.contact_id").
		Where("conversations.owner_id = ?", ownerID).
		Order("COALESCE(conversations.last_message_at, conversations.created_at) DESC").
		Scan(&out).Error
	return out, err
}

// Search returns conversations for an owner whose contact name, phone,
// labels, or any message body contains the query (case-insensitive).
// Kept simple: SQL LIKE across the joined text; scales fine for the
// dashboard-sized inbox and swaps for FTS/Elastic if a tenant grows.
func Search(db *gorm.DB, ownerID, query string) ([]ConversationWithContact, error) {
	q := strings.TrimSpace(query)
	if q == "" {
		return List(db, ownerID)
	}
	like := "%" + strings.ToLower(q) + "%"

	// Find conversation IDs whose messages match the query, so a
	// search for "invoice #42" finds a conversation even when the
	// contact name is unrelated.
	var msgConvIDs []string
	db.Model(&Message{}).
		Distinct("conversation_id").
		Where("LOWER(body) LIKE ?", like).
		Pluck("conversation_id", &msgConvIDs)

	var out []ConversationWithContact
	qb := db.
		Table("conversations").
		Select("conversations.*, contacts.name as contact_name, contacts.phone_number as contact_phone_number").
		Joins("LEFT JOIN contacts ON contacts.id = conversations.contact_id").
		Where("conversations.owner_id = ?", ownerID)

	if len(msgConvIDs) > 0 {
		qb = qb.Where(
			"LOWER(COALESCE(contacts.name,'')) LIKE ? OR LOWER(contacts.phone_number) LIKE ? OR LOWER(conversations.labels) LIKE ? OR conversations.id IN ?",
			like, like, like, msgConvIDs,
		)
	} else {
		qb = qb.Where(
			"LOWER(COALESCE(contacts.name,'')) LIKE ? OR LOWER(contacts.phone_number) LIKE ? OR LOWER(conversations.labels) LIKE ?",
			like, like, like,
		)
	}
	err := qb.Order("COALESCE(conversations.last_message_at, conversations.created_at) DESC").
		Limit(100).Scan(&out).Error
	return out, err
}

// BulkAssign assigns many conversations at once. Cross-tenant safe —
// the WHERE clause pins to ownerID so a rogue caller can't retag
// another tenant's threads.
func BulkAssign(db *gorm.DB, ownerID string, ids []string, agentUserID, agentName string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := db.Model(&Conversation{}).
		Where("owner_id = ? AND id IN ?", ownerID, ids).
		Updates(map[string]interface{}{
			"assigned_to":   agentUserID,
			"assigned_name": agentName,
		})
	return res.RowsAffected, res.Error
}

// BulkMarkRead zeros unread_count for many conversations.
func BulkMarkRead(db *gorm.DB, ownerID string, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	res := db.Model(&Conversation{}).
		Where("owner_id = ? AND id IN ?", ownerID, ids).
		Update("unread_count", 0)
	return res.RowsAffected, res.Error
}

// FindLatestByContact returns the most recent conversation for a
// contact regardless of channel — used by the AutoAwayLog dedup so
// we don't spam away messages every message.
func FindLatestByContact(db *gorm.DB, ownerID, contactID string) (*Conversation, error) {
	var c Conversation
	err := db.Where("owner_id = ? AND contact_id = ?", ownerID, contactID).
		Order("last_message_at DESC").First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func GetByID(db *gorm.DB, ownerID, id string) (*Conversation, error) {
	var c Conversation
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrConversationNotFound
	}
	return &c, err
}

// getOrCreate finds the (owner, contact, channel) conversation or creates
// it. Used by both the "send first outbound" path and the simulated
// inbound path.
func getOrCreate(db *gorm.DB, ownerID, contactID, channel string) (*Conversation, error) {
	var c Conversation
	err := db.Where("owner_id = ? AND contact_id = ? AND channel = ?", ownerID, contactID, channel).First(&c).Error
	if err == nil {
		return &c, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	c = Conversation{
		OwnerID:   ownerID,
		ContactID: contactID,
		Channel:   channel,
	}
	if err := db.Create(&c).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

// verifyContactOwned rejects sends targeting another tenant's contact.
func verifyContactOwned(db *gorm.DB, ownerID, contactID string) error {
	var count int64
	if err := db.Model(&contacts.Contact{}).
		Where("id = ? AND owner_id = ?", contactID, ownerID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrContactNotFound
	}
	return nil
}

// ListMessages returns all messages in a conversation, oldest-first (chat
// order — the frontend appends new ones at the bottom).
func ListMessages(db *gorm.DB, convID string) ([]Message, error) {
	var out []Message
	err := db.Where("conversation_id = ?", convID).Order("created_at ASC").Find(&out).Error
	return out, err
}

// SendOutbound appends an outbound message to the conversation. If the
// 24h window is closed a template_id is required.
type SendInput struct {
	ContactID  string  `json:"contact_id" binding:"required"`
	Channel    string  `json:"channel" binding:"required"`
	Body       string  `json:"body" binding:"required"`
	TemplateID *string `json:"template_id"`
}

func SendOutbound(db *gorm.DB, ownerID string, in *SendInput) (*Message, *Conversation, error) {
	if err := verifyContactOwned(db, ownerID, in.ContactID); err != nil {
		return nil, nil, err
	}
	conv, err := getOrCreate(db, ownerID, in.ContactID, in.Channel)
	if err != nil {
		return nil, nil, err
	}
	if !conv.IsWindowOpen() && in.TemplateID == nil {
		return nil, nil, ErrWindowClosed
	}

	now := time.Now()
	msg := &Message{
		ConversationID: conv.ID,
		Direction:      Outbound,
		Body:           in.Body,
		Status:         MsgStatusSent,
		TemplateID:     in.TemplateID,
	}
	if err := db.Create(msg).Error; err != nil {
		return nil, nil, err
	}

	conv.LastOutboundAt = &now
	conv.LastMessageAt = &now
	if err := db.Save(conv).Error; err != nil {
		return nil, nil, err
	}

	for _, h := range outboundHooks {
		func(hook OutboundHook) {
			defer func() { _ = recover() }()
			hook(ownerID, msg, conv)
		}(h)
	}

	return msg, conv, nil
}

// RecordInbound simulates a contact replying — the real channel connector
// will call this from a webhook once wired. Resets the 24h window and
// bumps unread_count for the inbox UI.
type InboundInput struct {
	ContactID string `json:"contact_id" binding:"required"`
	Channel   string `json:"channel" binding:"required"`
	Body      string `json:"body" binding:"required"`
}

func RecordInbound(db *gorm.DB, ownerID string, in *InboundInput) (*Message, *Conversation, error) {
	if err := verifyContactOwned(db, ownerID, in.ContactID); err != nil {
		return nil, nil, err
	}
	conv, err := getOrCreate(db, ownerID, in.ContactID, in.Channel)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	msg := &Message{
		ConversationID: conv.ID,
		Direction:      Inbound,
		Body:           in.Body,
		Status:         MsgStatusDelivered,
		DeliveredAt:    &now,
	}
	if err := db.Create(msg).Error; err != nil {
		return nil, nil, err
	}

	conv.LastInboundAt = &now
	conv.LastMessageAt = &now
	conv.UnreadCount++
	if err := db.Save(conv).Error; err != nil {
		return nil, nil, err
	}

	// Also refresh the contact's LastInboundAt so the Contacts page's
	// "Window Open" badge stays consistent with the conversation.
	_ = db.Model(&contacts.Contact{}).
		Where("id = ?", in.ContactID).
		Update("last_inbound_at", now).Error

	// Fire registered hooks (e.g. automation auto-reply). Recover per hook
	// so a broken listener can't kill the inbound path.
	for _, h := range inboundHooks {
		func(hook InboundHook) {
			defer func() { _ = recover() }()
			hook(ownerID, msg, conv)
		}(h)
	}

	return msg, conv, nil
}

type UpdateInput struct {
	Labels     *[]string `json:"labels"`
	AssignedTo *string   `json:"assigned_to"`
	AssignedName *string `json:"assigned_name"`
}

func UpdateConversation(db *gorm.DB, conv *Conversation, in *UpdateInput) (*Conversation, error) {
	if in.Labels != nil {
		b, _ := json.Marshal(*in.Labels)
		conv.Labels = string(b)
	}
	if in.AssignedTo != nil {
		if *in.AssignedTo == "" {
			conv.AssignedTo = nil
			conv.AssignedName = nil
		} else {
			conv.AssignedTo = in.AssignedTo
			if in.AssignedName != nil {
				conv.AssignedName = in.AssignedName
			}
		}
	}
	if err := db.Save(conv).Error; err != nil {
		return nil, err
	}
	return conv, nil
}

// MarkRead resets the unread counter — called when a user opens the thread.
func MarkRead(db *gorm.DB, conv *Conversation) error {
	if conv.UnreadCount == 0 {
		return nil
	}
	conv.UnreadCount = 0
	return db.Save(conv).Error
}
