import fs from "node:fs";
import path from "node:path";

export const skillsRoot = path.resolve(process.cwd(), "skills");

export function parseFrontmatter(raw) {
  const match = raw.match(/^---\n([\s\S]*?)\n---\n?/);
  if (!match) return {};
  const map = {};
  for (const line of match[1].split("\n")) {
    const trimmed = line.trim();
    if (!trimmed || trimmed.startsWith("#")) continue;
    const idx = trimmed.indexOf(":");
    if (idx < 0) continue;
    const key = trimmed.slice(0, idx).trim();
    let value = trimmed.slice(idx + 1).trim();
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1);
    }
    map[key] = value;
  }
  return map;
}

export function readAllSkills() {
  if (!fs.existsSync(skillsRoot)) return [];
  const entries = fs.readdirSync(skillsRoot, { withFileTypes: true });
  return entries
    .filter((d) => d.isDirectory())
    .map((d) => {
      const dirName = d.name;
      const skillDir = path.join(skillsRoot, dirName);
      const skillFile = path.join(skillDir, "SKILL.md");
      if (!fs.existsSync(skillFile)) {
        return { dirName, skillFile, exists: false, meta: {} };
      }
      const raw = fs.readFileSync(skillFile, "utf8");
      return { dirName, skillFile, exists: true, meta: parseFrontmatter(raw) };
    })
    .sort((a, b) => a.dirName.localeCompare(b.dirName));
}

