import { test } from "node:test";
import assert from "node:assert/strict";
import { ContractError } from "../src/contracts.ts";
import { EvidenceBundle, ItemKind } from "../src/evidence.ts";

test("bundle seals and verifies", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0001",
    runRef: "run.1",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: "sha256:" + "a".repeat(64) });
  b.seal();
  assert.ok(b.bundleDigest);
  b.verify();
});

test("bundle rejects tampering", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0002",
    runRef: "run.2",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.2", digest: "sha256:" + "b".repeat(64) });
  b.seal();
  b.items[0].uri = "tampered";
  assert.throws(() => b.verify());
});

test("bundle cannot seal with no items", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0003",
    runRef: "run.3",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  assert.throws(() => b.seal());
});

test("missing evidence is evidence", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0004",
    runRef: "run.4",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addMissing("local://authority.jsonl", "sha256:" + "c".repeat(64), "authority perspective not captured");
  b.seal();
  assert.equal(b.missing().length, 1);
});

test("addItem validates digest", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0005",
    runRef: "run.5",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  assert.throws(() => b.addItem({ kind: ItemKind.Trace, uri: "t", digest: "bad" }));
});

test("body omits environment_digest when unset", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0006",
    runRef: "run.6",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  assert.equal("environment_digest" in b.body(), false);
});

test("body includes environment_digest when set", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0006",
    runRef: "run.6",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
    environmentDigest: "sha256:" + "d".repeat(64),
  });
  assert.equal(b.body().environment_digest, "sha256:" + "d".repeat(64));
});

test("bundle digest matches Python reference when environmentDigest unset", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0001",
    runRef: "run.1",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: "sha256:" + "a".repeat(64) });
  assert.equal(b.digest(), "sha256:18b4223c8528f03c945964fe0a6590dafcefc77c377034a46bfc3374b4aee470");
});

test("bundle digest matches Python reference when environmentDigest set", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0001",
    runRef: "run.1",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
    environmentDigest: "sha256:" + "d".repeat(64),
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: "sha256:" + "a".repeat(64) });
  assert.equal(b.digest(), "sha256:88e88a573fe31224487abe800f920abf41b31527ba7d5045b86255d8e3fad6e6");
});

test("body is stable across seal", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0007",
    runRef: "run.7",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: "sha256:" + "a".repeat(64) });
  const before = b.digest();
  b.seal();
  assert.equal(b.digest(), before);
  b.verify();
});

test("toJSON emits bundle_digest after seal and survives JSON round-trip", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0008",
    runRef: "run.8",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: "sha256:" + "a".repeat(64) });
  b.seal();
  const wire = JSON.parse(JSON.stringify(b));
  assert.equal(wire.bundle_digest, b.bundleDigest);
  assert.equal("environment_digest" in wire, false);
  assert.equal(wire.kind, "EvidenceBundle");
  assert.equal("bundle_digest" in b.body(), false);
});

test("bundle scope serializes camelCase keys to snake_case wire form", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0009",
    runRef: "run.9",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
    scope: { intentRef: "intent.checkout", capabilityRef: "cap.checkout", timeWindow: "2026-08-15/2026-08-16" },
  });
  assert.deepEqual(b.body().scope, {
    intent_ref: "intent.checkout",
    capability_ref: "cap.checkout",
    time_window: "2026-08-15/2026-08-16",
  });
});

test("bundle scope omits unset keys from the wire body", () => {
  const b = new EvidenceBundle({
    id: "bundle.e0010",
    runRef: "run.10",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
    scope: { capabilityRef: "cap.checkout" },
  });
  assert.deepEqual(b.body().scope, { capability_ref: "cap.checkout" });
});

test("bundle scope rejects unknown keys at runtime", () => {
  assert.throws(
    () =>
      new EvidenceBundle({
        id: "bundle.e0011",
        runRef: "run.11",
        verificationContractRef: "checkout.atomicity@v1",
        harnessVersionRef: "checkout.double_spend@v1",
        scope: { intentRef: "intent.checkout", bogus: "x" } as Record<string, string>,
      }),
    ContractError
  );
});
