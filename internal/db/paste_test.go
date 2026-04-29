package db

import (
	"embed"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knivey/dave-web/internal/models"
)

//go:embed all:migrations
var migrationsFS embed.FS

func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := New(":memory:", migrationsFS)
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCreatePaste(t *testing.T) {
	db := newTestDB(t)
	now := time.Now()
	paste := &models.Paste{
		Slug:      "test1",
		Title:     "Test",
		Content:   "# Hello",
		Rendered:  "<h1>Hello</h1>",
		CreatedAt: now,
		Language:  "markdown",
	}
	err := db.CreatePaste(paste)
	assert.NoError(t, err)
}

func TestGetPaste_Found(t *testing.T) {
	db := newTestDB(t)
	now := time.Now()
	paste := &models.Paste{
		Slug:      "abc123",
		Title:     "My Paste",
		Content:   "**bold**",
		Rendered:  "<p><strong>bold</strong></p>",
		CreatedAt: now,
		Language:  "markdown",
	}
	require.NoError(t, db.CreatePaste(paste))

	got, err := db.GetPaste("abc123")
	assert.NoError(t, err)
	assert.Equal(t, "abc123", got.Slug)
	assert.Equal(t, "My Paste", got.Title)
	assert.Equal(t, "**bold**", got.Content)
}

func TestGetPaste_NotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.GetPaste("nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestListPastes_DefaultLimit(t *testing.T) {
	db := newTestDB(t)
	for i := 0; i < 60; i++ {
		paste := &models.Paste{
			Slug:      fmt.Sprintf("paste%d", i),
			Content:   "content",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
			Language:  "markdown",
		}
		require.NoError(t, db.CreatePaste(paste))
	}

	pastes, err := db.ListPastes(0)
	assert.NoError(t, err)
	assert.Len(t, pastes, 50)
}

func TestListPastes_CustomLimit(t *testing.T) {
	db := newTestDB(t)
	for i := 0; i < 10; i++ {
		paste := &models.Paste{
			Slug:      fmt.Sprintf("paste%d", i),
			Content:   "content",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
			Language:  "markdown",
		}
		require.NoError(t, db.CreatePaste(paste))
	}

	pastes, err := db.ListPastes(5)
	assert.NoError(t, err)
	assert.Len(t, pastes, 5)
}

func TestListPastes_Ordering(t *testing.T) {
	db := newTestDB(t)
	for i, name := range []string{"old", "mid", "new"} {
		paste := &models.Paste{
			Slug:      name,
			Content:   "content",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Hour),
			Language:  "markdown",
		}
		require.NoError(t, db.CreatePaste(paste))
	}

	pastes, err := db.ListPastes(10)
	assert.NoError(t, err)
	assert.Len(t, pastes, 3)
	assert.Equal(t, "new", pastes[0].Slug, "newest should be first")
	assert.Equal(t, "old", pastes[2].Slug, "oldest should be last")
}

func TestDeletePaste_Found(t *testing.T) {
	db := newTestDB(t)
	paste := &models.Paste{
		Slug:      "todelete",
		Content:   "bye",
		CreatedAt: time.Now(),
		Language:  "markdown",
	}
	require.NoError(t, db.CreatePaste(paste))

	err := db.DeletePaste("todelete")
	assert.NoError(t, err)

	_, err = db.GetPaste("todelete")
	assert.Error(t, err)
}

func TestDeletePaste_NotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.DeletePaste("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDeleteExpired(t *testing.T) {
	db := newTestDB(t)
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	expired := &models.Paste{
		Slug:      "expired1",
		Content:   "gone",
		CreatedAt: time.Now(),
		ExpiresAt: &past,
		Language:  "markdown",
	}
	expired2 := &models.Paste{
		Slug:      "expired2",
		Content:   "gone2",
		CreatedAt: time.Now(),
		ExpiresAt: &past,
		Language:  "markdown",
	}
	active := &models.Paste{
		Slug:      "active",
		Content:   "stay",
		CreatedAt: time.Now(),
		ExpiresAt: &future,
		Language:  "markdown",
	}
	require.NoError(t, db.CreatePaste(expired))
	require.NoError(t, db.CreatePaste(expired2))
	require.NoError(t, db.CreatePaste(active))

	_, err := db.GetPaste("expired1")
	assert.ErrorIs(t, err, ErrNotFound, "expired paste should not be returned by GetPaste")

	_, err = db.GetPaste("active")
	assert.NoError(t, err, "active paste should be returned by GetPaste")

	n, err := db.DeleteExpired()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), n)

	_, err = db.GetPaste("active")
	assert.NoError(t, err, "active paste should still exist")
}

