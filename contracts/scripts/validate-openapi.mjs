#!/usr/bin/env node
// Validates the OpenAPI document against the contract and API conventions:
//   - info.version matches the contract api_version
//   - every operation has an operationId
//   - every request/response body that $refs a contract schema file resolves
//   - every non-2xx response references the shared ProblemDetail envelope
//   - the documented resource families are present
import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { parse } from "yaml";

const root = dirname(dirname(fileURLToPath(import.meta.url)));
const openapiPath = resolve(root, "../docs/api/tekmovela.v1alpha1.openapi.yaml");
const apiDir = dirname(openapiPath);
const document = parse(await readFile(openapiPath, "utf8"));

assert.equal(document.info.version, "v1alpha1", "OpenAPI info.version must match the contract api_version");

async function assertContractRefExists(ref, label) {
  const target = resolve(apiDir, ref);
  await readFile(target, "utf8");
  return target;
}

let operationCount = 0;
for (const [path, item] of Object.entries(document.paths)) {
  for (const [method, op] of Object.entries(item)) {
    operationCount += 1;
    const label = `${method.toUpperCase()} ${path}`;
    assert.ok(op.operationId, `operationId missing for ${label}`);

    if (op.requestBody) {
      for (const body of Object.values(op.requestBody.content)) {
        if (body.schema?.$ref?.startsWith("../../")) {
          await assertContractRefExists(body.schema.$ref, `${label} request body`);
        }
      }
    }

    for (const [status, resp] of Object.entries(op.responses)) {
      if (Number(status) >= 200 && Number(status) < 300) {
        for (const body of Object.values(resp.content ?? {})) {
          if (body.schema?.$ref?.startsWith("../../")) {
            await assertContractRefExists(body.schema.$ref, `${label} ${status} response`);
          }
        }
      } else {
        assert.ok(resp.$ref?.startsWith("#/components/responses/"),
          `${label} ${status} must reference a components/responses entry`);
        const name = resp.$ref.slice("#/components/responses/".length);
        const envelope = document.components.responses[name];
        assert.ok(envelope, `${label} ${status} references unknown response ${name}`);
        const problemRef = envelope.content?.["application/json"]?.schema?.$ref;
        assert.equal(problemRef, "#/components/schemas/ProblemDetail",
          `${label} ${status} must use the ProblemDetail envelope`);
      }
    }
  }
}

const requiredFamilies = [
  "/verification-contracts",
  "/verification-contracts/{id}/versions/{version}",
  "/assurance-cases",
  "/gate-decisions",
  "/release-candidates",
  "/attestations",
  "/evidence-records/{id}/versions/{version}",
  "/coverage-statements",
  "/waivers",
  "/behavioral-bounds",
  "/anomaly-signals",
  "/reverification-requests",
  "/accountability-records",
  "/claims",
  "/coverage-debts",
  "/failure-clusters",
  "/regression-candidates",
  "/role-profiles",
  "/development-contracts",
  "/engineering-state-links",
  "/contributor-decisions",
  "/evaluation-sets",
  "/evidence-records",
  "/contributor-evaluations",
  "/acceptance-criteria",
  "/constraints",
  "/requirement-trace-links",
  "/harness-definitions",
  "/harness-versions",
  "/scenarios",
  "/runs",
  "/test-control-loops",
  "/fixtures",
  "/oracle-profiles",
  "/harness-compositions",
  "/test-budgets",
  "/intent-spec-refs",
  "/assurance-evidence",
  "/cognitive-traces",
  "/improvement-actions",
  "/incident-refs",
  "/benchmark-definitions",
  "/benchmark-runs",
];
for (const family of requiredFamilies) {
  assert.ok(document.paths[family], `missing documented resource ${family}`);
}

// Server route parity, both directions, from backend/internal/server/server.go:
//   - every route the server registers must have a documented OpenAPI path
//   - every documented OpenAPI path must be served by the server
// so a server route without an API description, or a documented API that no
// handler serves, both fail CI.
{
  const serverFile = resolve(root, "../backend/internal/server/server.go");
  const src = await readFile(serverFile, "utf8");
  const normalize = (p) => p.replace(/\{[a-zA-Z0-9_]+\}/g, "{param}");
  const serverRoutes = [...src.matchAll(/registerRecordRoutes\("(?:GET|POST) \/v1(\S+)"/g)].map((m) => m[1]);
  const directRoutes = [...src.matchAll(/mux\.Handle\("(?:GET|POST) \/v1(\S+)"/g)].map((m) => m[1]);
  const allRoutes = [...serverRoutes, ...directRoutes];
  assert.ok(allRoutes.length >= 40, `parsed ${allRoutes.length} server routes, want >= 40`);
  const documentedPaths = Object.keys(document.paths);
  for (const route of allRoutes) {
    const documented = documentedPaths.some((p) => normalize(p) === normalize(route));
    assert.ok(documented, `server route /v1${route} is not documented in OpenAPI`);
  }
  for (const p of documentedPaths) {
    const served = allRoutes.some((r) => normalize(r) === normalize(p));
    assert.ok(served, `documented OpenAPI path ${p} is not served by the server`);
  }
}

console.log(`openapi: ${operationCount} operations validated (operationIds, contract refs, ProblemDetail envelope, resource coverage, server route parity)`);
