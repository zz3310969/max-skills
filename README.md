# max-skills

A personal skills repository structured to work with:

```bash
npx skills add <repo> --skill <name>
```

## Quickstart (Out Of The Box)

1. Install skills into your target project:

```bash
npx skills add https://github.com/zz3310969/max-skills --skill '*'
```

2. Install `rw` CLI:

```bash
curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/scripts/install-rw.sh | bash
```

3. Point `rw` to your MCP endpoint (recommended HTTP mode):

```bash
rw setup --server-url https://your-mcp-gateway.example.com
rw doctor
rw tools
```

4. Run financial queries:

```bash
rw call --tool read_company_overview --args '{"ticker":"NVDA"}'
rw call --tool read_market_snapshot --args '{"tickers":["NVDA","AAPL"],"fields":["price","technicals"]}'
rw call --tool read_reports --args '{"ticker":"NVDA","limit":20}'
```

If you only have local stdio server:

```bash
rw setup --server-cmd "node /path/to/mcp_server/server.js"
```

## OpenClaw Skill Install

Use this path if you want to install skills directly into OpenClaw without `npx skills add`.

Target directory:

```bash
~/.openclaw/workspace/skills/<skill-name>/SKILL.md
```

Install all skills:

```bash
mkdir -p ~/.openclaw/workspace/skills
for s in research-summary mcp-us-equities-read mcp-us-equities-ops mcp-us-equities-response; do
  mkdir -p ~/.openclaw/workspace/skills/$s
  curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/skills/$s/SKILL.md \
    -o ~/.openclaw/workspace/skills/$s/SKILL.md
done
```

Install one skill (`mcp-us-equities-read` example):

```bash
mkdir -p ~/.openclaw/workspace/skills/mcp-us-equities-read
curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/skills/mcp-us-equities-read/SKILL.md \
  -o ~/.openclaw/workspace/skills/mcp-us-equities-read/SKILL.md
```

Verify installation:

```bash
find ~/.openclaw/workspace/skills -name SKILL.md
```

For MCP querying skills (`mcp-us-equities-read` / `mcp-us-equities-ops`), install `rw` too:

```bash
curl -fsSL https://raw.githubusercontent.com/zz3310969/max-skills/main/scripts/install-rw.sh | bash
```

## Repository Structure

```text
skills/
  <skill-name>/
    SKILL.md
scripts/
  create-skill.mjs
  list-skills.mjs
  validate-skills.mjs
cli/
  rw/
```

`npx skills add` discovers skills by scanning for `SKILL.md` files and their frontmatter.

## Included Skills

- `research-summary`
- `mcp-us-equities-read`
- `mcp-us-equities-ops`
- `mcp-us-equities-response`

Shared MCP mapping reference:
- `docs/mcp-tool-map-v2.md`

## Skills Mode (Synced To MCP Source)

This repo's MCP-related skills are synced against your MCP server docs:
- query guide (tool names + parameters)
- response contract (`ok/data/meta/warnings/error`)

When MCP source changes, update skill routing by this order:
1. Query guide (tool names + parameters)
2. Response contract (`ok/data/meta/warnings/error`)
3. `docs/mcp-tool-map-v2.md`
4. `skills/mcp-us-equities-read/SKILL.md` and `skills/mcp-us-equities-ops/SKILL.md`

## Go CLI (`rw`) For MCP Calls

Location:
- `cli/rw`

Build from source:

```bash
cd cli/rw
go build -o rw .
```

Examples:

```bash
./rw doctor
./rw tools
./rw call --tool read_company_overview --args '{"ticker":"NVDA"}'
./rw call --tool read_market_snapshot --args '{"tickers":["NVDA","AAPL"],"fields":["price","technicals"]}'
```

## Skill Format

Each skill should live in its own directory under `skills/` with a `SKILL.md`:

```md
---
name: my-skill
description: "One-line purpose."
---
```

Recommendation: keep `frontmatter.name` the same as directory name.

## Local Development

Create a skill:

```bash
npm run skills:new -- --name my-skill --description "What this skill does"
```

List skills:

```bash
npm run skills:list
```

Validate skills:

```bash
npm run skills:validate
```

List skills in JSON:

```bash
npm run skills:list -- --json
```

## Use From Any Project

After pushing this repository to GitHub:

```bash
npx skills add https://github.com/<your-org>/<your-repo> --skill my-skill
```

Install all skills in the repo:

```bash
npx skills add https://github.com/<your-org>/<your-repo> --skill '*'
```

Preview discoverable skills without installing:

```bash
npx skills add https://github.com/<your-org>/<your-repo> --list
```
