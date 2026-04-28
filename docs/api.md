# REST API Reference

## Authentication

The `POST /api/pastes` endpoint requires an API key passed in the `X-API-Key` header. All other endpoints are unauthenticated.

```bash
curl -H "X-API-Key: dave_YOUR_KEY_HERE" ...
```

API keys are created and managed via the CLI. See [CLI Reference](cli.md#keys).

## Endpoints

### `POST /api/pastes`

Create a new paste.

**Request**

```
POST /api/pastes
Content-Type: application/json
X-API-Key: dave_...
```

```json
{
  "content": "required — the markdown content",
  "title": "optional — paste title",
  "ttl": 3600
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `content` | string | Yes | Markdown content of the paste |
| `title` | string | No | Title displayed on the paste page |
| `ttl` | integer | No | Time-to-live in seconds. Paste is automatically deleted after this duration. Omit or set to 0 for no expiry. |

**Response — 201 Created**

```json
{
  "slug": "aB3xK9Qm",
  "url": "http://localhost:8080/p/aB3xK9Qm"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `slug` | string | 8-character alphanumeric slug |
| `url` | string | Full URL to the rendered paste |

**Error responses**

| Status | Body | Cause |
|--------|------|-------|
| `400` | `{"error": "content is required"}` | Missing or empty `content` field |
| `401` | `{"error": "missing X-API-Key header"}` | No API key header provided |
| `401` | `{"error": "invalid API key"}` | Key not found in database |
| `500` | `{"error": "failed to render markdown"}` | Renderer error |
| `500` | `{"error": "failed to create paste"}` | Database error |

**Example — no expiry**

```bash
curl -X POST http://localhost:8080/api/pastes \
  -H "X-API-Key: dave_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Example",
    "content": "# Hello\n\nSome **bold** text.\n\n```go\nfmt.Println(\"hi\")\n```"
  }'
```

**Example — with 1 hour TTL**

```bash
curl -X POST http://localhost:8080/api/pastes \
  -H "X-API-Key: dave_a1b2c3..." \
  -H "Content-Type: application/json" \
  -d '{"content": "temporary", "title": "Expires in 1h", "ttl": 3600}'
```

### `GET /`

List recent pastes as an HTML page. Shows up to 50 pastes ordered by creation date (newest first). Each row displays the title (linked to the view page), creation date, and expiration date.

### `GET /p/:slug`

View a rendered paste as an HTML page. The markdown is rendered at creation time and stored as HTML. The page includes the title, creation metadata, rendered content, and a link to the raw view.

Returns `404` if the paste does not exist.

### `GET /p/:slug/raw`

View the raw markdown content of a paste as `text/plain`.

Returns `404` if the paste does not exist.

**Example**

```bash
curl http://localhost:8080/p/aB3xK9Qm/raw
```

### `GET /static/*filepath`

Serves embedded static assets (CSS). No authentication required.
