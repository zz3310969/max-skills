#!/usr/bin/env node
import { readAllSkills } from "./lib-skill.mjs";

const json = process.argv.includes("--json");
const skills = readAllSkills();

if (json) {
  const result = skills.map((s) => ({
    dir: s.dirName,
    name: s.meta.name ?? null,
    description: s.meta.description ?? null,
    hasSkillFile: s.exists
  }));
  console.log(JSON.stringify(result, null, 2));
  process.exit(0);
}

if (!skills.length) {
  console.log("No skills found under ./skills");
  process.exit(0);
}

for (const s of skills) {
  const name = s.meta.name ?? "(missing name)";
  const description = s.meta.description ?? "(missing description)";
  const status = s.exists ? "" : " [missing SKILL.md]";
  console.log(`${s.dirName}${status}`);
  console.log(`  name: ${name}`);
  console.log(`  description: ${description}`);
}

