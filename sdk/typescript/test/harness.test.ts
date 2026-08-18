import { test } from "node:test";
import assert from "node:assert/strict";
import { HarnessVersion, LocalRunner } from "../src/harness.ts";
import type { RuntimeAdapter } from "../src/harness.ts";
import { ItemKind } from "../src/evidence.ts";

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
  const hv = new HarnessVersion({
    id: "checkout.double_spend",
    version: "v1",
    verificationContractRef: "checkout.atomicity@v1",
    componentLock: {
      runner: { id: "r.1", version: "v1", digest: "sha256:" + "c".repeat(64) },
      adapter: { id: "a.1", version: "v1", digest: "sha256:" + "d".repeat(64) },
      oracle: { id: "o.1", version: "v1", digest: "sha256:" + "e".repeat(64) },
      reporter: { id: "p.1", version: "v1", digest: "sha256:" + "f".repeat(64) },
    },
    evidenceSchema: "evidence.bundle@v1",
  });
  hv.publish();
  assert.match(hv.digest ?? "", /^sha256:[0-9a-f]{64}$/);
});

test("local runner executes a scenario and seals a bundle", async () => {
  const runner = new LocalRunner(stubRuntime());
  const result = await runner.execute(
    { id: "scenario.double_spend", version: "v1", verificationContractRef: "checkout.atomicity", seed: "s1" },
    "cap/v2",
    "run.e0001"
  );
  assert.ok(result.bundle.bundleDigest);
  result.bundle.verify();
  assert.ok(result.bundle.items.length > 0);
});
