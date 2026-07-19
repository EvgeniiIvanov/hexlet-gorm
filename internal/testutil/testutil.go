package testutil

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/EvgeniiIvanov/hexlet-gorm/internal/models"
)

// SetupTestDB creates an in-memory SQLite database for testing
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Enable foreign keys for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

	// Run migrations
	if err := db.AutoMigrate(&models.Movie{}, &models.Actor{}, &models.Director{}, &models.Review{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	// Create unique index on (movie_id, text) for reviews (SQLite-specific)
	db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS reviews_movie_text_idx ON reviews (movie_id, text)")

	return db
}

// WithTx wraps test in a transaction that is rolled back after test completes
func WithTx(t *testing.T, db *gorm.DB, fn func(tx *gorm.DB)) {
	t.Helper()

	tx := db.Begin()
	if tx.Error != nil {
		t.Fatalf("begin tx: %v", tx.Error)
	}

	// Rollback is guaranteed even with t.FailNow or panic
	defer tx.Rollback()

	fn(tx)
}
