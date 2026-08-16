#!/usr/bin/env node
// Snapshot the canonical digest of every published contract schema into
// baseline/schemas.json. Published contract versions are immutable; a schema
// edit must therefore fail `check:compatibility` until a new version ships.
import { createHash } from "node:crypto";
import { mkdirSync, readdirSync, readFileSync, writeFileSync } from "node:fs";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const baselineDir = join(root, "baseline");
const baselinePath = join(baselineDir, "schemas.json");

function walk(dir) {
  const out = [];
  for (const entry of readdirSync(dir, { withFileTypes: true })) {
    const full = join(dir, entry.name);
    if (entry.isDirectory()) {
      out.push(...walk(full));
    } else if (entry.name.endsWith(".schema.json")) {
      out.push(full);
    }
  }
  return out;
}

const schemas = {};
for (const file of walk(root)) {
  if (file.includes(`${sep()}node_modules${sep()}`)) continue;
  const content = readFileSync(file, "utf8");
  schemas[relative(root, file).replace(/\\/g, "/")] = createHash("sha256").update(content).digest("hex");
}

mkdirSync(baselineDir, { recursive: true });
writeFileSync(baselinePath, JSON.stringify(schemas, null, 2) + "\n");
console.log(`baseline snapshot: ${Object.keys(schemas).length} schemas -> baseline/schemas.json`);

function sep() {
  return process.platform === "win32" ? "\\" : "/";
}
