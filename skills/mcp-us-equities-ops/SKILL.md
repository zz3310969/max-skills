---
name: mcp-us-equities-ops
description: "Run, list, track, and health-check MCP pipelines on VPS."
---

# MCP US Equities Ops

## Purpose
Use this skill for operational tools on the VPS MCP service:
- `list_pipelines`
- `run_pipeline`
- `track_run`
- `health_check`

## Prerequisites
1. Install and configure `rw` to the VPS endpoint:
```bash
rw setup --server-url http://113.44.56.214:18080/mcp/
rw doctor
```
2. Use direct calls only:
```bash
rw call --tool <tool_name> --args '<json>'
```

## Operational Workflow
1. Discover deployment names using `list_pipelines`.
2. Validate trigger payload with `run_pipeline` and `dry_run=true`.
3. Execute with optional `idempotency_key`.
4. Poll status with `track_run`.
5. Pull summarized logs only when needed.

## Commands

### List pipelines
```bash
rw call --tool list_pipelines --args '{"name_like":"barchart","limit":100}'
```

### Dry run pipeline
```bash
rw call --tool run_pipeline --args '{"deployment_name":"Market API Daily Technicals/daily-barchart-technicals","params":{"ticker":"NVDA"},"dry_run":true,"idempotency_key":"dryrun-20260309-nvda"}'
```

### Execute pipeline (idempotent)
```bash
rw call --tool run_pipeline --args '{"deployment_name":"Market API Daily Technicals/daily-barchart-technicals","params":{"ticker":"NVDA"},"dry_run":false,"idempotency_key":"run-20260309-nvda"}'
```

### Track run status
```bash
rw call --tool track_run --args '{"run_id":"<FLOW_RUN_ID>","include_logs":false}'
```

### Track run with compressed logs
```bash
rw call --tool track_run --args '{"run_id":"<FLOW_RUN_ID>","include_logs":true,"log_lines":120}'
```

### Health checks
```bash
rw call --tool health_check --args '{"checks":["db","prefect_server","prefect_worker","market_api"]}'
```

## Safety Rules
- Always start with `dry_run=true` for manual triggers.
- Use stable `idempotency_key` to avoid duplicate runs.
- If `error.code=INVALID_MARKET_SCOPE` in `run_pipeline`, remove non-US tickers from `params` and retry.
- If `error.retryable=true`, use bounded retry/backoff.

Reference mapping: `docs/mcp-tool-map-v2.md`.
