// Package flows implements multi-step chatbot flows — the next step up
// from automation.Rule (single keyword → single reply). A Flow is a
// directed graph of Steps: each Step is a typed action (send message,
// wait, branch on reply, assign agent, add tag, end) with a NextStepID
// pointing to what runs after.
//
// Execution is driven by triggers (keyword match on inbound today; more
// coming) and runs step-by-step against a persisted RunState so a wait
// or branch can span messages.
package flows

import (
	"github.com/coreaxissoftware/talkex_business/internal/database"
	"gorm.io/datatypes"
)

// Flow — a named, versioned chatbot flow.
type Flow struct {
	database.Base
	OwnerID     string         `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Description string         `gorm:"type:text" json:"description"`
	// Trigger config — how this flow starts.
	TriggerType     string         `gorm:"type:varchar(32);not null;default:'keyword'" json:"trigger_type"` // keyword | new_contact | manual
	TriggerKeywords datatypes.JSON `gorm:"type:json;default:'[]'" json:"trigger_keywords"`
	// Steps — the graph. Stored as JSON for flexibility during editing.
	Steps         datatypes.JSON `gorm:"type:json;default:'[]'" json:"steps"`
	FirstStepID   string         `gorm:"type:varchar(36)" json:"first_step_id"`
	Active        bool           `gorm:"not null;default:false" json:"active"`
	RunCount      int64          `gorm:"not null;default:0" json:"run_count"`
	CompleteCount int64          `gorm:"not null;default:0" json:"complete_count"`
}

// Step — one node in the flow graph. Kept as a plain struct so it
// marshals cleanly into the Flow.Steps JSON payload.
//
// Type semantics:
//   send_message   → post OutputText to the contact, then advance
//   send_template  → send the referenced template, then advance
//   wait           → pause until at least WaitMinutes elapse
//   branch         → advance to BranchYesID if inbound matches
//                    BranchKeyword, else BranchNoID (no reply → wait)
//   assign_agent   → assign the conversation to AgentUserID
//   add_tag        → add TagName to the contact's tags
//   end            → terminal — mark run complete
type Step struct {
	ID           string   `json:"id"`
	Type         string   `json:"type"`
	Label        string   `json:"label"`
	OutputText   string   `json:"output_text,omitempty"`
	TemplateID   string   `json:"template_id,omitempty"`
	WaitMinutes  int      `json:"wait_minutes,omitempty"`
	BranchKeyword string  `json:"branch_keyword,omitempty"`
	BranchYesID  string   `json:"branch_yes_id,omitempty"`
	BranchNoID   string   `json:"branch_no_id,omitempty"`
	AgentUserID  string   `json:"agent_user_id,omitempty"`
	TagName      string   `json:"tag_name,omitempty"`
	NextStepID   string   `json:"next_step_id,omitempty"`

	// Journey-builder additions (parity with Wati / DoubleTick / Gupshup):
	//
	//   split          → condition-based branch on a contact attribute.
	//                    Reads Contact.{Tags,LifecycleStage,LeadScore,
	//                    CustomFields[ConditionField]} and compares to
	//                    ConditionValue with ConditionOp. Splits to
	//                    BranchYesID / BranchNoID.
	//   webhook        → POST WebhookURL with { contact, run_state, step }.
	//                    Non-2xx routes to BranchNoID; success → BranchYesID
	//                    or NextStepID.
	ConditionField string `json:"condition_field,omitempty"` // tag | lifecycle_stage | lead_score | custom.<key>
	ConditionOp    string `json:"condition_op,omitempty"`    // eq | ne | contains | gt | lt | gte | lte
	ConditionValue string `json:"condition_value,omitempty"`

	WebhookURL     string `json:"webhook_url,omitempty"`
	WebhookTimeoutS int   `json:"webhook_timeout_s,omitempty"`
}

// RunState — one in-flight execution of a Flow for a specific contact.
type RunState struct {
	database.Base
	OwnerID       string `gorm:"type:varchar(36);index;not null" json:"owner_id"`
	FlowID        string `gorm:"type:varchar(36);index;not null" json:"flow_id"`
	ContactID     string `gorm:"type:varchar(36);index;not null" json:"contact_id"`
	Channel       string `gorm:"type:varchar(32)" json:"channel"`
	CurrentStepID string `gorm:"type:varchar(36)" json:"current_step_id"`
	Status        string `gorm:"type:varchar(20);not null;default:'active'" json:"status"` // active | completed | failed | cancelled
	// WaitingUntil is set by wait/branch steps so a background sweep can
	// advance them without an inbound message.
	WaitingUntil  *int64 `json:"waiting_until,omitempty"`
}
