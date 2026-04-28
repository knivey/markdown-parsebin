package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_InMemory(t *testing.T) {
	db, err := New(":memory:", migrationsFS)
	require.NoError(t, err)
	defer db.Close()
}

func TestNew_MigrationsRun(t *testing.T) {
	db, err := New(":memory:", migrationsFS)
	require.NoError(t, err)
	defer db.Close()

	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='pastes'").Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "pastes", name)

	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='api_keys'").Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "api_keys", name)
}

func TestNew_IdempotentMigrations(t *testing.T) {
	db1, err := New(":memory:", migrationsFS)
	require.NoError(t, err)
	db1.Close()

	db2, err := New(":memory:", migrationsFS)
	require.NoError(t, err)
	db2.Close()
}
