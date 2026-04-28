# CLI Reference

## Global flags

These flags apply to all subcommands.

| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `dave-web.db` | Path to the SQLite database file |
| `--base-url` | `http://localhost:8080` | Base URL used for paste links in MCP and API responses |

## `serve`

Start the web server and MCP server in a single process.

```bash
dave-web serve [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--addr` | `:8080` | Listen address for the web server |
| `--mcp-addr` | `:8081` | Listen address for the MCP server |

The web server serves the paste list, rendered views, raw markdown, and the JSON API. A background TTL cleaner runs on a 5-minute interval to delete expired pastes. The MCP server exposes SSE at `/sse` and a message endpoint at `/message`.

### Example

```bash
dave-web serve --addr :3000 --mcp-addr :3001 --db /data/pastes.db --base-url https://paste.example.com
```

## `keys`

Manage API keys for the `POST /api/pastes` endpoint.

### `keys create`

Create a new API key and print it to stdout.

```bash
dave-web keys create [--description "description"]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--description` | `""` | Human-readable description for the key |

### Example

```bash
$ dave-web keys create --description "CI pipeline"
dave_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
```

The full key is printed once and cannot be retrieved again. Store it securely.

### `keys list`

List all API keys.

```bash
dave-web keys list
```

Output is a table showing the first 12 characters of each key (truncated), description, and creation date.

### Example

```
$ dave-web keys list
KEY             DESCRIPTION     CREATED
dave_a1b2c3...  CI pipeline     2026-04-28 12:00
dave_f7e8d9...  Personal        2026-04-27 09:30
```

### `keys revoke`

Revoke (delete) an API key.

```bash
dave-web keys revoke --key <full-api-key>
```

| Flag | Default | Description |
|------|---------|-------------|
| `--key` | (required) | The full API key to revoke |

### Example

```
$ dave-web keys revoke --key dave_a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2
Key dave_a1b2c3... revoked
```
