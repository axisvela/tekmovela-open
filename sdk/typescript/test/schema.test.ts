import { test } from "node:test";
import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { dirname, join } from "node:path";
import Ajv2020Pkg from "ajv/dist/2020.js";
import addFormats from "ajv-formats";
const Ajv = Ajv2020Pkg.Ajv2020;
import { Reference } from "../src/contracts.ts";
import { EvidenceBundle, ItemKind } from "../src/evidence.ts";
import { HarnessVersion } from "../src/harness.ts";
import { UiRepairFailureEvidence } from "../src/ui-sfe.ts";
import { StaticUICapture } from "../src/ui-sfe.ts";
import { TraceEnvelope, TracePerspective } from "../src/trace.ts";

const ROOT = join(dirname(fileURLToPath(import.meta.url)), "..", "..", "..");
const DIGEST = "sha256:" + "a".repeat(64);

function loadSchema(rel: string): Record<string, unknown> {
  return JSON.parse(readFileSync(join(ROOT, "contracts", rel), "utf-8"));
}

function validator(schema: Record<string, unknown>): (doc: unknown) => void {
  const ajv = new Ajv({ allErrors: true, strict: false });
  addFormats(ajv);
  const validate = ajv.compile(schema);
  return (doc) => {
    if (!validate(doc)) {
      throw new Error(ajv.errorsText(validate.errors ?? []));
    }
  };
}

function sealedBundle(): EvidenceBundle {
  const b = new EvidenceBundle({
    id: "bundle.e0001",
    runRef: "run.1",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: DIGEST });
  b.addItem({ kind: ItemKind.Effect, uri: "effect.1", digest: "sha256:" + "b".repeat(64) });
  b.seal();
  return b;
}

function harnessVersion(): HarnessVersion {
  return new HarnessVersion({
    id: "checkout.double_spend",
    version: "v1",
    verificationContractRef: new Reference("VerificationContract", "checkout.atomicity", "v1", DIGEST),
    componentLock: {
      runner: { kind: "HarnessComponent", id: "r.1", version: "v1", digest: "sha256:" + "c".repeat(64) },
      adapter: { kind: "HarnessComponent", id: "a.1", version: "v1", digest: "sha256:" + "d".repeat(64) },
      oracle: { kind: "HarnessComponent", id: "o.1", version: "v1", digest: "sha256:" + "e".repeat(64) },
      reporter: { kind: "HarnessComponent", id: "p.1", version: "v1", digest: "sha256:" + "f".repeat(64) },
    },
    evidenceSchema: "tekmovela.evidence-bundle.v1.schema.json",
    seed: "s1",
  });
}

test("sealed EvidenceBundle.toJSON validates against the evidence-bundle schema", () => {
  const body = sealedBundle().toJSON();
  validator(loadSchema("evidence/tekmovela.evidence-bundle.v1.schema.json"))(body);
});

test("UiRepairFailureEvidence.body validates against the failure-evidence schema", async () => {
  const cap = new StaticUICapture({ screenshot: new Uint8Array([1]), dom: "<html>", alert: "boom" });
  const fe = await UiRepairFailureEvidence.capture(cap, "failure.ui.e0001", "local://trace/run.jsonl");
  validator(loadSchema("evidence/tekmovela.failure-evidence.v1.schema.json"))(fe.body());
});

test("TraceEnvelope.body validates against the trace-envelope schema", () => {
  const env = new TraceEnvelope("trace.e0001", "run.1", "tenant.1", "trace.1");
  env.addPerspective(TracePerspective.Execution, "local://trace/exec.jsonl", DIGEST);
  env.addPerspective(TracePerspective.Ppes, "local://trace/ppes.jsonl", "sha256:" + "b".repeat(64));
  validator(loadSchema("evidence/tekmovela.trace-envelope.v1.schema.json"))(env.body());
});

test("HarnessVersion.body validates against the harness-version schema", () => {
  validator(loadSchema("harness/tekmovela.harness-version.v1.schema.json"))(harnessVersion().body());
});
