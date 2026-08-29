package flows

import (
	"encoding/json"
	"errors"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("flow not found")
)

type CreateInput struct {
	Name            string   `json:"name" binding:"required"`
	Description     string   `json:"description"`
	TriggerType     string   `json:"trigger_type"`
	TriggerKeywords []string `json:"trigger_keywords"`
	Steps           []Step   `json:"steps"`
	FirstStepID     string   `json:"first_step_id"`
	Active          bool     `json:"active"`
}

type UpdateInput struct {
	Name            *string   `json:"name"`
	Description     *string   `json:"description"`
	TriggerType     *string   `json:"trigger_type"`
	TriggerKeywords *[]string `json:"trigger_keywords"`
	Steps           *[]Step   `json:"steps"`
	FirstStepID     *string   `json:"first_step_id"`
	Active          *bool     `json:"active"`
}

func List(db *gorm.DB, ownerID string) ([]Flow, error) {
	var out []Flow
	err := db.Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&out).Error
	return out, err
}

func Create(db *gorm.DB, ownerID string, in *CreateInput) (*Flow, error) {
	tt := in.TriggerType
	if tt == "" {
		tt = "keyword"
	}
	kw, _ := json.Marshal(in.TriggerKeywords)
	steps, _ := json.Marshal(in.Steps)
	f := &Flow{
		OwnerID:         ownerID,
		Name:            in.Name,
		Description:     in.Description,
		TriggerType:     tt,
		TriggerKeywords: datatypes.JSON(kw),
		Steps:           datatypes.JSON(steps),
		FirstStepID:     in.FirstStepID,
		Active:          in.Active,
	}
	if err := db.Create(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

func GetByID(db *gorm.DB, ownerID, id string) (*Flow, error) {
	var f Flow
	if err := db.Where("id = ? AND owner_id = ?", id, ownerID).First(&f).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &f, nil
}

func Update(db *gorm.DB, f *Flow, in *UpdateInput) (*Flow, error) {
	if in.Name != nil {
		f.Name = *in.Name
	}
	if in.Description != nil {
		f.Description = *in.Description
	}
	if in.TriggerType != nil {
		f.TriggerType = *in.TriggerType
	}
	if in.TriggerKeywords != nil {
		b, _ := json.Marshal(*in.TriggerKeywords)
		f.TriggerKeywords = datatypes.JSON(b)
	}
	if in.Steps != nil {
		b, _ := json.Marshal(*in.Steps)
		f.Steps = datatypes.JSON(b)
	}
	if in.FirstStepID != nil {
		f.FirstStepID = *in.FirstStepID
	}
	if in.Active != nil {
		f.Active = *in.Active
	}
	if err := db.Save(f).Error; err != nil {
		return nil, err
	}
	return f, nil
}

func Delete(db *gorm.DB, f *Flow) error {
	return db.Delete(f).Error
}

// FindMatchingByKeyword returns the first active keyword-triggered flow
// whose keyword appears in the inbound body.
func FindMatchingByKeyword(db *gorm.DB, ownerID, body string) (*Flow, error) {
	body = strings.ToLower(strings.TrimSpace(body))
	if body == "" {
		return nil, nil
	}

	var flows []Flow
	if err := db.Where("owner_id = ? AND active = ? AND trigger_type = ?",
		ownerID, true, "keyword").Find(&flows).Error; err != nil {
		return nil, err
	}

	for i := range flows {
		var kws []string
		_ = json.Unmarshal(flows[i].TriggerKeywords, &kws)
		for _, k := range kws {
			k = strings.ToLower(strings.TrimSpace(k))
			if k != "" && strings.Contains(body, k) {
				return &flows[i], nil
			}
		}
	}
	return nil, nil
}

// GetSteps decodes the JSON step graph into a lookup map.
func (f *Flow) GetSteps() (map[string]Step, error) {
	var steps []Step
	if err := json.Unmarshal(f.Steps, &steps); err != nil {
		return nil, err
	}
	m := make(map[string]Step, len(steps))
	for _, s := range steps {
		m[s.ID] = s
	}
	return m, nil
}

// BumpRun increments the run counter. Best-effort.
func BumpRun(db *gorm.DB, id string) {
	db.Model(&Flow{}).Where("id = ?", id).UpdateColumn("run_count", gorm.Expr("run_count + 1"))
}

// BumpComplete increments the complete counter. Best-effort.
func BumpComplete(db *gorm.DB, id string) {
	db.Model(&Flow{}).Where("id = ?", id).UpdateColumn("complete_count", gorm.Expr("complete_count + 1"))
}
