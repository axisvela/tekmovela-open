#!/usr/bin/env node
// Validates version.json and asserts it is consistent across the open repo.
import { readFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const root = join(dirname(fileURLToPath(import.meta.url)), "..");
const versionPath = join(root, "version.json");
const v = JSON.parse(readFileSync(versionPath, "utf8"));

const semver = /^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$/;
for (const key of ["version", "core"]) {
  if (!semver.test(v[key] ?? "")) {
    console.error(`version.json "${key}" must be semver, got ${v[key]}`);
    process.exit(1);
  }
}
if (v.product !== "tekmovela" || v.edition !== "open") {
  console.error("version.json must declare product=tekmovela edition=open");
  process.exit(1);
}
console.log(`tekmovela-open ${v.version} (core ${v.core}) ok`);
