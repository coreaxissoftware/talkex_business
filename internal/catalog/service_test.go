package catalog

import (
	"testing"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/database"
)

func mustDB(t *testing.T) *gorm.DB {
	t.Helper()
	if err := database.Connect("sqlite://:memory:", true); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if err := database.DB.AutoMigrate(&Product{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	database.DB.Exec("DELETE FROM catalog_products")
	return database.DB
}

func TestCreateAppliesDefaults(t *testing.T) {
	db := mustDB(t)
	p := &Product{RetailerID: "SKU-1", Name: "Silk saree", Price: 4500}
	if err := Create(db, "owner-1", p); err != nil {
		t.Fatal(err)
	}
	if p.Currency != "INR" {
		t.Errorf("currency=%s want INR", p.Currency)
	}
	if p.Availability != AvailabilityInStock {
		t.Errorf("availability=%s want %s", p.Availability, AvailabilityInStock)
	}
}

func TestCreateRejectsMissingFields(t *testing.T) {
	db := mustDB(t)
	cases := []struct {
		name string
		p    Product
	}{
		{"no retailer id", Product{Name: "x", Price: 1}},
		{"no name", Product{RetailerID: "s", Price: 1}},
		{"zero price", Product{RetailerID: "s", Name: "x", Price: 0}},
		{"negative price", Product{RetailerID: "s", Name: "x", Price: -1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := Create(db, "owner-1", &tc.p); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

func TestCreateDetectsDuplicateSKU(t *testing.T) {
	db := mustDB(t)
	first := &Product{RetailerID: "DUP", Name: "a", Price: 100}
	if err := Create(db, "owner-1", first); err != nil {
		t.Fatal(err)
	}
	dup := &Product{RetailerID: "DUP", Name: "b", Price: 200}
	err := Create(db, "owner-1", dup)
	if err != ErrDuplicateSKU {
		t.Fatalf("want ErrDuplicateSKU, got %v", err)
	}

	// Same SKU under a different owner is fine.
	other := &Product{RetailerID: "DUP", Name: "c", Price: 300}
	if err := Create(db, "owner-2", other); err != nil {
		t.Fatalf("cross-tenant duplicate rejected: %v", err)
	}
}

func TestListScopedToOwner(t *testing.T) {
	db := mustDB(t)
	_ = Create(db, "owner-1", &Product{RetailerID: "A", Name: "a", Price: 1})
	_ = Create(db, "owner-1", &Product{RetailerID: "B", Name: "b", Price: 2, Category: "Sarees"})
	_ = Create(db, "owner-2", &Product{RetailerID: "C", Name: "c", Price: 3})

	got, err := List(db, "owner-1", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf("owner-1 got %d rows, want 2", len(got))
	}

	filtered, err := List(db, "owner-1", "Sarees")
	if err != nil {
		t.Fatal(err)
	}
	if len(filtered) != 1 || filtered[0].RetailerID != "B" {
		t.Errorf("category filter wrong: %+v", filtered)
	}
}
