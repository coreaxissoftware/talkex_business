package deals

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrInvalidStage   = errors.New("stage not part of pipeline")
	ErrPipelineExists = errors.New("default pipeline already exists")
)

// defaultStages is the seed set used when no pipeline exists yet — the
// simplest kanban that still tells a real sales story.
var defaultStages = []string{"New Lead", "Qualified", "Proposal", "Negotiation", "Won", "Lost"}

// EnsureDefaultPipeline seeds a pipeline for a tenant that has none.
// Idempotent: returns the existing default when one is present.
func EnsureDefaultPipeline(db *gorm.DB, ownerID string) (*Pipeline, error) {
	var p Pipeline
	err := db.Where("owner_id = ? AND is_default = ?", ownerID, true).First(&p).Error
	if err == nil {
		return &p, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	stages, _ := json.Marshal(defaultStages)
	p = Pipeline{
		OwnerID:   ownerID,
		Name:      "Sales Pipeline",
		Stages:    stages,
		IsDefault: true,
	}
	if err := db.Create(&p).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

// ListPipelines returns every pipeline for the tenant.
func ListPipelines(db *gorm.DB, ownerID string) ([]Pipeline, error) {
	var out []Pipeline
	err := db.Where("owner_id = ?", ownerID).Order("is_default DESC, created_at ASC").Find(&out).Error
	return out, err
}

// Kanban returns deals grouped by stage for the given pipeline. The
// order of the outer slice matches Pipeline.Stages exactly so the
// dashboard can render columns left-to-right without re-sorting.
type Column struct {
	Stage string  `json:"stage"`
	Deals []*Deal `json:"deals"`
	Total float64 `json:"total_value"`
}

func Kanban(db *gorm.DB, ownerID, pipelineID string) ([]Column, error) {
	var p Pipeline
	err := db.Where("id = ? AND owner_id = ?", pipelineID, ownerID).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var stages []string
	_ = json.Unmarshal(p.Stages, &stages)

	var all []Deal
	if err := db.Where("owner_id = ? AND pipeline_id = ?", ownerID, pipelineID).
		Order("stage_changed_at DESC").Find(&all).Error; err != nil {
		return nil, err
	}
	// Bucket by stage; unknown stages fall into the first column (they
	// exist only when a stage is deleted from the pipeline after use).
	buckets := make(map[string][]*Deal, len(stages))
	totals := make(map[string]float64, len(stages))
	for i := range all {
		d := &all[i]
		buckets[d.Stage] = append(buckets[d.Stage], d)
		totals[d.Stage] += d.Value
	}
	cols := make([]Column, 0, len(stages))
	for _, s := range stages {
		cols = append(cols, Column{Stage: s, Deals: buckets[s], Total: totals[s]})
	}
	return cols, nil
}

// CreateDeal persists a deal. Stage must exist in the pipeline.
func CreateDeal(db *gorm.DB, ownerID string, d *Deal) error {
	if d.Title == "" || d.PipelineID == "" || d.Stage == "" {
		return errors.New("title, pipeline_id, and stage are required")
	}
	d.OwnerID = ownerID
	d.StageChangedAt = time.Now()
	if d.Currency == "" {
		d.Currency = "INR"
	}
	// Validate stage.
	var p Pipeline
	if err := db.Where("id = ? AND owner_id = ?", d.PipelineID, ownerID).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	var stages []string
	_ = json.Unmarshal(p.Stages, &stages)
	if !contains(stages, d.Stage) {
		return ErrInvalidStage
	}
	return db.Create(d).Error
}

// MoveDeal changes stage. Records the transition timestamp so per-stage
// dwell-time analytics stay accurate.
func MoveDeal(db *gorm.DB, ownerID, dealID, newStage string) (*Deal, error) {
	var d Deal
	if err := db.Where("id = ? AND owner_id = ?", dealID, ownerID).First(&d).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	var p Pipeline
	if err := db.Where("id = ? AND owner_id = ?", d.PipelineID, ownerID).First(&p).Error; err != nil {
		return nil, err
	}
	var stages []string
	_ = json.Unmarshal(p.Stages, &stages)
	if !contains(stages, newStage) {
		return nil, ErrInvalidStage
	}
	d.Stage = newStage
	d.StageChangedAt = time.Now()
	// Terminal stages record ClosedAt.
	if newStage == "Won" || newStage == "Lost" {
		now := time.Now()
		d.ClosedAt = &now
	} else {
		d.ClosedAt = nil
	}
	return &d, db.Save(&d).Error
}

func contains(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}
