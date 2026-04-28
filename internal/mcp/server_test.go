package mcp

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/knivey/dave-web/internal/db"
	"github.com/knivey/dave-web/internal/models"
	"github.com/knivey/dave-web/internal/testutil"
)

func newTestMCPServer(store db.Store) *MCPServer {
	m := &MCPServer{
		db:      store,
		baseURL: "http://localhost:8080",
	}
	return m
}

func makeCallRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func TestHandleCreate(t *testing.T) {
	store := &testutil.MockStore{
		CreatePasteFunc: func(paste *models.Paste) error {
			assert.NotEmpty(t, paste.Slug)
			assert.Equal(t, "test content", paste.Content)
			assert.Equal(t, "My Title", paste.Title)
			assert.Equal(t, "markdown", paste.Language)
			assert.Len(t, paste.Slug, 8)
			assert.Nil(t, paste.ExpiresAt)
			return nil
		},
	}
	srv := newTestMCPServer(store)

	result, err := srv.handleCreate(context.Background(), makeCallRequest(map[string]any{
		"content": "test content",
		"title":   "My Title",
	}))
	require.NoError(t, err)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Paste created:")
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "http://localhost:8080/p/")
}

func TestHandleCreate_NoTitle(t *testing.T) {
	store := &testutil.MockStore{
		CreatePasteFunc: func(paste *models.Paste) error {
			assert.Empty(t, paste.Title)
			return nil
		},
	}
	srv := newTestMCPServer(store)

	result, err := srv.handleCreate(context.Background(), makeCallRequest(map[string]any{
		"content": "just content",
	}))
	require.NoError(t, err)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Paste created:")
}

func TestHandleCreate_MissingContent(t *testing.T) {
	store := &testutil.MockStore{}
	srv := newTestMCPServer(store)

	_, err := srv.handleCreate(context.Background(), makeCallRequest(map[string]any{}))
	assert.Error(t, err)
}

func TestHandleCreate_DBError(t *testing.T) {
	store := &testutil.MockStore{
		CreatePasteFunc: func(paste *models.Paste) error {
			return fmt.Errorf("db connection lost")
		},
	}
	srv := newTestMCPServer(store)

	_, err := srv.handleCreate(context.Background(), makeCallRequest(map[string]any{
		"content": "test",
	}))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "save:")
}

func TestHandleGet_Found(t *testing.T) {
	store := &testutil.MockStore{
		GetPasteFunc: func(slug string) (*models.Paste, error) {
			assert.Equal(t, "abc123", slug)
			return &models.Paste{Slug: "abc123", Content: "raw content"}, nil
		},
	}
	srv := newTestMCPServer(store)

	result, err := srv.handleGet(context.Background(), makeCallRequest(map[string]any{
		"slug": "abc123",
	}))
	require.NoError(t, err)
	assert.Equal(t, "raw content", result.Content[0].(mcp.TextContent).Text)
}

func TestHandleGet_NotFound(t *testing.T) {
	store := &testutil.MockStore{}
	srv := newTestMCPServer(store)

	_, err := srv.handleGet(context.Background(), makeCallRequest(map[string]any{
		"slug": "nonexistent",
	}))
	assert.Error(t, err)
}

func TestHandleGet_DBError(t *testing.T) {
	store := &testutil.MockStore{
		GetPasteFunc: func(slug string) (*models.Paste, error) {
			return nil, fmt.Errorf("connection refused")
		},
	}
	srv := newTestMCPServer(store)

	_, err := srv.handleGet(context.Background(), makeCallRequest(map[string]any{
		"slug": "abc",
	}))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found:")
}

func TestHandleList_Empty(t *testing.T) {
	store := &testutil.MockStore{
		ListPastesFunc: func(limit int) ([]*models.Paste, error) {
			return nil, nil
		},
	}
	srv := newTestMCPServer(store)

	result, err := srv.handleList(context.Background(), makeCallRequest(map[string]any{}))
	require.NoError(t, err)
	assert.Equal(t, "No pastes found.", result.Content[0].(mcp.TextContent).Text)
}

func TestHandleList_WithItems(t *testing.T) {
	store := &testutil.MockStore{
		ListPastesFunc: func(limit int) ([]*models.Paste, error) {
			return []*models.Paste{
				{Slug: "a1", Title: "First", CreatedAt: time.Now()},
				{Slug: "b2", Title: "", CreatedAt: time.Now()},
			}, nil
		},
	}
	srv := newTestMCPServer(store)

	result, err := srv.handleList(context.Background(), makeCallRequest(map[string]any{}))
	require.NoError(t, err)
	text := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, text, "First")
	assert.Contains(t, text, "Untitled")
	assert.Contains(t, text, "http://localhost:8080/p/a1")
}

func TestHandleList_CustomLimit(t *testing.T) {
	store := &testutil.MockStore{
		ListPastesFunc: func(limit int) ([]*models.Paste, error) {
			assert.Equal(t, 10, limit)
			return nil, nil
		},
	}
	srv := newTestMCPServer(store)

	_, err := srv.handleList(context.Background(), makeCallRequest(map[string]any{
		"limit": float64(10),
	}))
	require.NoError(t, err)
}

func TestHandleList_DBError(t *testing.T) {
	store := &testutil.MockStore{
		ListPastesFunc: func(limit int) ([]*models.Paste, error) {
			return nil, fmt.Errorf("db error")
		},
	}
	srv := newTestMCPServer(store)

	_, err := srv.handleList(context.Background(), makeCallRequest(map[string]any{}))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "list:")
}

func TestHandleDelete_Found(t *testing.T) {
	store := &testutil.MockStore{
		DeletePasteFunc: func(slug string) error {
			assert.Equal(t, "abc123", slug)
			return nil
		},
	}
	srv := newTestMCPServer(store)

	result, err := srv.handleDelete(context.Background(), makeCallRequest(map[string]any{
		"slug": "abc123",
	}))
	require.NoError(t, err)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "abc123 deleted")
}

func TestHandleDelete_NotFound(t *testing.T) {
	store := &testutil.MockStore{
		DeletePasteFunc: func(slug string) error {
			return fmt.Errorf("not found")
		},
	}
	srv := newTestMCPServer(store)

	_, err := srv.handleDelete(context.Background(), makeCallRequest(map[string]any{
		"slug": "nonexistent",
	}))
	assert.Error(t, err)
}

func TestHandleDelete_DBError(t *testing.T) {
	store := &testutil.MockStore{
		DeletePasteFunc: func(slug string) error {
			return fmt.Errorf("connection refused")
		},
	}
	srv := newTestMCPServer(store)

	_, err := srv.handleDelete(context.Background(), makeCallRequest(map[string]any{
		"slug": "abc",
	}))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete:")
}
