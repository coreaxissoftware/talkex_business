// Package database provides the GORM connection and shared model mixins.
//
// DATABASE_URL determines the dialect:
//   - "sqlite://path"         → SQLite (local dev, pure-Go driver, no CGO)
//   - "postgres://..."        → Postgres (production)
//
// AutoMigrate is called on connect so the schema stays in sync during
// development; production deployments should switch to explicit migration
// files (golang-migrate) once the schema stabilises — noted in CONTEXT.md
// Phase 2 gaps.
package database

import (
	"fmt"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Base is embedded in every domain model: UUID primary key + timestamps.
type Base struct {
	ID        string    `gorm:"type:varchar(36);primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BeforeCreate generates a UUID if the ID field is empty.
func (b *Base) BeforeCreate(tx *gorm.DB) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	return nil
}

// Connect opens the DB based on the DATABASE_URL scheme and sets DB.
func Connect(databaseURL string, isDev bool) error {
	var dialector gorm.Dialector

	switch {
	case strings.HasPrefix(databaseURL, "sqlite://"):
		path := strings.TrimPrefix(databaseURL, "sqlite://")
		dialector = sqlite.Open(path)
	case strings.HasPrefix(databaseURL, "postgres://"), strings.HasPrefix(databaseURL, "postgresql://"):
		dialector = postgres.Open(databaseURL)
	default:
		return fmt.Errorf("unsupported DATABASE_URL scheme: %s", databaseURL)
	}

	logLevel := logger.Silent
	if isDev {
		logLevel = logger.Info
	}

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// Connection pool tuning (Postgres only; SQLite is single-connection).
	if !strings.HasPrefix(databaseURL, "sqlite://") {
		sqlDB, _ := DB.DB()
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	return nil
}

// AutoMigrate runs GORM auto-migration for the given models.
func AutoMigrate(models ...interface{}) error {
	return DB.AutoMigrate(models...)
}
