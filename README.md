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
rw company --ticker NVDA
rw macro --days 30
rw semantic --query "CoWoS capacity expansion 2026" --limit 5
```

If you only have local stdio server:

```bash
rw setup --server-cmd "node /path/to/mcp_server/server.js"
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
- `mcp-query-router`
- `mcp-result-interpreter`
- `mcp-query-playbook`

Shared MCP mapping reference:
- `docs/mcp-tool-map.md`

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
./rw company --ticker NVDA
./rw call --tool query_company --args '{"ticker":"NVDA"}'
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
