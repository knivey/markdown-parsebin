# Development

## Prerequisites

- **Go 1.24+**
- **GCC** — required for CGO (the SQLite driver `mattn/go-sqlite3` is a C library)

On Debian/Ubuntu:

```bash
sudo apt install build-essential
```

On macOS, Xcode command line tools provide GCC.

## Building

```bash
go build -o dave-web ./cmd/dave-web/
```

For a smaller binary:

```bash
go build -ldflags="-s -w" -o dave-web ./cmd/dave-web/
```

## Running

```bash
# Start web + MCP servers
./dave-web serve --addr :8080 --mcp-addr :8081

# Create an API key
./dave-web keys create --description "dev"
```

Both servers share the same database file. SQLite WAL mode handles concurrent access.

## Testing

All tests require `CGO_ENABLED=1` for the SQLite driver.

```bash
# Run all tests
CGO_ENABLED=1 go test ./...

# Run with verbose output
CGO_ENABLED=1 go test ./... -v

# Run a specific package
CGO_ENABLED=1 go test ./internal/web/ -v

# Run a specific test
CGO_ENABLED=1 go test ./internal/web/ -v -run TestAPICreate_WithTTL
```

### Test categories

| Package | Type | Count | Description |
|---------|------|-------|-------------|
| `internal/renderer` | Unit | 19 | Markdown rendering (GFM, code blocks, links, etc.) |
| `internal/util` | Unit | 3 | Slug generation (length, charset, uniqueness) |
| `internal/db` | Integration | 24 | SQLite CRUD operations (in-memory database) |
| `internal/web` (server_test) | Unit | 19 | HTTP handlers with mocked database |
| `internal/web` (integration_test) | Integration | 12 | Full HTTP server with real SQLite |
| `internal/mcp` | Unit | 13 | MCP tool handlers with mocked database |
| `internal/ttl` | Integration | 2 | Background expiry cleaner |
| `cmd/dave-web` | Integration | 8 | CLI key management with real SQLite |

**Total: 100 tests**

Unit tests use `testutil.MockStore` to mock the `db.Store` interface. Integration tests use in-memory SQLite (`:memory:`) with migrations embedded via `//go:embed`.

## Linting

```bash
go vet ./...
```

## Architecture

```
cmd/dave-web/
  main.go              CLI entry point (cobra commands)
  migrations/          SQL schema files (embedded)
  templates/           HTML templates (embedded)
  static/              CSS assets (embedded)

internal/
  db/
    db.go              SQLite connection + migration runner
    interface.go       Store, PasteStore, APIKeyStore interfaces
    paste.go           Paste CRUD + DeleteExpired
    apikey.go          API key CRUD (dave_ prefix + hex)
  web/
    server.go          Gin server setup, route registration
    handlers.go        HTML handlers: list, view, raw
    api.go             JSON API: POST /api/pastes (with TTL support)
    middleware.go       X-API-Key authentication middleware
  mcp/
    server.go          MCP server with 4 tools over SSE
  renderer/
    markdown.go        Goldmark + Chroma (Dracula) pipeline
  ttl/
    cleaner.go         Background goroutine for expired pastes
  util/
    slug.go            8-char crypto/rand slug generation
  models/
    paste.go           Paste struct
```

### Data flow

1. A paste arrives via `POST /api/pastes` or the `paste_create` MCP tool
2. The markdown content is rendered to HTML via `renderer.Render()`
3. An 8-character slug is generated via `util.GenerateSlug()`
4. The paste is stored in SQLite with both raw content and pre-rendered HTML
5. The web view (`GET /p/:slug`) serves the stored HTML — no re-rendering
6. The TTL cleaner periodically deletes pastes where `expires_at < now()`

### Store interface

`db.Store` is the interface consumed by `web.Server` and `mcp.MCPServer`:

```go
type Store interface {
    PasteStore    // CreatePaste, GetPaste, ListPastes, DeletePaste, DeleteExpired
    APIKeyStore   // CreateAPIKey, GetAPIKey, ListAPIKeys, DeleteAPIKey
}
```

The concrete `*db.DB` type satisfies this interface. Tests use `testutil.MockStore` with function fields to mock individual methods.

### Database

SQLite with WAL journal mode and 5-second busy timeout. Migrations are embedded in the binary and run automatically on startup. Schema:

- **`pastes`** — `slug` (PK), `title`, `content`, `rendered`, `created_at`, `expires_at` (nullable), `language`
- **`api_keys`** — `key` (PK), `description`, `created_at`

### Slug generation

Paste slugs are 8 characters from `[a-zA-Z0-9]` (62 chars, ~218 trillion keyspace) using `crypto/rand` for unbiased selection.
