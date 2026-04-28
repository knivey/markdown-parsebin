package db

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateAPIKey_Format(t *testing.T) {
	db := newTestDB(t)
	ak, err := db.CreateAPIKey("test key")
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(ak.Key, "dave_"), "key should start with dave_")
	assert.Equal(t, 5+64, len(ak.Key), "key should be dave_ (5) + 64 hex chars")
}

func TestCreateAPIKey_Stored(t *testing.T) {
	db := newTestDB(t)
	ak, err := db.CreateAPIKey("my description")
	require.NoError(t, err)

	got, err := db.GetAPIKey(ak.Key)
	require.NoError(t, err)
	assert.Equal(t, ak.Key, got.Key)
	assert.Equal(t, "my description", got.Description)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestGetAPIKey_NotFound(t *testing.T) {
	db := newTestDB(t)
	_, err := db.GetAPIKey("nonexistent")
	assert.Error(t, err)
}

func TestListAPIKeys(t *testing.T) {
	db := newTestDB(t)
	_, err := db.CreateAPIKey("first")
	require.NoError(t, err)
	_, err = db.CreateAPIKey("second")
	require.NoError(t, err)
	_, err = db.CreateAPIKey("third")
	require.NoError(t, err)

	keys, err := db.ListAPIKeys()
	require.NoError(t, err)
	assert.Len(t, keys, 3)
}

func TestDeleteAPIKey_Found(t *testing.T) {
	db := newTestDB(t)
	ak, err := db.CreateAPIKey("to delete")
	require.NoError(t, err)

	err = db.DeleteAPIKey(ak.Key)
	assert.NoError(t, err)

	_, err = db.GetAPIKey(ak.Key)
	assert.Error(t, err)
}

func TestDeleteAPIKey_NotFound(t *testing.T) {
	db := newTestDB(t)
	err := db.DeleteAPIKey("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCreateAPIKey_EmptyDescription(t *testing.T) {
	db := newTestDB(t)
	ak, err := db.CreateAPIKey("")
	require.NoError(t, err)

	got, err := db.GetAPIKey(ak.Key)
	require.NoError(t, err)
	assert.Equal(t, "", got.Description)
}

func TestListAPIKeys_Empty(t *testing.T) {
	db := newTestDB(t)
	keys, err := db.ListAPIKeys()
	require.NoError(t, err)
	assert.Empty(t, keys)
}

func TestListAPIKeys_Ordering(t *testing.T) {
	db := newTestDB(t)
	first, err := db.CreateAPIKey("first")
	require.NoError(t, err)
	second, err := db.CreateAPIKey("second")
	require.NoError(t, err)

	keys, err := db.ListAPIKeys()
	require.NoError(t, err)
	assert.Len(t, keys, 2)
	assert.Equal(t, second.Key, keys[0].Key, "most recently created should be first")
	assert.Equal(t, first.Key, keys[1].Key, "oldest should be last")
}