func TestCreatePaste_WithNilExpiresAt(t *testing.T) {
	db := newTestDB(t)
	paste := &models.Paste{
		Slug:      "noexpire",
		Content:   "forever",
		CreatedAt: time.Now(),
		ExpiresAt: nil,
		Language:  "markdown",
	}
	require.NoError(t, db.CreatePaste(paste))

	got, err := db.GetPaste("noexpire")
	assert.NoError(t, err)
	assert.Nil(t, got.ExpiresAt)
}

func TestCreatePaste_DuplicateSlug(t *testing.T) {
	db := newTestDB(t)
	paste := &models.Paste{
		Slug:      "dup1",
		Content:   "first",
		CreatedAt: time.Now(),
		Language:  "markdown",
	}
	require.NoError(t, db.CreatePaste(paste))

	paste2 := &models.Paste{
		Slug:      "dup1",
		Content:   "second",
		CreatedAt: time.Now(),
		Language:  "markdown",
	}
	err := db.CreatePaste(paste2)
	assert.Error(t, err)
}

func TestGetPaste_AllFields(t *testing.T) {
	db := newTestDB(t)
	expiresAt := time.Date(2026, 12, 25, 10, 0, 0, 0, time.UTC)
	now := time.Now()
	paste := &models.Paste{
		Slug:      "fields1",
		Title:     "Full Fields",
		Content:   "**bold**",
		Rendered:  "<p><strong>bold</strong></p>",
		CreatedAt: now,
		ExpiresAt: &expiresAt,
		Language:  "markdown",
	}
	require.NoError(t, db.CreatePaste(paste))

	got, err := db.GetPaste("fields1")
	require.NoError(t, err)
	assert.Equal(t, "fields1", got.Slug)
	assert.Equal(t, "Full Fields", got.Title)
	assert.Equal(t, "**bold**", got.Content)
	assert.Equal(t, "<p><strong>bold</strong></p>", got.Rendered)
	assert.Equal(t, "markdown", got.Language)
	assert.NotNil(t, got.ExpiresAt)
	assert.WithinDuration(t, expiresAt, *got.ExpiresAt, time.Second)
	assert.WithinDuration(t, now, got.CreatedAt, time.Second)
}

func TestListPastes_EmptyDB(t *testing.T) {
	db := newTestDB(t)
	pastes, err := db.ListPastes(10)
	assert.NoError(t, err)
	assert.Empty(t, pastes)
}

func TestListPastes_NegativeLimit(t *testing.T) {
	db := newTestDB(t)
	for i := 0; i < 55; i++ {
		paste := &models.Paste{
			Slug:      fmt.Sprintf("paste%d", i),
			Content:   "content",
			CreatedAt: time.Now().Add(time.Duration(i) * time.Second),
			Language:  "markdown",
		}
		require.NoError(t, db.CreatePaste(paste))
	}

	pastes, err := db.ListPastes(-1)
	assert.NoError(t, err)
	assert.Len(t, pastes, 50)
}

func TestDeleteExpired_NoExpired(t *testing.T) {
	db := newTestDB(t)
	future := time.Now().Add(1 * time.Hour)
	paste := &models.Paste{
		Slug:      "active",
		Content:   "stay",
		CreatedAt: time.Now(),
		ExpiresAt: &future,
		Language:  "markdown",
	}
	require.NoError(t, db.CreatePaste(paste))

	n, err := db.DeleteExpired()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), n)
}

