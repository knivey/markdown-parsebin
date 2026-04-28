package testutil

import (
	"embed"
	"testing"

	"github.com/knivey/dave-web/internal/db"
)

var (
	//go:embed test_migrations/*.sql
	migrationsFS embed.FS
)

func NewTestDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.New(":memory:", migrationsFS)
	if err != nil {
		t.Fatalf("failed to create test db: %v", err)
	}
	t.Cleanup(func() { database.Close() })
	return database
}
