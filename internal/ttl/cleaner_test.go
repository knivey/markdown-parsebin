package ttl

import (
	"embed"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knivey/dave-web/internal/db"
	"github.com/knivey/dave-web/internal/models"
)

//go:embed all:migrations
var testMigrationsFS embed.FS

func newTestDB(t *testing.T) *db.DB {
	t.Helper()
	database, err := db.New(":memory:", testMigrationsFS)
	require.NoError(t, err)
	t.Cleanup(func() { database.Close() })
	return database
}

func TestCleaner_DeletesExpired(t *testing.T) {
	database := newTestDB(t)

	past := time.Now().Add(-1 * time.Hour)
	paste := &models.Paste{
		Slug:      "expired",
		Content:   "gone",
		CreatedAt: time.Now(),
		ExpiresAt: &past,
		Language:  "markdown",
	}
	require.NoError(t, database.CreatePaste(paste))

	StartCleaner(database, 100*time.Millisecond)
	time.Sleep(250 * time.Millisecond)

	_, err := database.GetPaste("expired")
	assert.Error(t, err, "expired paste should be deleted")
}

func TestCleaner_KeepsActive(t *testing.T) {
	database := newTestDB(t)

	future := time.Now().Add(1 * time.Hour)
	paste := &models.Paste{
		Slug:      "active",
		Content:   "stay",
		CreatedAt: time.Now(),
		ExpiresAt: &future,
		Language:  "markdown",
	}
	require.NoError(t, database.CreatePaste(paste))

	StartCleaner(database, 100*time.Millisecond)
	time.Sleep(250 * time.Millisecond)

	got, err := database.GetPaste("active")
	assert.NoError(t, err, "active paste should still exist")
	assert.Equal(t, "active", got.Slug)
}
