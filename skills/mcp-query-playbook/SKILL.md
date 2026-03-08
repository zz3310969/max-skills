---
name: mcp-query-playbook
description: "金融问题到 MCP 查询路径的操作手册（主查/补查/停查）。"
---

# MCP Query Playbook

## Purpose
Provide reusable financial query playbooks so the router can execute consistent, auditable query chains.
Tool selection and parameters must follow `~/Max/research-warehouse/docs/query-guide.md`.

Each playbook defines:
- `question_pattern`
- `primary_tool`
- `confirmatory_tool`
- `stop_condition`
- `output_focus`

## Core Playbooks

### 1) 财报后异动归因
- `question_pattern`: 财报后涨跌原因、业绩是否超预期、电话会关键信号
- `primary_tool`: `get_earnings_call`
- `confirmatory_tool`: `query_options_flow`, `query_technicals`
- `stop_condition`: 出现 `quality=error` 或两次连续 `partial`
- `output_focus`: 管理层表述、期权行为、价格结构是否一致

### 2) 产业链冲击传导
- `question_pattern`: 上游涨价/断供如何影响核心公司
- `primary_tool`: `query_bottleneck`
- `confirmatory_tool`: `query_supply_chain`, `query_company`
- `stop_condition`: 关键公司字段持续缺失
- `output_focus`: 传导路径、受影响公司、活跃信号

### 3) 宏观冲击行业
- `question_pattern`: 利率/通胀/风险偏好对行业或主题的影响
- `primary_tool`: `query_macro`
- `confirmatory_tool`: `search_semantic`
- `stop_condition`: 宏观指标为空且语义检索为空
- `output_focus`: 指标方向、宏观信号、历史可比叙事

### 4) 机构资金动向
- `question_pattern`: 近期是否有机构资金偏好变化
- `primary_tool`: `query_etf_flow`
- `confirmatory_tool`: `query_options_flow`
- `stop_condition`: flow 工具连续不可用
- `output_focus`: 资金流强弱、持仓变化、异常合约

### 5) 个股风险排查
- `question_pattern`: 某标的当前主要风险点是什么
- `primary_tool`: `query_company`
- `confirmatory_tool`: `query_macro`, `query_technicals`
- `stop_condition`: 公司主数据缺失
- `output_focus`: 基本面风险、宏观暴露、技术弱化信号

### 6) 研报观点验证
- `question_pattern`: 某观点是否被数据支持
- `primary_tool`: `get_research_reports`
- `confirmatory_tool`: `search_semantic`, `query_company`
- `stop_condition`: 无可用研报且检索为空
- `output_focus`: 观点共识、反例证据、数据一致性

### 7) 技术面与基本面冲突诊断
- `question_pattern`: 基本面看多但走势转弱是否合理
- `primary_tool`: `query_technicals`
- `confirmatory_tool`: `query_company`, `query_options_flow`
- `stop_condition`: 技术数据为空
- `output_focus`: 趋势结构、资金行为、基本面背离解释

### 8) 标的发现到深挖
- `question_pattern`: 先找标的，再挑选深入研究对象
- `primary_tool`: `search_companies`
- `confirmatory_tool`: `query_company`, `query_supply_chain`
- `stop_condition`: 搜索结果为空
- `output_focus`: 候选池、基本特征、链条位置

## Quality Guardrails
1. Any `quality=error` on primary tool: stop and repair input first.
2. Two consecutive `partial`: switch to "information insufficient" mode.
3. `empty` on primary tool: try one adjacent-tool fallback only.

## Parameter Canonical Rules
- `query_supply_chain`: use `entity` (not `ticker`), optional `direction/max_depth`
- `query_bottleneck`: use `domain` in `memory|photonics|packaging|power|gpu`
- `query_options_flow`: use `ticker` + optional `days`
- `query_etf_flow`: optional `ticker`, optional `weeks`
- `query_technicals`: use `ticker` + optional `days`

## Output Focus Rules
- State evidence first, inference second.
- If evidence is cross-tool inconsistent, mark as unresolved.
- Always include next actionable query when confidence is low.

## Command-Level Examples (`rw`)
First configure once:

```bash
rw setup --server-url https://your-mcp-gateway.example.com
rw doctor
```

### 财报后异动归因
```bash
rw earnings --ticker NVDA
rw options --ticker NVDA --days 30
rw technicals --ticker NVDA --days 7
```

### 产业链冲击传导
```bash
rw bottleneck --domain packaging
rw chain --entity NVDA --direction both --max-depth 3
rw company --ticker NVDA
```

### 宏观冲击行业
```bash
rw macro --days 30
rw semantic --query "semiconductor demand under high rates" --limit 5
```

### 标的发现到深挖
```bash
rw search --sector photonics --tier chokepoint
rw company --ticker LITE
rw chain --entity LITE --direction upstream
```
