# dave-web

A self-hosted markdown pastebin with an MCP server for AI tool integration. Pastes are stored in SQLite, rendered to HTML with syntax highlighting, and served via a dark-themed web UI.

## Features

- **Markdown rendering** with GitHub Flavored Markdown, syntax highlighting (Dracula theme via Chroma), auto-linkified URLs, and hard wraps
- **Web interface** with paste list, rendered view, and raw markdown endpoint
- **JSON API** for creating pastes programmatically (`POST /api/pastes`) with optional TTL
- **MCP server** exposing four tools (`paste_create`, `paste_get`, `paste_list`, `paste_delete`) over HTTP/SSE for AI agent integration
- **API key authentication** for the create endpoint, managed via CLI subcommands
- **TTL support** — optional time-to-live on pastes via the API; background cleaner removes expired pastes
- **Single binary** with embedded templates, static assets, and migrations — no external dependencies
- **Docker support** with a multi-stage Alpine build

## Quick Start

### Build

```bash
go build -o dave-web ./cmd/dave-web/
```

Requires Go 1.24+ and GCC (for CGO/SQLite).

### Create an API key

```bash
./dave-web keys create --description "my key"
```

### Start the server

```bash
./dave-web serve --addr :8080 --mcp-addr :8081 --db dave-web.db
```

This starts both the web server and MCP server in a single process. The web server listens on `:8080` and the MCP server on `:8081`.

### Create a paste

```bash
curl -X POST http://localhost:8080/api/pastes \
  -H "X-API-Key: dave_YOUR_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{"content": "# Hello\n\nThis is **markdown**.", "title": "My First Paste"}'
```

Response:

```json
{"slug":"aB3xK9Qm","url":"http://localhost:8080/p/aB3xK9Qm"}
```

Open the URL in a browser to see the rendered paste.

### Create a paste with TTL

```bash
curl -X POST http://localhost:8080/api/pastes \
  -H "X-API-Key: dave_YOUR_KEY_HERE" \
  -H "Content-Type: application/json" \
  -d '{"content": "temporary note", "title": "Expires in 1 hour", "ttl": 3600}'
```

### Connect an MCP client

The MCP server is available at `http://localhost:8081/sse`.

## Docker

```bash
docker build -t dave-web .
docker run -p 8080:8080 -p 8081:8081 -v ./data:/data dave-web \
  serve --db /data/dave-web.db
```

Ports: **8080** (web), **8081** (MCP).

## Documentation

| Document | Description |
|----------|-------------|
| [CLI Reference](docs/cli.md) | All commands, flags, and examples |
| [REST API](docs/api.md) | Endpoint reference with request/response schemas |
| [MCP Integration](docs/mcp.md) | Connecting MCP clients and tool reference |
| [Development](docs/development.md) | Building, testing, and architecture |

## License

[MIT](LICENSE)
