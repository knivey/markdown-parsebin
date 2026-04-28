# MCP Integration

dave-web includes an MCP (Model Context Protocol) server that exposes paste management as tools for AI agents. The server communicates over HTTP/SSE and is started automatically with the `serve` command.

## Starting the MCP server

```bash
dave-web serve --addr :8080 --mcp-addr :8081 --db dave-web.db
```

The MCP server runs on the port specified by `--mcp-addr` (default `:8081`) in the same process as the web server. Both share the same database.

## Connecting a client

The MCP server exposes two HTTP endpoints:

| Path | Purpose |
|------|---------|
| `/sse` | SSE stream for server-to-client messages |
| `/message` | HTTP POST for client-to-server messages |

Configure your MCP client with the SSE URL:

```
http://localhost:8081/sse
```

### Example client configuration

For clients that use a JSON config file:

```json
{
  "mcpServers": {
    "dave-web": {
      "url": "http://localhost:8081/sse"
    }
  }
}
```

## Tools

### `paste_create`

Create a new markdown paste.

**Parameters**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `content` | string | Yes | Markdown content |
| `title` | string | No | Paste title |

**Returns**

```
Paste created: http://localhost:8080/p/aB3xK9Qm
```

**Example invocation**

```json
{
  "content": "# My Snippet\n\n```python\nprint('hello')\n```",
  "title": "Python Hello"
}
```

**Note:** The MCP `paste_create` tool does not support TTL. To create pastes with an expiry time, use the [REST API](api.md) with the `ttl` field.

### `paste_get`

Retrieve the raw markdown content of a paste by its slug.

**Parameters**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `slug` | string | Yes | Paste slug |

**Returns**

The raw markdown content as a text result.

**Example invocation**

```json
{ "slug": "aB3xK9Qm" }
```

### `paste_list`

List recent pastes.

**Parameters**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `limit` | number | No | Maximum number of pastes to return (default: 50) |

**Returns**

One line per paste in the format `YYYY-MM-DD HH:MM  <title>  <url>`. Untitled pastes show as "Untitled". Returns "No pastes found." if there are no pastes.

**Example output**

```
2026-04-28 12:00  Python Hello  http://localhost:8080/p/aB3xK9Qm
2026-04-28 11:30  Untitled      http://localhost:8080/p/xY7mN2pQ
```

### `paste_delete`

Delete a paste by its slug.

**Parameters**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `slug` | string | Yes | Paste slug to delete |

**Returns**

```
Paste aB3xK9Qm deleted
```

## Authentication

The MCP message endpoint (`/message`) requires an API key in the `X-API-Key` header. The SSE endpoint (`/sse`) does not require authentication (it is read-only event stream). API keys are the same ones used by the REST API and managed via the `keys` CLI subcommands.

MCP clients must be configured to send the `X-API-Key` header with each request. Not all MCP clients support custom headers — check your client's documentation.

## Notes

- The `--base-url` flag controls the URLs returned by `paste_create` and `paste_list`. Set it to your public URL when running behind a reverse proxy.
- Paste creation via MCP uses the same renderer and slug generation as the REST API.
- Expired pastes are automatically excluded from `paste_get` and `paste_list` results.
