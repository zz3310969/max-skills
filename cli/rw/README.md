# rw CLI

Go CLI wrapper for Research Warehouse MCP tools over stdio.

## Build

```bash
cd cli/rw
go build -o rw .
```

## Configure MCP Server

```bash
export RW_MCP_SERVER_CMD="python /Users/zhengliangtian/Max/research-warehouse/mcp_server/server.py"
```

If your project uses a managed environment, point to that Python executable instead.

## Commands

```bash
./rw tools
./rw company --ticker NVDA
./rw chain --entity NVDA --direction both --max-depth 3
./rw macro --days 30
./rw semantic --query "CoWoS capacity expansion 2026" --limit 5
./rw call --tool query_company --args '{"ticker":"NVDA"}'
```

Output follows MCP response contract (`data/as_of/quality/source/errors`), and retries once on `INTERNAL_ERROR` by default.

