# MCP Tool Map (Shared)

This shared map is referenced by `mcp-query-router` and `mcp-query-playbook`.
Source of truth: `~/Max/research-warehouse/docs/query-guide.md`.

## Schema

Each tool mapping should include:

- `tool`
- `intent_tags`
- `required_params`
- `optional_params`
- `key_fields_for_quality`
- `fallback_on_partial`
- `fallback_on_error`
- `evidence_fields`
- `financial_caveats`

## Mappings

### `query_company`
- `intent_tags`: company, fundamentals, risk-review
- `required_params`: ticker
- `optional_params`: none
- `key_fields_for_quality`: company, supply_chain, active_signals
- `fallback_on_partial`: query_supply_chain, query_macro
- `fallback_on_error`: search_companies
- `evidence_fields`: company, active_signals
- `financial_caveats`: company data may lag latest market move

### `query_supply_chain`
- `intent_tags`: supply-chain, transmission, exposure
- `required_params`: entity
- `optional_params`: direction, max_depth
- `key_fields_for_quality`: upstream, downstream, active_signals
- `fallback_on_partial`: query_company, query_bottleneck
- `fallback_on_error`: query_company
- `evidence_fields`: upstream, downstream, active_signals
- `financial_caveats`: relationship coverage may be incomplete

### `query_bottleneck`
- `intent_tags`: bottleneck, chokepoint, material-risk
- `required_params`: domain
- `optional_params`: none
- `key_fields_for_quality`: key_companies, material_trends, active_signals
- `fallback_on_partial`: query_supply_chain
- `fallback_on_error`: query_macro
- `evidence_fields`: key_companies, material_trends
- `financial_caveats`: trend signals are not direct causality proof

### `query_macro`
- `intent_tags`: macro, regime, risk
- `required_params`: none
- `optional_params`: days
- `key_fields_for_quality`: indicators, macro_signals
- `fallback_on_partial`: search_semantic
- `fallback_on_error`: query_technicals
- `evidence_fields`: indicators, macro_signals
- `financial_caveats`: macro indicators are often lagging

### `search_companies`
- `intent_tags`: discovery, screening
- `required_params`: none
- `optional_params`: sector, tier, min_cap, max_cap, country
- `key_fields_for_quality`: companies
- `fallback_on_partial`: query_company
- `fallback_on_error`: none
- `evidence_fields`: companies
- `financial_caveats`: search hit quality depends on query specificity

### `search_semantic`
- `intent_tags`: semantic, narrative, thesis-check
- `required_params`: query
- `optional_params`: limit, source_filter
- `key_fields_for_quality`: research_reports, expert_posts
- `fallback_on_partial`: get_research_reports
- `fallback_on_error`: query_company
- `evidence_fields`: research_reports, expert_posts
- `financial_caveats`: text relevance does not imply correctness

### `get_earnings_call`
- `intent_tags`: earnings, management-guidance
- `required_params`: ticker
- `optional_params`: quarter
- `key_fields_for_quality`: ticker, transcript_preview
- `fallback_on_partial`: query_options_flow
- `fallback_on_error`: get_research_reports
- `evidence_fields`: transcript_preview
- `financial_caveats`: statement tone may differ from realized results

### `get_research_reports`
- `intent_tags`: reports, consensus, thesis
- `required_params`: none
- `optional_params`: ticker, topic, source, limit
- `key_fields_for_quality`: reports
- `fallback_on_partial`: search_semantic
- `fallback_on_error`: query_company
- `evidence_fields`: reports
- `financial_caveats`: report timeliness and bias vary by source

### `query_options_flow`
- `intent_tags`: options, sentiment, positioning
- `required_params`: ticker
- `optional_params`: days
- `key_fields_for_quality`: latest_snapshot, flow_history, unusual_contracts
- `fallback_on_partial`: query_etf_flow, query_technicals
- `fallback_on_error`: query_technicals
- `evidence_fields`: latest_snapshot, unusual_contracts
- `financial_caveats`: flow spikes can be hedging, not directional bets

### `query_etf_flow`
- `intent_tags`: etf, fund-flow, positioning
- `required_params`: none
- `optional_params`: ticker, weeks
- `key_fields_for_quality`: aum_history, latest_holdings, etf_flow_radar
- `fallback_on_partial`: query_options_flow
- `fallback_on_error`: query_macro
- `evidence_fields`: etf_flow_radar, latest_holdings
- `financial_caveats`: holdings snapshots may be delayed

### `query_technicals`
- `intent_tags`: technicals, trend, momentum
- `required_params`: ticker
- `optional_params`: days
- `key_fields_for_quality`: latest_technicals, history
- `fallback_on_partial`: query_options_flow
- `fallback_on_error`: query_company
- `evidence_fields`: latest_technicals, history
- `financial_caveats`: technical patterns have regime sensitivity
