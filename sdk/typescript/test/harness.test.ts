import { test } from "node:test";
import assert from "node:assert/strict";
import { HarnessVersion, LocalRunner } from "../src/harness.ts";
import type { RuntimeAdapter } from "../src/harness.ts";
import { ContractError, Reference } from "../src/contracts.ts";
import { ItemKind } from "../src/evidence.ts";

const D = "sha256:" + "a".repeat(64);
const VCONTRACT = new Reference("VerificationContract", "checkout.atomicity", "v1", D);

const GOLDEN_HARNESS_VERSION_DIGEST = "sha256:bc8520405b1d90daea66d17a62c415f0edee8ecb4a52b9ed1460bc6a5a9b0499";

function componentLock(): Record<string, { kind: string; id: string; version: string; digest: string }> {
  return {
    runner: { kind: "HarnessComponent", id: "runner.deterministic", version: "v1", digest: "sha256:" + "c".repeat(64) },
    adapter: { kind: "HarnessComponent", id: "adapter.praxovela", version: "v1", digest: "sha256:" + "d".repeat(64) },
    oracle: { kind: "HarnessComponent", id: "oracle.effect_ledger", version: "v1", digest: "sha256:" + "e".repeat(64) },
    reporter: { kind: "HarnessComponent", id: "reporter.evidence", version: "v1", digest: "sha256:" + "f".repeat(64) },
  };
}

function sampleHarnessVersion(seed?: string): HarnessVersion {
  return new HarnessVersion({
    id: "checkout.double_spend",
    version: "v1",
    verificationContractRef: VCONTRACT,
    componentLock: componentLock(),
    evidenceSchema: "tekmovela.evidence-bundle.v1.schema.json",
    ...(seed !== undefined ? { seed } : {}),
  });
}

function stubRuntime(): RuntimeAdapter {
  return {
    async startRun(req) {
      return { runRef: `${req.scenarioRef}/run-${req.seed}` };
    },
    async fetchEffects(handle) {
      return [{ action: "debit", argsDigest: "sha256:" + "a".repeat(64), resource: "ledger", stateDelta: "-1" }];
    },
    async cancelRun() {},
    async checkpointRun(handle) {
      return { checkpointRef: `ckp/${handle.runRef}`, evidenceRef: `evidence/${handle.runRef}`, digest: "sha256:" + "b".repeat(64) };
    },
    async status() {
      return "passed";
    },
    async probe() {
      return { effects: [], evidenceRef: "evidence/probe" };
    },
  };
}

test("harness version publishes a digest", () => {
  const hv = sampleHarnessVersion();
  hv.publish();
  assert.match(hv.digest ?? "", /^sha256:[0-9a-f]{64}$/);
});

test("harness version digest matches the shared cross-SDK golden digest", () => {
  const hv = sampleHarnessVersion();
  hv.publish();
  assert.equal(hv.digest, GOLDEN_HARNESS_VERSION_DIGEST);
});

test("harness version body emits the verification_contract_ref reference object", () => {
  const hv = sampleHarnessVersion();
  assert.deepEqual(hv.body().verification_contract_ref, {
    kind: "VerificationContract",
    id: "checkout.atomicity",
    version: "v1",
    digest: D,
  });
});

test("harness version body emits schema-required isolation_profile", () => {
  const hv = sampleHarnessVersion();
  assert.deepEqual(hv.body().isolation_profile, {});
});

test("harness version body omits environment_contract when seed is unset", () => {
  const hv = sampleHarnessVersion();
  assert.equal("environment_contract" in hv.body(), false);
});

test("harness version body emits environment_contract.seed when set", () => {
  const hv = sampleHarnessVersion("s1");
  assert.deepEqual(hv.body().environment_contract, { seed: "s1" });
});

test("harness version rejects a malformed verification_contract_ref", () => {
  assert.throws(
    () =>
      new HarnessVersion({
        id: "checkout.double_spend",
        version: "v1",
        verificationContractRef: { kind: "VerificationContract", id: "c", version: "v1", digest: "nope" },
        componentLock: componentLock(),
        evidenceSchema: "evidence.bundle@v1",
      }),
    ContractError
  );
});

test("harness version rejects a component lock entry without kind", () => {
  const lock = componentLock();
  delete lock.runner.kind;
  assert.throws(
    () =>
      new HarnessVersion({
        id: "checkout.double_spend",
        version: "v1",
        verificationContractRef: VCONTRACT,
        componentLock: lock,
        evidenceSchema: "evidence.bundle@v1",
      }),
    ContractError
  );
});

test("harness version rejects a component lock entry with a bad digest", () => {
  const lock = componentLock();
  lock.adapter.digest = "not-a-digest";
  assert.throws(
    () =>
      new HarnessVersion({
        id: "checkout.double_spend",
        version: "v1",
        verificationContractRef: VCONTRACT,
        componentLock: lock,
        evidenceSchema: "evidence.bundle@v1",
      }),
    ContractError
  );
});

test("harness version rejects an unsupported determinism mode", () => {
  assert.throws(
    () =>
      new HarnessVersion({
        id: "checkout.double_spend",
        version: "v1",
        verificationContractRef: VCONTRACT,
        componentLock: componentLock(),
        evidenceSchema: "evidence.bundle@v1",
        determinismMode: "random" as "statistical",
      }),
    ContractError
  );
});

test("local runner executes a scenario and seals a bundle", async () => {
  const runner = new LocalRunner(stubRuntime());
  const scenario = {
    id: "scenario.double_spend",
    version: "v1",
    verificationContractRef: "checkout.atomicity",
    seed: "s1",
    harnessVersionRef: "local@v2",
  };
  const result = await runner.execute(scenario, "cap/v2", "run.e0001");
  assert.ok(result.bundle.bundleDigest);
  result.bundle.verify();
  assert.ok(result.bundle.items.length > 0);
  assert.equal(result.runId, result.bundle.runRef);
  assert.equal(result.bundle.harnessVersionRef, scenario.harnessVersionRef);
  assert.equal(result.bundle.items[0].kind, ItemKind.Effect);
});

test("local runner defaults harness version to local@v1", async () => {
  const runner = new LocalRunner(stubRuntime());
  const result = await runner.execute(
    { id: "scenario.double_spend", version: "v1", verificationContractRef: "checkout.atomicity", seed: "s1" },
    "cap/v2",
    "run.e0002"
  );
  assert.equal(result.bundle.harnessVersionRef, "local@v1");
});
