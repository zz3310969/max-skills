#!/usr/bin/env node
import { readAllSkills } from "./lib-skill.mjs";

const skills = readAllSkills();

if (!skills.length) {
  console.error("No skills found under ./skills");
  process.exit(1);
}

const issues = [];

for (const s of skills) {
  if (!s.exists) {
    issues.push(`${s.dirName}: missing SKILL.md`);
    continue;
  }
  if (!s.meta.name) issues.push(`${s.dirName}: missing frontmatter 'name'`);
  if (!s.meta.description) {
    issues.push(`${s.dirName}: missing frontmatter 'description'`);
  }
  if (s.meta.name && s.meta.name !== s.dirName) {
    issues.push(
      `${s.dirName}: frontmatter name "${s.meta.name}" should match directory name`
    );
  }
}

if (issues.length) {
  console.error("Skill validation failed:");
  for (const issue of issues) console.error(`- ${issue}`);
  process.exit(1);
}

console.log(`Skill validation passed (${skills.length} skill(s)).`);

