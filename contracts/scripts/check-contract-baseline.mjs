#!/usr/bin/env node
// Verify that every published contract schema still matches the baseline
// snapshot. Any drift is a breaking change and must be released as a new
// schema version, not an in-place edit.
import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { readdirSync, readFileSync } from "node:fs";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const baseline = JSON.parse(readFileSync(join(root, "baseline", "schemas.json"), "utf8"));

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

const sep = process.platform === "win32" ? "\\" : "/";
const current = {};
for (const file of walk(root)) {
  if (file.includes(`${sep}node_modules${sep}`)) continue;
  const content = readFileSync(file, "utf8");
  current[relative(root, file).replace(/\\/g, "/")] = createHash("sha256").update(content).digest("hex");
}

const drift = [];
for (const [file, digest] of Object.entries(current)) {
  if (!(file in baseline)) {
    drift.push(`NEW schema not in baseline: ${file}`);
  } else if (baseline[file] !== digest) {
    drift.push(`EDITED published schema: ${file} (breaking change; release a new version)`);
  }
}
for (const file of Object.keys(baseline)) {
  if (!(file in current)) {
    drift.push(`MISSING schema removed from baseline: ${file}`);
  }
}

if (drift.length > 0) {
  console.error(drift.join("\n"));
  process.exit(1);
}
console.log(`compatibility: ${Object.keys(current).length} schemas match baseline`);
assert.ok(Object.keys(current).length > 0);
