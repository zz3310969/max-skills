#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { skillsRoot } from "./lib-skill.mjs";

function getArg(flag) {
  const idx = process.argv.indexOf(flag);
  if (idx < 0 || idx + 1 >= process.argv.length) return "";
  return process.argv[idx + 1];
}

const name = getArg("--name").trim();
const description = getArg("--description").trim();

if (!name || !description) {
  console.error(
    "Usage: npm run skills:new -- --name <skill-name> --description \"...\""
  );
  process.exit(1);
}

if (!/^[a-z0-9-]+$/.test(name)) {
  console.error("Invalid --name. Use lowercase letters, numbers, and hyphens.");
  process.exit(1);
}

const dir = path.join(skillsRoot, name);
const file = path.join(dir, "SKILL.md");

if (fs.existsSync(dir)) {
  console.error(`Skill already exists: ${dir}`);
  process.exit(1);
}

fs.mkdirSync(dir, { recursive: true });

const content = `---
name: ${name}
description: "${description}"
---

# ${name}

## Purpose
Explain when this skill should be used.

## Instructions
- Add concrete, step-by-step instructions for the agent.
- Keep examples short and practical.
- Include constraints and failure handling.
`;

fs.writeFileSync(file, content, "utf8");
console.log(`Created ${file}`);

