package mcp

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/knivey/dave-web/internal/db"
	"github.com/knivey/dave-web/internal/models"
	"github.com/knivey/dave-web/internal/renderer"
	"github.com/knivey/dave-web/internal/util"
)

type MCPServer struct {
	mcpServer *server.MCPServer
	db        db.Store
	baseURL   string
}

func NewMCPServer(database db.Store, baseURL string) *MCPServer {
	s := server.NewMCPServer(
		"dave-web",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	m := &MCPServer{
		mcpServer: s,
		db:        database,
		baseURL:   baseURL,
	}

	s.AddTool(m.createTool(), m.handleCreate)
	s.AddTool(m.getTool(), m.handleGet)
	s.AddTool(m.listTool(), m.handleList)
	s.AddTool(m.deleteTool(), m.handleDelete)

	return m
}

func (m *MCPServer) createTool() mcp.Tool {
	return mcp.NewTool("paste_create",
		mcp.WithDescription("Create a new markdown paste"),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("Markdown content for the paste"),
		),
		mcp.WithString("title",
			mcp.Description("Optional title for the paste"),
		),
	)
}

func (m *MCPServer) getTool() mcp.Tool {
	return mcp.NewTool("paste_get",
		mcp.WithDescription("Get raw paste content by slug"),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("Paste slug"),
		),
	)
}

func (m *MCPServer) listTool() mcp.Tool {
	return mcp.NewTool("paste_list",
		mcp.WithDescription("List recent pastes"),
		mcp.WithNumber("limit",
			mcp.Description("Max number of pastes to return (default 50)"),
		),
	)
}

func (m *MCPServer) deleteTool() mcp.Tool {
	return mcp.NewTool("paste_delete",
		mcp.WithDescription("Delete a paste by slug"),
		mcp.WithString("slug",
			mcp.Required(),
			mcp.Description("Paste slug to delete"),
		),
	)
}

func (m *MCPServer) handleCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	content, err := req.RequireString("content")
	if err != nil {
		return nil, err
	}

	title := req.GetString("title", "")

	rendered, err := renderer.Render(content)
	if err != nil {
		return nil, fmt.Errorf("render: %w", err)
	}

	var slug string
	var paste *models.Paste
	for attempt := 0; attempt < 3; attempt++ {
		slug, err = util.GenerateSlug()
		if err != nil {
			return nil, fmt.Errorf("generate slug: %w", err)
		}
		paste = &models.Paste{
			Slug:      slug,
			Title:     title,
			Content:   content,
			Rendered:  rendered,
			CreatedAt: time.Now(),
			Language:  "markdown",
		}
		err = m.db.CreatePaste(paste)
		if !db.IsDuplicateSlug(err) {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("save: %w", err)
	}

	url := fmt.Sprintf("%s/p/%s", m.baseURL, slug)
	return mcp.NewToolResultText(fmt.Sprintf("Paste created: %s", url)), nil
}

func (m *MCPServer) handleGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	slug, err := req.RequireString("slug")
	if err != nil {
		return nil, err
	}

	paste, err := m.db.GetPaste(slug)
	if err != nil {
		return nil, fmt.Errorf("not found: %w", err)
	}

	return mcp.NewToolResultText(paste.Content), nil
}

func (m *MCPServer) handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := req.GetInt("limit", 50)

	pastes, err := m.db.ListPastes(limit)
	if err != nil {
		return nil, fmt.Errorf("list: %w", err)
	}

	result := ""
	for _, p := range pastes {
		title := p.Title
		if title == "" {
			title = "Untitled"
		}
		result += fmt.Sprintf("%s  %s  %s/p/%s\n", p.CreatedAt.Format("2006-01-02 15:04"), title, m.baseURL, p.Slug)
	}

	if result == "" {
		result = "No pastes found."
	}

	return mcp.NewToolResultText(result), nil
}

func (m *MCPServer) handleDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	slug, err := req.RequireString("slug")
	if err != nil {
		return nil, err
	}

	if err := m.db.DeletePaste(slug); err != nil {
		return nil, fmt.Errorf("delete: %w", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Paste %s deleted", slug)), nil
}

func (m *MCPServer) Run(addr string) error {
	sseServer := server.NewSSEServer(m.mcpServer, server.WithBaseURL(fmt.Sprintf("http://localhost%s", addr)))

	mux := http.NewServeMux()
	mux.Handle("/sse", sseServer.SSEHandler())
	mux.Handle("/message", sseServer.MessageHandler())

	var handler http.Handler = mux
	handler = m.requireAPIKey(handler)

	log.Printf("MCP server listening on %s", addr)
	return http.ListenAndServe(addr, handler)
}

func (m *MCPServer) requireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/sse" {
			next.ServeHTTP(w, r)
			return
		}
		key := r.Header.Get("X-API-Key")
		if key == "" {
			http.Error(w, "missing X-API-Key header", http.StatusUnauthorized)
			return
		}
		_, err := m.db.GetAPIKey(key)
		if err != nil {
			http.Error(w, "invalid API key", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
