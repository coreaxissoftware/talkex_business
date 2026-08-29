package templates

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/gorm"
)

var ErrTemplateNotFound = errors.New("template not found")

func List(db *gorm.DB, ownerID string) ([]MessageTemplate, error) {
	var tpls []MessageTemplate
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&tpls).Error
	return tpls, err
}

func GetByID(db *gorm.DB, ownerID, templateID string) (*MessageTemplate, error) {
	var t MessageTemplate
	err := db.Where("id = ? AND owner_id = ?", templateID, ownerID).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrTemplateNotFound
	}
	return &t, err
}

// Button is one quick-reply / URL / phone button. Meta permits up to
// 3 buttons per template.
type Button struct {
	Type  string `json:"type"`  // quick_reply | url | phone
	Text  string `json:"text"`
	URL   string `json:"url,omitempty"`
	Phone string `json:"phone,omitempty"`
}

// ListRow is one row inside a list picker section.
type ListRow struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

type CreateInput struct {
	Name      string           `json:"name" binding:"required"`
	Category  TemplateCategory `json:"category" binding:"required"`
	Channel   string           `json:"channel" binding:"required"`
	Body      string           `json:"body" binding:"required"`
	Variables []string         `json:"variables"`
	Buttons   []Button         `json:"buttons"`
	ListRows  []ListRow        `json:"list_rows"`
	Header    string           `json:"header"`
	Footer    string           `json:"footer"`
	MediaType string           `json:"media_type"`
	MediaURL  string           `json:"media_url"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*MessageTemplate, error) {
	varsJSON, _ := json.Marshal(in.Variables)
	if in.Variables == nil {
		varsJSON = []byte("[]")
	}

	btnJSON, _ := json.Marshal(nonNilButtons(in.Buttons))
	rowJSON, _ := json.Marshal(nonNilRows(in.ListRows))

	t := &MessageTemplate{
		OwnerID:   ownerID,
		Name:      in.Name,
		Category:  in.Category,
		Channel:   in.Channel,
		Body:      in.Body,
		Variables: varsJSON,
		Buttons:   btnJSON,
		ListRows:  rowJSON,
		Header:    in.Header,
		Footer:    in.Footer,
		MediaType: in.MediaType,
		MediaURL:  in.MediaURL,
	}
	if err := db.Create(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func nonNilButtons(b []Button) []Button {
	if b == nil {
		return []Button{}
	}
	return b
}

func nonNilRows(r []ListRow) []ListRow {
	if r == nil {
		return []ListRow{}
	}
	return r
}

type UpdateInput struct {
	Name      *string         `json:"name"`
	Body      *string         `json:"body"`
	Variables *[]string       `json:"variables"`
	Status    *TemplateStatus `json:"status"`
	Buttons   *[]Button       `json:"buttons"`
	ListRows  *[]ListRow      `json:"list_rows"`
	Header    *string         `json:"header"`
	Footer    *string         `json:"footer"`
	MediaType *string         `json:"media_type"`
	MediaURL  *string         `json:"media_url"`
}

func Update(db *gorm.DB, t *MessageTemplate, in *UpdateInput) (*MessageTemplate, error) {
	if in.Name != nil {
		t.Name = *in.Name
	}
	if in.Body != nil {
		t.Body = *in.Body
	}
	if in.Variables != nil {
		varsJSON, _ := json.Marshal(*in.Variables)
		t.Variables = varsJSON
	}
	if in.Status != nil {
		t.Status = *in.Status
	}
	if in.Buttons != nil {
		b, _ := json.Marshal(nonNilButtons(*in.Buttons))
		t.Buttons = b
	}
	if in.ListRows != nil {
		r, _ := json.Marshal(nonNilRows(*in.ListRows))
		t.ListRows = r
	}
	if in.Header != nil {
		t.Header = *in.Header
	}
	if in.Footer != nil {
		t.Footer = *in.Footer
	}
	if in.MediaType != nil {
		t.MediaType = *in.MediaType
	}
	if in.MediaURL != nil {
		t.MediaURL = *in.MediaURL
	}
	if err := db.Save(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

// SubmitToMeta pushes a WhatsApp template to Meta for approval.
// Production: POST https://graph.facebook.com/v18.0/<wabaID>/message_templates
// with the template body + components. Dev mode (no META token): mark
// pending_review locally, log the payload, and stamp a fake ExternalRef.
func SubmitToMeta(db *gorm.DB, t *MessageTemplate) (*MessageTemplate, error) {
	token := os.Getenv("META_WHATSAPP_TOKEN")
	wabaID := os.Getenv("META_WHATSAPP_WABA_ID")

	now := time.Now().Unix()
	t.SubmittedAt = &now
	t.Status = StatusPendingReview
	t.ExternalStatus = "PENDING"
	t.RejectReason = ""

	if token == "" || wabaID == "" {
		// Dev mode — log the shape and skip the HTTP call.
		t.ExternalRef = "dev_" + t.ID
		log.Printf("templates (dev): would submit template %q (%s) to Meta WABA %q", t.Name, t.ID, wabaID)
		if err := db.Save(t).Error; err != nil {
			return nil, err
		}
		return t, nil
	}

	// Production would build the components array here and POST it.
	// Left as a TODO — the shape depends on which fields the template
	// uses (header/body/footer/buttons/media) and Meta rejects payloads
	// with mismatched components.
	log.Printf("templates: production Meta submission not yet implemented for template %s", t.ID)
	if err := db.Save(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func Delete(db *gorm.DB, t *MessageTemplate) error {
	return db.Delete(t).Error
}

// RenderBody substitutes {{1}}, {{2}}, etc. in the template body with
// the provided values map. Keys are "1", "2", etc. This is the main
// personalization hook for campaigns.
func RenderBody(body string, vars map[string]string) string {
	result := body
	for k, v := range vars {
		result = strings.ReplaceAll(result, "{{"+k+"}}", v)
	}
	return result
}

// ContactVars builds the standard variable map from a contact's fields.
// {{1}} = name, {{2}} = phone, {{3}} = email.
func ContactVars(name, phone, email string) map[string]string {
	return map[string]string{
		"1": name,
		"2": phone,
		"3": email,
	}
}
