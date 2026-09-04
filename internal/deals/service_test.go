package deals

import (
	"encoding/json"
	"testing"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// mustDB spins up an in-memory SQLite for one test, migrates the deals
// schema, and returns a ready DB handle. Same pattern the rest of the
// backend uses in its unit tests.
func mustDB(t *testing.T) *gorm.DB {
	t.Helper()
	if err := database.Connect("sqlite://:memory:", true); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := database.DB.AutoMigrate(&Pipeline{}, &Deal{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	// Wipe between runs so a shared connection doesn't leak rows.
	database.DB.Exec("DELETE FROM deals")
	database.DB.Exec("DELETE FROM pipelines")
	return database.DB
}

func TestEnsureDefaultPipelineIsIdempotent(t *testing.T) {
	db := mustDB(t)
	p1, err := EnsureDefaultPipeline(db, "owner-A")
	if err != nil {
		t.Fatal(err)
	}
	p2, err := EnsureDefaultPipeline(db, "owner-A")
	if err != nil {
		t.Fatal(err)
	}
	if p1.ID != p2.ID {
		t.Fatalf("second call returned a different pipeline: %s vs %s", p1.ID, p2.ID)
	}

	// Different owner gets their own pipeline.
	p3, err := EnsureDefaultPipeline(db, "owner-B")
	if err != nil {
		t.Fatal(err)
	}
	if p3.ID == p1.ID {
		t.Fatal("owner-B got owner-A's pipeline — cross-tenant leak")
	}
}

func TestEnsureDefaultPipelineHasSeedStages(t *testing.T) {
	db := mustDB(t)
	p, err := EnsureDefaultPipeline(db, "owner-1")
	if err != nil {
		t.Fatal(err)
	}
	var stages []string
	if err := json.Unmarshal(p.Stages, &stages); err != nil {
		t.Fatal(err)
	}
	// The default six stages the frontend kanban relies on.
	want := []string{"New Lead", "Qualified", "Proposal", "Negotiation", "Won", "Lost"}
	if len(stages) != len(want) {
		t.Fatalf("stage count=%d want %d", len(stages), len(want))
	}
	for i, s := range want {
		if stages[i] != s {
			t.Errorf("stage[%d]=%q want %q", i, stages[i], s)
		}
	}
}

func TestCreateDealRejectsInvalidStage(t *testing.T) {
	db := mustDB(t)
	p, _ := EnsureDefaultPipeline(db, "owner-1")
	err := CreateDeal(db, "owner-1", &Deal{
		PipelineID: p.ID,
		Title:      "test",
		Stage:      "NotARealStage",
	})
	if err != ErrInvalidStage {
		t.Fatalf("want ErrInvalidStage, got %v", err)
	}
}

func TestMoveDealStampsClosedOnTerminal(t *testing.T) {
	db := mustDB(t)
	p, _ := EnsureDefaultPipeline(db, "owner-1")
	d := &Deal{PipelineID: p.ID, Title: "big deal", Stage: "Qualified", Value: 500}
	if err := CreateDeal(db, "owner-1", d); err != nil {
		t.Fatal(err)
	}

	moved, err := MoveDeal(db, "owner-1", d.ID, "Won")
	if err != nil {
		t.Fatal(err)
	}
	if moved.Stage != "Won" {
		t.Errorf("stage=%s", moved.Stage)
	}
	if moved.ClosedAt == nil {
		t.Fatal("ClosedAt should be stamped when moving to Won")
	}

	// Moving back off a terminal stage clears ClosedAt.
	moved2, err := MoveDeal(db, "owner-1", d.ID, "Negotiation")
	if err != nil {
		t.Fatal(err)
	}
	if moved2.ClosedAt != nil {
		t.Fatal("ClosedAt should clear when moving away from a terminal stage")
	}
}

func TestKanbanBucketsAndTotals(t *testing.T) {
	db := mustDB(t)
	p, _ := EnsureDefaultPipeline(db, "owner-1")
	_ = CreateDeal(db, "owner-1", &Deal{PipelineID: p.ID, Title: "a", Stage: "New Lead", Value: 100})
	_ = CreateDeal(db, "owner-1", &Deal{PipelineID: p.ID, Title: "b", Stage: "New Lead", Value: 200})
	_ = CreateDeal(db, "owner-1", &Deal{PipelineID: p.ID, Title: "c", Stage: "Qualified", Value: 500})

	cols, err := Kanban(db, "owner-1", p.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cols) != 6 {
		t.Fatalf("cols=%d want 6", len(cols))
	}
	if cols[0].Stage != "New Lead" || len(cols[0].Deals) != 2 || cols[0].Total != 300 {
		t.Errorf("col[0] wrong: stage=%s deals=%d total=%.2f", cols[0].Stage, len(cols[0].Deals), cols[0].Total)
	}
	if cols[1].Stage != "Qualified" || len(cols[1].Deals) != 1 || cols[1].Total != 500 {
		t.Errorf("col[1] wrong: stage=%s deals=%d total=%.2f", cols[1].Stage, len(cols[1].Deals), cols[1].Total)
	}
}
