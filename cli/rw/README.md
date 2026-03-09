# rw CLI

Go CLI wrapper for Research Warehouse MCP tools.

Supports two transports:
- HTTP endpoint (`RW_MCP_SERVER_URL`) for out-of-box usage
- Local stdio command (`RW_MCP_SERVER_CMD`) for development

## Install

Recommended (no manual build):

```bash
curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/scripts/install-rw.sh | bash
```

Or build from source:

```bash
cd cli/rw
go build -o rw .
```

## Configure

HTTP mode (recommended):

```bash
rw setup --server-url https://your-mcp-gateway.example.com --auth-token <MCP_AUTH_TOKEN>
rw doctor
```

Stdio mode:

```bash
rw setup --server-cmd "node /path/to/mcp_server/server.js"
rw doctor
```

Config is stored at:
- `~/.config/rw/config.env`

## Commands

```bash
rw tools
rw company --ticker NVDA
rw chain --entity NVDA --direction both --max-depth 3
rw macro --days 30
rw semantic --query "CoWoS capacity expansion 2026" --limit 5
rw call --tool query_company --args '{"ticker":"NVDA"}'
```

Output follows MCP response contract (`data/as_of/quality/source/errors`) and retries once on `INTERNAL_ERROR` by default.

You can also pass token via env var:

```bash
export RW_MCP_AUTH_TOKEN=<MCP_AUTH_TOKEN>
rw --server-url https://your-mcp-gateway.example.com tools
```
