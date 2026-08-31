package flows

import (
	"testing"

	"github.com/coreaxissoftware/talkex_business/internal/contacts"
	"gorm.io/datatypes"
)

func TestEvaluateSplit(t *testing.T) {
	c := &contacts.Contact{
		LifecycleStage: "active",
		LeadScore:      42,
		Tags:           datatypes.JSON([]byte(`["vip","new"]`)),
		CustomFields:   datatypes.JSON([]byte(`{"city":"Bengaluru"}`)),
	}
	cases := []struct {
		name string
		step Step
		want bool
	}{
		{"stage eq active", Step{ConditionField: "lifecycle_stage", ConditionOp: "eq", ConditionValue: "active"}, true},
		{"stage eq churned", Step{ConditionField: "lifecycle_stage", ConditionOp: "eq", ConditionValue: "churned"}, false},
		{"score gt 40", Step{ConditionField: "lead_score", ConditionOp: "gt", ConditionValue: "40"}, true},
		{"score lt 40", Step{ConditionField: "lead_score", ConditionOp: "lt", ConditionValue: "40"}, false},
		{"score gte 42", Step{ConditionField: "lead_score", ConditionOp: "gte", ConditionValue: "42"}, true},
		{"tag contains vip", Step{ConditionField: "tag", ConditionOp: "contains", ConditionValue: "vip"}, true},
		{"tag contains gold", Step{ConditionField: "tag", ConditionOp: "contains", ConditionValue: "gold"}, false},
		{"custom city eq", Step{ConditionField: "custom.city", ConditionOp: "eq", ConditionValue: "Bengaluru"}, true},
		{"unknown field", Step{ConditionField: "made_up", ConditionOp: "eq", ConditionValue: "x"}, false},
		{"empty step", Step{}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := EvaluateSplit(c, tc.step); got != tc.want {
				t.Fatalf("got=%v want=%v", got, tc.want)
			}
		})
	}
}
