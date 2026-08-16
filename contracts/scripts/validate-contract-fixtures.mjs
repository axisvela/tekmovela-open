import assert from "node:assert/strict";
import { readdir, readFile } from "node:fs/promises";
import { dirname, join, relative } from "node:path";
import { fileURLToPath } from "node:url";
import Ajv from "ajv/dist/2020.js";
import addFormats from "ajv-formats";

const contractsRoot = dirname(dirname(fileURLToPath(import.meta.url)));
const ajv = new Ajv({ strict: false, allErrors: true });
addFormats(ajv);

async function listSchemas() {
  const out = [];
  for (const domain of await readdir(contractsRoot, { withFileTypes: true })) {
    if (!domain.isDirectory() || domain.name.startsWith(".") || domain.name === "node_modules" || domain.name === "scripts") {
      continue;
    }
    const domainDir = join(contractsRoot, domain.name);
    for (const entry of await readdir(domainDir, { withFileTypes: true })) {
      if (entry.isFile() && entry.name.endsWith(".schema.json")) {
        const path = join(domainDir, entry.name);
        const doc = JSON.parse(await readFile(path, "utf8"));
        out.push({ domain: domain.name, file: entry.name, path, doc });
      }
    }
  }
  return out;
}

const schemaByKind = new Map();

for (const s of await listSchemas()) {
  const props = s.doc.properties ?? {};
  const kind = props.kind?.const ?? props.kind?.enum?.[0];
  const api_version = props.api_version?.const ?? props.api_version?.enum?.[0];
  assert.ok(kind, `schema ${s.file} must declare properties.kind`);
  assert.ok(api_version, `schema ${s.file} must declare properties.api_version`);
  const key = `${api_version}/${kind}`;
  assert.equal(schemaByKind.has(key), false, `duplicate schema kind ${key} in ${s.file}`);
  schemaByKind.set(key, s);
  ajv.addSchema(s.doc, key);
}

async function listFixtures() {
  const out = [];
  for (const domain of await readdir(contractsRoot, { withFileTypes: true })) {
    if (!domain.isDirectory() || domain.name.startsWith(".") || domain.name === "node_modules" || domain.name === "scripts") {
      continue;
    }
    for (const expect of ["valid", "invalid"]) {
      const dir = join(contractsRoot, domain.name, "fixtures", expect);
      let files = [];
      try {
        files = await readdir(dir);
      } catch {
        continue;
      }
      for (const f of files) {
        if (f.endsWith(".json")) {
          out.push({ domain: domain.name, expect, path: join(dir, f) });
        }
      }
    }
  }
  return out;
}

let passed = 0;
let failed = 0;
const problems = [];

// Track per-schema coverage so that a schema without fixtures is a hard failure.
const coveredByValid = new Set();
const coveredByInvalid = new Set();
const familiesWithValidFixture = new Set();

for (const fx of await listFixtures()) {
  const doc = JSON.parse(await readFile(fx.path, "utf8"));
  const key = `${doc.api_version}/${doc.kind}`;
  const schema = schemaByKind.get(key);
  assert.ok(schema, `no schema for ${key} referenced by ${relative(contractsRoot, fx.path)}`);
  const validate = ajv.getSchema(key);
  const ok = validate(doc);
  const label = `${fx.domain}/fixtures/${fx.expect}/${fx.path.split(/[\\/]/).pop()}`;
  if (fx.expect === "valid") {
    coveredByValid.add(key);
    if (ok && key === "tekmovela.io/v1alpha1/BenchmarkDefinition") {
      familiesWithValidFixture.add(doc.family);
    }
    if (ok) {
      passed += 1;
    } else {
      failed += 1;
      problems.push(`VALID fixture failed: ${label}\n  ${JSON.stringify(validate.errors)}`);
    }
  } else {
    coveredByInvalid.add(key);
    if (!ok) {
      passed += 1;
    } else {
      failed += 1;
      problems.push(`INVALID fixture did not fail: ${label}`);
    }
  }
}

for (const [key, s] of schemaByKind) {
  if (!coveredByValid.has(key)) {
    failed += 1;
    problems.push(`Schema ${key} (${s.file}) has no VALID fixture`);
  }
  if (!coveredByInvalid.has(key)) {
    failed += 1;
    problems.push(`Schema ${key} (${s.file}) has no INVALID fixture`);
  }
}

// TEK-Bench family coverage: every family enumerated in the BenchmarkDefinition
// schema must have at least one valid definition fixture, so a benchmark family
// can never silently ship without a frozen, comparable definition.
{
  const bench = schemaByKind.get("tekmovela.io/v1alpha1/BenchmarkDefinition");
  if (bench) {
    const families = bench.doc.properties?.family?.enum ?? [];
    assert.ok(families.length >= 12, `BenchmarkDefinition family enum has ${families.length} entries, want >= 12`);
    for (const family of families) {
      if (!familiesWithValidFixture.has(family)) {
        failed += 1;
        problems.push(`Benchmark family ${family} has no VALID BenchmarkDefinition fixture`);
      }
    }
  }
}

if (problems.length > 0) {
  console.error(problems.join("\n"));
}
assert.equal(failed, 0, `${failed} contract fixture check(s) failed`);
console.log(`contract fixtures: ${passed} passed, ${failed} failed`);