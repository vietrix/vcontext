# vcontext MCP server (Go)

High-performance, lightweight MCP server that stores long-term context for LLM agents using SQLite + FTS5. Communicates over stdin/stdout via JSON-RPC 2.0.

## Build

```bash
CGO_ENABLED=0 go build -o vcontext ./cmd/vcontext
```

## Install

### Option 1: GitHub release binary (recommended)

- Download from GitHub Releases, then place `vcontext` on your `PATH`.

### Option 2: Install script (Linux/macOS)

```bash
REPO=vietrix/vcontext bash scripts/install.sh
```

### Option 3: Install script (Windows PowerShell)

```powershell
.\scripts\install.ps1 -Repo "vietrix/vcontext"
```

### Option 4: Go install

```bash
GOBIN=$(pwd)/bin go install ./cmd/vcontext
```

## Update

Self-update via GitHub Releases:

```bash
vcontext update
```

Optional override:

- `VCONTEXT_UPDATE_REPO` (defaults to `vietrix/vcontext`)
- `GITHUB_TOKEN` or `GH_TOKEN` for higher API rate limits

## Run

The server reads JSON-RPC requests line-by-line from stdin and writes responses to stdout. The SQLite database path is resolved in this order:

1. `-db` flag
2. `VCONTEXT_DB_PATH` environment variable
3. `./vcontext.db` (default)

## MCP setup

### OpenAI Codex (CLI)

Add the server via CLI:

```bash
codex mcp add vcontext -- /path/to/vcontext -db /path/to/vcontext.db
```

Or in `~/.codex/config.toml`:

```toml
[mcp_servers.vcontext]
command = "/path/to/vcontext"
args = ["-db", "/path/to/vcontext.db"]
```

### Claude Code

Add the server via CLI:

```bash
claude mcp add --transport stdio vcontext -- /path/to/vcontext -db /path/to/vcontext.db
```

## Version

```bash
vcontext version
```

## JSON-RPC methods

Method names:
- `tools/save_context/invoke`
- `tools/search_context/invoke`
- `tools/get_context/invoke`

### save_context

Input:
```json
{
  "source": "string?",
  "thread_id": "string?",
  "role": "string?",
  "title": "string?",
  "content": "string (required)",
  "tags": ["string?"],
  "importance": 3
}
```

Output:
```json
{ "id": "uuid", "created_at": 1234567890 }
```

### search_context

Input:
```json
{
  "query": "string (required)",
  "top_k": 5,
  "thread_id": "string?",
  "min_importance": 1
}
```

Output:
```json
{
  "items": [
    {
      "id": "uuid",
      "title": "string?",
      "source": "string?",
      "thread_id": "string?",
      "created_at": 1234567890,
      "importance": 3,
      "snippet": "preview..."
    }
  ]
}
```

### get_context

Input:
```json
{ "id": "uuid" }
```

Output: full `ContextItem`
```json
{
  "id": "uuid",
  "created_at": 1234567890,
  "source": "string?",
  "thread_id": "string?",
  "role": "string?",
  "title": "string?",
  "content": "full text",
  "tags": ["tag1", "tag2"],
  "importance": 3
}
```

## Example request

Each request must be on a single line (newline-terminated):
```json
{"jsonrpc":"2.0","id":1,"method":"tools/save_context/invoke","params":{"content":"hello world"}}
```
