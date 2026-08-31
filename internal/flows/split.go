package flows

import (
	"encoding/json"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/contacts"
)

// EvaluateSplit runs a journey-builder condition against the given
// contact and returns true when the "yes" branch should be taken. Pure
// enough to unit-test — the DB lookup on the contact happens outside.
//
// Supported fields (ConditionField):
//   tag                 — Contact.Tags contains ConditionValue
//   lifecycle_stage     — Contact.LifecycleStage equals ConditionValue
//   lead_score          — Contact.LeadScore compared against ConditionValue (int)
//   custom.<key>        — Contact.CustomFields[<key>] compared as string
//
// Supported ops (ConditionOp):
//   eq | ne             — exact string match
//   contains            — substring match (case-insensitive)
//   gt | lt | gte | lte — numeric compare (attempts strconv on both sides)
func EvaluateSplit(c *contacts.Contact, step Step) bool {
	if c == nil || step.ConditionField == "" {
		return false
	}

	got := extractField(c, step.ConditionField)
	want := step.ConditionValue

	switch step.ConditionOp {
	case "eq":
		return got == want
	case "ne":
		return got != want
	case "contains":
		return strings.Contains(strings.ToLower(got), strings.ToLower(want))
	case "gt", "lt", "gte", "lte":
		gi, gErr := strconv.ParseFloat(got, 64)
		wi, wErr := strconv.ParseFloat(want, 64)
		if gErr != nil || wErr != nil {
			return false
		}
		switch step.ConditionOp {
		case "gt":
			return gi > wi
		case "lt":
			return gi < wi
		case "gte":
			return gi >= wi
		case "lte":
			return gi <= wi
		}
	}
	return false
}

// extractField reads one field from a contact by the ConditionField
// selector; missing values return "".
func extractField(c *contacts.Contact, field string) string {
	switch field {
	case "lifecycle_stage":
		return c.LifecycleStage
	case "lead_score":
		return strconv.Itoa(c.LeadScore)
	case "tag":
		// Match against any tag in Tags — return the joined string so
		// EvaluateSplit's contains/eq work naturally.
		var tags []string
		_ = json.Unmarshal(c.Tags, &tags)
		return strings.Join(tags, ",")
	default:
		if strings.HasPrefix(field, "custom.") {
			key := strings.TrimPrefix(field, "custom.")
			m := map[string]string{}
			_ = json.Unmarshal(c.CustomFields, &m)
			return m[key]
		}
	}
	return ""
}

// LoadContact fetches the contact scoped to owner so a split can consult
// its live attributes. Small helper kept out of the engine so the engine
// stays independent of the contacts package's private DB usage.
func LoadContact(db *gorm.DB, ownerID, contactID string) (*contacts.Contact, error) {
	var c contacts.Contact
	err := db.Where("id = ? AND owner_id = ?", contactID, ownerID).First(&c).Error
	return &c, err
}
