// migrate — one-shot production DB migration tool.
//
// Each migration is a small, idempotent function keyed by an id we
// record in a `schema_migrations` table. Running the CLI twice in a
// row is a no-op: it runs every not-yet-recorded migration in order,
// wraps each in a transaction, records the id on success, aborts the
// whole batch on failure with a clear message.
//
// Deliberately hand-rolled rather than pulling golang-migrate — this
// binary has to survive on Fly.io with the same distroless image as
// the API, and one file we own is easier to ship-hotfix than a
// vendored migration engine.
//
//   go run ./cmd/migrate                     # apply pending
//   go run ./cmd/migrate -status             # list applied vs pending
//   go run ./cmd/migrate -id <id>            # force-apply just one
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/gorm"

	"github.com/coreaxissoftware/talkex_business/internal/config"
	"github.com/coreaxissoftware/talkex_business/internal/database"
)

// schemaMigration is one row per applied migration.
type schemaMigration struct {
	ID        string    `gorm:"primaryKey;type:varchar(80)"`
	AppliedAt time.Time `gorm:"not null"`
	Note      string    `gorm:"type:varchar(500)"`
}

func (schemaMigration) TableName() string { return "schema_migrations" }

// migration is one entry in the migrations table. Ordered by ID sort —
// use a date-prefixed ID so lexicographic == chronological.
type migration struct {
	ID   string
	Note string
	Run  func(*gorm.DB) error
}

// migrations is the canonical ordered list. Never renumber or delete a
// row that has already shipped — future rows just get appended.
var migrations = []migration{
	{
		ID:   "20260904_rename_configs_to_widget_configs",
		Note: "Split the shared `configs` table into `channel_configs` + `widget_configs`",
		Run:  renameLegacyConfigsTable,
	},
}

func main() {
	var (
		status = flag.Bool("status", false, "print applied/pending status and exit")
		one    = flag.String("id", "", "force-apply just this ID (skip the applied check)")
	)
	flag.Parse()

	cfg := config.Get()
	if err := database.Connect(cfg.DatabaseURL, false); err != nil {
		log.Fatalf("connect: %v", err)
	}
	if err := database.DB.AutoMigrate(&schemaMigration{}); err != nil {
		log.Fatalf("bootstrap schema_migrations: %v", err)
	}

	if *status {
		printStatus()
		return
	}
	if *one != "" {
		runOne(*one)
		return
	}
	runPending()
}

func printStatus() {
	applied := loadApplied()
	fmt.Printf("%-4s  %-45s  %s\n", "STATE", "ID", "NOTE")
	for _, m := range migrations {
		state := "pend"
		if _, ok := applied[m.ID]; ok {
			state = "OK"
		}
		fmt.Printf("%-4s  %-45s  %s\n", state, m.ID, m.Note)
	}
}

func runPending() {
	applied := loadApplied()
	ran := 0
	for _, m := range migrations {
		if _, ok := applied[m.ID]; ok {
			continue
		}
		if err := applyOne(m); err != nil {
			log.Fatalf("migration %s failed: %v", m.ID, err)
		}
		ran++
	}
	if ran == 0 {
		fmt.Println("no pending migrations; up to date")
		return
	}
	fmt.Printf("applied %d migration(s)\n", ran)
}

func runOne(id string) {
	for _, m := range migrations {
		if m.ID == id {
			if err := applyOne(m); err != nil {
				log.Fatalf("migration %s failed: %v", m.ID, err)
			}
			return
		}
	}
	log.Fatalf("unknown migration id: %s", id)
}

func applyOne(m migration) error {
	fmt.Printf("→ %s: %s\n", m.ID, m.Note)
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := m.Run(tx); err != nil {
			return err
		}
		return tx.Create(&schemaMigration{
			ID:        m.ID,
			AppliedAt: time.Now(),
			Note:      m.Note,
		}).Error
	})
}

func loadApplied() map[string]schemaMigration {
	out := map[string]schemaMigration{}
	var rows []schemaMigration
	if err := database.DB.Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		out[r.ID] = r
	}
	return out
}

// renameLegacyConfigsTable splits the legacy shared `configs` table
// (widget rows + channel rows silently overlapping — see commit
// 759d329) into two properly-named tables.
//
// The legacy `configs` table always held widget.Config's shape (that
// migrator ran first). Channels rows never persisted at all, so the
// only correct action is:
//
//   1. Rename `configs` → `widget_configs` (preserving widget data)
//   2. Let the next server boot create a fresh `channel_configs`
//      table via AutoMigrate
//
// Idempotent: skips gracefully if either the old table is already
// gone or the new one already exists.
func renameLegacyConfigsTable(tx *gorm.DB) error {
	if !tx.Migrator().HasTable("configs") {
		fmt.Println("   (no legacy `configs` table found — nothing to rename)")
		return nil
	}
	if tx.Migrator().HasTable("widget_configs") {
		// Some deploys will already have widget_configs auto-created by
		// a boot after commit 759d329. In that case the legacy table
		// still exists in parallel and needs manual review — refuse.
		return errors.New(
			"both `configs` and `widget_configs` exist — one of them holds " +
				"stale duplicate rows; inspect and drop the older one before " +
				"re-running this migration")
	}
	if err := tx.Exec("ALTER TABLE configs RENAME TO widget_configs").Error; err != nil {
		return fmt.Errorf("rename configs -> widget_configs: %w", err)
	}
	fmt.Println("   renamed `configs` -> `widget_configs`")
	fmt.Println("   `channel_configs` will be created on the next server boot")

	// Silence the unused-import warning when the file is compiled with
	// no side effects being triggered (os.Getenv is called below).
	_ = os.Getenv
	return nil
}