func TestDeleteExpired_NullExpiresAtPreserved(t *testing.T) {
	db := newTestDB(t)
	past := time.Now().Add(-1 * time.Hour)

	expired := &models.Paste{
		Slug:      "expired",
		Content:   "gone",
		CreatedAt: time.Now(),
		ExpiresAt: &past,
		Language:  "markdown",
	}
	noExpiry := &models.Paste{
		Slug:      "forever",
		Content:   "stays",
		CreatedAt: time.Now(),
		ExpiresAt: nil,
		Language:  "markdown",
	}
	require.NoError(t, db.CreatePaste(expired))
	require.NoError(t, db.CreatePaste(noExpiry))

	n, err := db.DeleteExpired()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), n)

	_, err = db.GetPaste("forever")
	assert.NoError(t, err, "paste with NULL expires_at should not be deleted")
}

func TestCountPastes(t *testing.T) {
	db := newTestDB(t)
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	stats, err := db.CountPastes()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), stats.Total)
	assert.Equal(t, int64(0), stats.Active)
	assert.Equal(t, int64(0), stats.Expired)

	noExpiry := &models.Paste{
		Slug: "noexpire", Content: "forever", CreatedAt: time.Now(), Language: "markdown",
	}
	withFutureExpiry := &models.Paste{
		Slug: "future", Content: "stay", CreatedAt: time.Now(), ExpiresAt: &future, Language: "markdown",
	}
	expired := &models.Paste{
		Slug: "expired", Content: "gone", CreatedAt: time.Now(), ExpiresAt: &past, Language: "markdown",
	}
	require.NoError(t, db.CreatePaste(noExpiry))
	require.NoError(t, db.CreatePaste(withFutureExpiry))
	require.NoError(t, db.CreatePaste(expired))

	stats, err = db.CountPastes()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), stats.Total)
	assert.Equal(t, int64(2), stats.Active)
	assert.Equal(t, int64(1), stats.Expired)
}

func TestListAllPastes(t *testing.T) {
	db := newTestDB(t)
	future := time.Now().Add(1 * time.Hour)
	past := time.Now().Add(-1 * time.Hour)

	active := &models.Paste{
		Slug: "active", Content: "stay", CreatedAt: time.Now(), ExpiresAt: &future, Language: "markdown",
	}
	noExpiry := &models.Paste{
		Slug: "forever", Content: "always", CreatedAt: time.Now(), Language: "markdown",
	}
	expired := &models.Paste{
		Slug: "expired", Content: "gone", CreatedAt: time.Now(), ExpiresAt: &past, Language: "markdown",
	}
	require.NoError(t, db.CreatePaste(active))
	require.NoError(t, db.CreatePaste(noExpiry))
	require.NoError(t, db.CreatePaste(expired))

	pastes, err := db.ListAllPastes()
	assert.NoError(t, err)
	assert.Len(t, pastes, 2)

	slugs := map[string]bool{}
	for _, p := range pastes {
		slugs[p.Slug] = true
		assert.NotEmpty(t, p.Content)
	}
	assert.True(t, slugs["active"])
	assert.True(t, slugs["forever"])
	assert.False(t, slugs["expired"])
}

func TestListAllPastes_EmptyDB(t *testing.T) {
	db := newTestDB(t)
	pastes, err := db.ListAllPastes()
	assert.NoError(t, err)
	assert.Empty(t, pastes)
}

func TestUpdatePasteRendered(t *testing.T) {
	db := newTestDB(t)
	paste := &models.Paste{
		Slug: "upd1", Content: "**bold**", Rendered: "old", CreatedAt: time.Now(), Language: "markdown",
	}
	require.NoError(t, db.CreatePaste(paste))

	err := db.UpdatePasteRendered("upd1", "<p><strong>bold</strong></p>")
	assert.NoError(t, err)

	got, err := db.GetPaste("upd1")
	assert.NoError(t, err)
	assert.Equal(t, "<p><strong>bold</strong></p>", got.Rendered)
}

func TestUpdatePasteRendered_NotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.UpdatePasteRendered("nonexistent", "html")
	assert.NoError(t, err)
}
