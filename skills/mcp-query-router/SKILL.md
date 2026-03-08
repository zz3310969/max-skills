---
name: mcp-query-router
description: "金融数据查询总路由，负责选工具、组参、调用与补查编排。"
---

# MCP Query Router

## Purpose
Route financial research requests to the right MCP tools, validate parameters, and build a minimal query plan.

This skill focuses on query orchestration:
- What to query
- Which tool to use first
- When to run fallback queries

It does not produce final user-facing conclusions by itself.

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
- Required params: macro topic or default preset
- Optional params: date window, market context

### 3) Research and semantic lookup
- Primary: `search_semantic`
- Fallback: `get_research_reports`
- Required params: `query`
- Optional params: `ticker`, date range, source preference

### 4) Earnings event and transcript context
- Primary: `get_earnings_call`
- Fallback: `query_options_flow`, `query_technicals`
- Required params: `ticker`
- Optional params: quarter, year

### 5) Flow and positioning
- Primary: `query_options_flow` or `query_etf_flow`
- Fallback: the other flow tool, then `query_technicals`
- Required params: `ticker` for single-name flow; optional for radar mode
- Optional params: lookback windows

### 6) Technicals and trend structure
- Primary: `query_technicals`
- Fallback: `query_options_flow`, `query_macro`
- Required params: `ticker`
- Optional params: timeframe

### 7) Company discovery and screening
- Primary: `search_companies`
- Fallback: `query_company`
- Required params: search string or filters
- Optional params: sector, region

## Parameter Precheck
Before calling MCP:

1. Ensure all required params are present.
2. Validate enum-like inputs (timeframe, mode, range).
3. Normalize ticker casing (`upper()` convention).
4. If invalid or missing, ask for parameter correction before calling.

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

