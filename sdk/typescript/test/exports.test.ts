import { test } from "node:test";
import assert from "node:assert/strict";
import * as mod from "../dist/index.mjs";

const PUBLIC_VALUES = [
  "API_VERSION",
  "Reference",
  "ContractError",
  "EvidenceBundle",
  "ItemKind",
  "HarnessVersion",
  "LocalRunner",
  "COMPONENT_KINDS",
  "DETERMINISM_MODES",
  "StaticUICapture",
  "UiRepairFailureEvidence",
  "REDACTION_STATES",
  "TraceEnvelope",
  "TracePerspective",
];

test("built package index re-exports every public value", () => {
  for (const name of PUBLIC_VALUES) {
    assert.ok(name in mod, `missing export: ${name}`);
  }
});

test("export const collections are present and non-empty", () => {
  assert.ok(Array.isArray(mod.COMPONENT_KINDS) && mod.COMPONENT_KINDS.length > 0);
  assert.ok(Array.isArray(mod.DETERMINISM_MODES) && mod.DETERMINISM_MODES.length > 0);
  assert.ok(Array.isArray(mod.REDACTION_STATES) && mod.REDACTION_STATES.length > 0);
});
