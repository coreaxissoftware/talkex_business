package automation

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/gorm"
)

var ErrRuleNotFound = errors.New("automation rule not found")

const (
	MatchContains   = "contains"
	MatchExact      = "exact"
	MatchStartsWith = "starts_with"
)

func List(db *gorm.DB, ownerID string) ([]Rule, error) {
	var out []Rule
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}

func GetByID(db *gorm.DB, ownerID, id string) (*Rule, error) {
	var r Rule
	err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&r).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrRuleNotFound
	}
	return &r, err
}

type CreateInput struct {
	Name            string   `json:"name" binding:"required"`
	TriggerKeywords []string `json:"trigger_keywords" binding:"required"`
	MatchType       string   `json:"match_type"`
	ResponseBody    string   `json:"response_body" binding:"required"`
	TemplateID      *string  `json:"template_id"`
	Active          bool     `json:"active"`
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*Rule, error) {
	match := in.MatchType
	if match == "" {
		match = MatchContains
	}
	kwJSON, _ := json.Marshal(in.TriggerKeywords)
	r := &Rule{
		OwnerID:         ownerID,
		Name:            in.Name,
		TriggerKeywords: kwJSON,
		MatchType:       match,
		ResponseBody:    in.ResponseBody,
		TemplateID:      in.TemplateID,
		Active:          in.Active,
	}
	if err := db.Create(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

type UpdateInput struct {
	Name            *string   `json:"name"`
	TriggerKeywords *[]string `json:"trigger_keywords"`
	MatchType       *string   `json:"match_type"`
	ResponseBody    *string   `json:"response_body"`
	TemplateID      *string   `json:"template_id"`
	Active          *bool     `json:"active"`
}

func Update(db *gorm.DB, r *Rule, in *UpdateInput) (*Rule, error) {
	if in.Name != nil {
		r.Name = *in.Name
	}
	if in.TriggerKeywords != nil {
		kwJSON, _ := json.Marshal(*in.TriggerKeywords)
		r.TriggerKeywords = kwJSON
	}
	if in.MatchType != nil {
		r.MatchType = *in.MatchType
	}
	if in.ResponseBody != nil {
		r.ResponseBody = *in.ResponseBody
	}
	if in.TemplateID != nil {
		r.TemplateID = in.TemplateID
	}
	if in.Active != nil {
		r.Active = *in.Active
	}
	return r, db.Save(r).Error
}

func Delete(db *gorm.DB, r *Rule) error {
	return db.Delete(r).Error
}

// FindMatching returns the first active rule whose keyword matches the
// given inbound body under the rule's MatchType. Callers use this from
// their inbound-message pipeline; a nil return means "no auto-reply".
func FindMatching(db *gorm.DB, ownerID, body string) (*Rule, error) {
	var rules []Rule
	if err := db.Where("owner_id = ? AND active = ?", ownerID, true).
		Order("created_at ASC").Find(&rules).Error; err != nil {
		return nil, err
	}

	needle := strings.ToLower(strings.TrimSpace(body))
	for i := range rules {
		var kws []string
		if err := json.Unmarshal(rules[i].TriggerKeywords, &kws); err != nil {
			continue
		}
		for _, kw := range kws {
			kw = strings.ToLower(strings.TrimSpace(kw))
			if kw == "" {
				continue
			}
			match := false
			switch rules[i].MatchType {
			case MatchExact:
				match = needle == kw
			case MatchStartsWith:
				match = strings.HasPrefix(needle, kw)
			default: // contains
				match = strings.Contains(needle, kw)
			}
			if match {
				return &rules[i], nil
			}
		}
	}
	return nil, nil
}

// BumpFireCount records that a rule was triggered — non-fatal on error,
// the reply already went out.
func BumpFireCount(db *gorm.DB, r *Rule) {
	_ = db.Model(r).Update("fire_count", gorm.Expr("fire_count + 1")).Error
}
