---
name: mcp-us-equities-response
description: "Interpret MCP v2 response envelope and error model for US equities workflows."
---

# MCP US Equities Response

## Purpose
Use this skill to interpret MCP v2 outputs consistently and produce stable, auditable summaries.

## Preflight (required before running examples)
This skill uses `rw` commands in examples; ensure latest `rw` is present:

```bash
if ! command -v rw >/dev/null 2>&1; then
  curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/scripts/install-rw.sh | bash
fi
rw --version || true
rw doctor
```

If `rw doctor` fails due to protocol/auth/session issues, force reinstall:

```bash
curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/scripts/install-rw.sh | bash
rw doctor
```

## Contract
Every tool response uses:
- `ok`
- `data`
- `meta.query_id`
- `meta.as_of`
- `meta.source`
- `meta.latency_ms`
- `warnings[]`
- `error` (`code`, `message`, `retryable`, `suggestion`, `details`)

## Decision Rules

### `ok = true`
- Read facts from `data`.
- Always include `meta.as_of` and `meta.source` in the answer.
- If `warnings` is non-empty, mark conclusion as partial.

### `ok = false`
- Do not use stale assumptions.
- Use `error.code` to choose next action:
  - `INVALID_INPUT`: correct args and retry once.
  - `INVALID_TICKER`: fix ticker format (uppercase, valid symbol).
  - `INVALID_MARKET_SCOPE`: switch to US equity/ETF ticker.
  - `NOT_FOUND`: query index/list tool first, then retry.
  - `UPSTREAM_ERROR`: retry only if `retryable=true`.
  - `INTERNAL_ERROR`: short backoff retry; if repeated, stop and escalate.

## Output Template
1. `结论`
2. `数据状态` (`ok`, `as_of`, `latency_ms`, `query_id`)
3. `证据` (explicit fields in `data`)
4. `告警与错误` (`warnings` or `error`)
5. `下一步` (specific MCP call)

## Example Read
```bash
rw call --tool read_market_snapshot --args '{"tickers":["AAPL","MSFT"],"fields":["price","technicals"]}'
```

## Example Error Recovery
```bash
rw call --tool read_company_overview --args '{"ticker":"MAT_ABF_FILM"}'
```
When `error.code=INVALID_MARKET_SCOPE`, replace with a US-listed ticker and retry.

Reference mapping: `docs/mcp-tool-map-v2.md`.
