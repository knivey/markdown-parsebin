package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func executeCommand(args ...string) (string, error) {
	cmd := buildRootCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = old
		w.Close()
	}()

	cmd.SetArgs(args)
	err := cmd.Execute()
	w.Close()
	buf.ReadFrom(r)
	return buf.String(), err
}

func newTempDB(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp("", "dave-web-test-*.db")
	require.NoError(t, err)
	path := f.Name()
	f.Close()
	t.Cleanup(func() { os.Remove(path) })
	return path
}

func TestFlagDefaults(t *testing.T) {
	cmd := buildRootCmd()

	dbVal, _ := cmd.PersistentFlags().GetString("db")
	assert.Equal(t, "dave-web.db", dbVal)

	bu, _ := cmd.PersistentFlags().GetString("base-url")
	assert.Equal(t, "http://localhost:8080", bu)

	for _, sub := range cmd.Commands() {
		if sub.Use == "serve" {
			addr, _ := sub.Flags().GetString("addr")
			assert.Equal(t, ":8080", addr)
			mcpAddr, _ := sub.Flags().GetString("mcp-addr")
			assert.Equal(t, ":8081", mcpAddr)
		}
	}
}

func TestKeysCreate_Success(t *testing.T) {
	tmpDB := newTempDB(t)

	output, err := executeCommand("keys", "create", "--db", tmpDB)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(strings.TrimSpace(output), "dave_"))
	assert.Len(t, strings.TrimSpace(output), 5+64)
}

func TestKeysCreate_WithDescription(t *testing.T) {
	tmpDB := newTempDB(t)

	output, err := executeCommand("keys", "create", "--db", tmpDB, "--description", "CI pipeline")
	require.NoError(t, err)
	key := strings.TrimSpace(output)
	assert.True(t, strings.HasPrefix(key, "dave_"))

	output, err = executeCommand("keys", "list", "--db", tmpDB)
	require.NoError(t, err)
	assert.Contains(t, output, "CI pipeline")
}

func TestKeysList_Empty(t *testing.T) {
	tmpDB := newTempDB(t)

	output, err := executeCommand("keys", "list", "--db", tmpDB)
	require.NoError(t, err)
	assert.Contains(t, output, "No API keys found.")
}

func TestKeysList_WithKeys(t *testing.T) {
	tmpDB := newTempDB(t)

	_, err := executeCommand("keys", "create", "--db", tmpDB, "--description", "key one")
	require.NoError(t, err)

	_, err = executeCommand("keys", "create", "--db", tmpDB, "--description", "key two")
	require.NoError(t, err)

	output, err := executeCommand("keys", "list", "--db", tmpDB)
	require.NoError(t, err)
	assert.Contains(t, output, "KEY")
	assert.Contains(t, output, "DESCRIPTION")
	assert.Contains(t, output, "key one")
	assert.Contains(t, output, "key two")
}

func TestKeysRevoke_Success(t *testing.T) {
	tmpDB := newTempDB(t)

	output, err := executeCommand("keys", "create", "--db", tmpDB)
	require.NoError(t, err)
	key := strings.TrimSpace(output)

	output, err = executeCommand("keys", "revoke", "--db", tmpDB, "--key", key)
	require.NoError(t, err)
	assert.Contains(t, output, "revoked")
	assert.Contains(t, output, key[:12])

	output, err = executeCommand("keys", "list", "--db", tmpDB)
	require.NoError(t, err)
	assert.Contains(t, output, "No API keys found.")
}

func TestKeysRevoke_Nonexistent(t *testing.T) {
	tmpDB := newTempDB(t)

	_, err := executeCommand("keys", "revoke", "--db", tmpDB, "--key", "dave_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestKeysRevoke_MissingFlag(t *testing.T) {
	tmpDB := newTempDB(t)

	_, err := executeCommand("keys", "revoke", "--db", tmpDB)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required")
}
