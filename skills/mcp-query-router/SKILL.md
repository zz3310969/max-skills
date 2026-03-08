---
name: mcp-query-router
description: "金融数据查询总路由，负责选工具、组参、调用与补查编排。"
---

# MCP Query Router

## Purpose
Route financial research requests to the right MCP tools, validate parameters, and build a minimal query plan.
This skill follows **query-guide-first mode**: parameter schema must match your MCP server query guide.

This skill focuses on query orchestration:
- What to query
- Which tool to use first
- When to run fallback queries

It does not produce final user-facing conclusions by itself.

## Prerequisites
1. Install `rw` CLI:
```bash
curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/scripts/install-rw.sh | bash
```
2. Configure MCP endpoint:
```bash
rw setup --server-url https://your-mcp-gateway.example.com
rw doctor
```
3. Keep tool/param mapping aligned with:
- your MCP query guide
- `docs/mcp-tool-map.md` in this repository

## Intent Classification
Map user requests to one of these domains:

1. Company and supply chain
2. Macro and risk regime
3. Research and semantic lookup
4. Earnings event and transcript context
5. Flow and positioning (options/ETF)
6. Technicals and trend structure
7. Company discovery and screening

## Tool Routing Matrix

### 1) Company and supply chain
- Primary: `query_company`
- Fallback: `query_supply_chain`, `query_bottleneck`
- Required params: `ticker`
- Optional params: domain-specific filters

### 2) Macro and risk regime
- Primary: `query_macro`
- Fallback: `search_semantic`
- Required params: none
- Optional params: `days`

### 3) Research and semantic lookup
- Primary: `search_semantic`
- Fallback: `get_research_reports`
- Required params: `query`
- Optional params: `limit`, `source_filter`

### 4) Earnings event and transcript context
- Primary: `get_earnings_call`
- Fallback: `query_options_flow`, `query_technicals`
- Required params: `ticker`
- Optional params: `quarter`

### 5) Flow and positioning
- Primary: `query_options_flow` or `query_etf_flow`
- Fallback: the other flow tool, then `query_technicals`
- Required params: `ticker` for single-name flow; optional for radar mode
- Optional params: `days` (options), `weeks` (ETF)

### 6) Technicals and trend structure
- Primary: `query_technicals`
- Fallback: `query_options_flow`, `query_macro`
- Required params: `ticker`
- Optional params: `days`

### 7) Company discovery and screening
- Primary: `search_companies`
- Fallback: `query_company`
- Required params: none
- Optional params: `sector`, `tier`, `min_cap`, `max_cap`, `country`

## Parameter Precheck
Before calling MCP:

1. Ensure all required params are present.
2. Validate enum-like inputs (timeframe, mode, range).
3. Normalize ticker casing (`upper()` convention).
4. If invalid or missing, ask for parameter correction before calling.

Enum guards (strict):
- `query_supply_chain.direction`: `upstream | downstream | both`
- `query_bottleneck.domain`: `memory | photonics | packaging | power | gpu`
- `search_companies.tier`: `chokepoint | enabler | beneficiary`

If precheck fails, do not call MCP; avoid predictable `INVALID_ARGUMENT`.

## Query Execution Strategy
Use this sequence:

1. Run one primary query.
2. Inspect response contract quickly (`quality`, `errors`, `source`).
3. Trigger at most two fallback queries only when needed:
- Needed if `quality` is `partial` and missing evidence blocks confidence.
- Needed if `quality` is `empty` but adjacent data may still answer intent.

## Handoff To Interpreter
Always hand over this normalized payload:

```json
{
  "intent_domain": "flow_and_positioning",
  "queries": [
    {"tool": "query_options_flow", "params": {"ticker": "NVDA"}}
  ],
  "responses": [
    {
      "tool": "query_options_flow",
      "quality": "partial",
      "as_of": "2026-03-08T03:31:00+00:00",
      "source": ["options_flow_daily"],
      "errors": [{"code": "PARTIAL_DATA"}],
      "data": {}
    }
  ]
}
```

Then delegate final synthesis to `mcp-result-interpreter`.

## Tool Binding (Executable Entry)
Use `rw` CLI as the concrete invocation layer.

Recommended setup (HTTP MCP gateway):

```bash
rw setup --server-url https://your-mcp-gateway.example.com
rw doctor
```

Then route intents with these commands:

- Company snapshot:
```bash
rw company --ticker NVDA
```

- Supply chain:
```bash
rw chain --entity NVDA --direction both --max-depth 3
```

- Macro:
```bash
rw macro --days 30
```

- Semantic research:
```bash
rw semantic --query "CoWoS capacity expansion 2026" --limit 5
```

- Generic direct call:
```bash
rw call --tool query_company --args '{"ticker":"NVDA"}'
```

- ETF flow radar:
```bash
rw etf --weeks 8
```

## Execution Notes
1. Always run one primary command first.
2. If `quality=partial`, execute one mapped fallback command.
3. If `quality=error` with `INVALID_ARGUMENT`, repair params before retry.
