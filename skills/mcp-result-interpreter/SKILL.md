---
name: mcp-result-interpreter
description: "按 MCP 契约解释结果，输出可追溯的金融研究结论与不确定性。"
---

# MCP Result Interpreter

## Purpose
Interpret MCP responses using contract-first logic:

1. `quality`
2. `errors`
3. `source`

Always keep research confidence explicit.

## Decision Policy

### `quality = complete`
- Produce normal conclusion.
- Cite primary evidence fields.
- Include `as_of` and `source`.

### `quality = partial`
- Produce directional conclusion only.
- Explicitly state missing/failed parts from `errors`.
- Mark confidence as reduced.
- Recommend one concrete补查 query.

### `quality = empty`
- State that current filters returned no usable data.
- Suggest relaxed filters or adjacent tools.
- Avoid speculative conclusion.

### `quality = error`
- Treat conclusion as invalid.
- First action: fix arguments if `INVALID_ARGUMENT`.
- Else suggest retry/backoff for `INTERNAL_ERROR`.

## Error Handling

### `INVALID_ARGUMENT`
- Provide a concise parameter correction checklist.
- Ask for missing required input before retry.

### `PARTIAL_DATA`
- Continue with available data.
- Name the missing stage or field if present in `details`.
- Add an uncertainty note in the final output.

### `INTERNAL_ERROR`
- Recommend retry with backoff.
- If repeated, switch to adjacent tool path.

## Financial Output Template
Use this exact section order:

1. `结论`
2. `数据状态` (`quality`, `as_of`)
3. `证据` (key fields)
4. `风险与不确定性` (from `errors`)
5. `下一步补查` (tool + reason)

## Compliance Rules
Required:
- Show `as_of` explicitly.
- Show `source` for traceability.
- Separate facts from inference language.

Forbidden:
- Do not give deterministic trading instructions.
- Do not hide partial/empty/error conditions.

## Inference Labels
When inferring across tools:
- Prefix with `推断` and keep one sentence.
- Tie inference back to explicit evidence fields.

## Runtime Integration
This interpreter consumes payloads returned by `rw` commands.

Recommended invocation pattern:

```bash
rw company --ticker NVDA --json
rw macro --days 30 --json
```

Interpreter input expectation:
- A single contract payload with top-level fields:
  - `data`
  - `as_of`
  - `quality`
  - `source`
  - `errors`

If upstream output is not a valid contract payload, treat as transport failure and request retry.
