# MCP Tool Map V2 (US Equities Only)

Source of truth: `research-warehouse` MCP server v2.

## Scope
- Market scope is fixed: US equities + US ETFs only.
- Non-US ticker or internal material code (e.g. `MAT_*`) should be treated as invalid scope.

## Response Contract
Every tool returns:
- `ok`
- `data`
- `meta` (`query_id`, `as_of`, `source`, `latency_ms`)
- `warnings`
- `error` (`code`, `message`, `retryable`, `suggestion`, `details`)

Standard error codes:
- `INVALID_INPUT`
- `INVALID_TICKER`
- `INVALID_MARKET_SCOPE`
- `NOT_FOUND`
- `UPSTREAM_ERROR`
- `INTERNAL_ERROR`

## Tools

### `read_company_overview`
- Required: `ticker`
- Optional: `include_relations`, `include_recent_signals`
- Use for: company fundamentals + supply chain position + recent signals

### `read_market_snapshot`
- Required: `tickers[]`
- Optional: `fields[]` (`price|technicals|options_flow`), `as_of`
- Use for: latest cross-ticker snapshot

### `read_price_history`
- Required: `ticker`
- Optional: `start`, `end`, `limit`, `order`
- Use for: historical bar pull

### `read_signals`
- Required: none
- Optional: `entity`, `signal_type`, `severity_min`, `since`, `cursor`, `order`, `status`, `limit`
- Use for: signal feed retrieval

### `read_reports`
- Required: none
- Optional: `ticker`, `topic`, `since`, `order`, `sources[]`, `limit`, `cursor`
- Use for: research report / expert feed retrieval

### `read_supply_chain_graph`
- Required: `ticker`
- Optional: `direction`, `max_depth`
- Use for: graph nodes/edges around ticker

### `list_pipelines`
- Required: none
- Optional: `name_like`, `limit`
- Use for: Prefect deployment discovery

### `run_pipeline`
- Required: `deployment_name`
- Optional: `params`, `idempotency_key`, `dry_run`
- Use for: pipeline trigger with idempotency and dry-run guard

### `track_run`
- Required: `run_id`
- Optional: `include_logs`, `log_lines`
- Use for: run status + compressed log summary

### `health_check`
- Required: none
- Optional: `checks[]` (`db|prefect_server|prefect_worker|market_api`)
- Use for: service readiness and dependency state
