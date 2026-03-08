# max-skills

A personal skills repository structured to work with:

```bash
npx skills add <repo> --skill <name>
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
```

`npx skills add` discovers skills by scanning for `SKILL.md` files and their frontmatter.

## Included Skills

- `research-summary`
- `mcp-query-router`
- `mcp-result-interpreter`
- `mcp-query-playbook`

Shared MCP mapping reference:
- `docs/mcp-tool-map.md`

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
