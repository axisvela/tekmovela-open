import { test } from "node:test";
import assert from "node:assert/strict";
import { EvidenceBundle, ItemKind } from "../src/evidence.ts";

test("bundle seals and verifies", () => {
  const b = new EvidenceBundle({
    id: "bundle.1",
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
    id: "bundle.2",
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
    id: "bundle.3",
    runRef: "run.3",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  assert.throws(() => b.seal());
});

test("missing evidence is evidence", () => {
  const b = new EvidenceBundle({
    id: "bundle.4",
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
    id: "bundle.5",
    runRef: "run.5",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  assert.throws(() => b.addItem({ kind: ItemKind.Trace, uri: "t", digest: "bad" }));
});

test("body emits environment_digest null when unset", () => {
  const b = new EvidenceBundle({
    id: "bundle.6",
    runRef: "run.6",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  assert.equal(b.body().environment_digest, null);
});

test("bundle digest matches Python reference when environmentDigest unset", () => {
  const b = new EvidenceBundle({
    id: "bundle.1",
    runRef: "run.1",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: "sha256:" + "a".repeat(64) });
  assert.equal(b.digest(), "sha256:8012d279ba54e66cae7f2f37d1ba0f308119d1c3396ff1b82c0e38df57740ef3");
});

test("bundle digest matches Python reference when environmentDigest set", () => {
  const b = new EvidenceBundle({
    id: "bundle.1",
    runRef: "run.1",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
    environmentDigest: "sha256:" + "d".repeat(64),
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: "sha256:" + "a".repeat(64) });
  assert.equal(b.digest(), "sha256:8d327fdb984043dd94cd9bc1b3bab92607193ee1af1a0e2703f447a2a1ac0d8a");
});

test("body is stable across seal", () => {
  const b = new EvidenceBundle({
    id: "bundle.7",
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
    id: "bundle.8",
    runRef: "run.8",
    verificationContractRef: "checkout.atomicity@v1",
    harnessVersionRef: "checkout.double_spend@v1",
  });
  b.addItem({ kind: ItemKind.Trace, uri: "trace.1", digest: "sha256:" + "a".repeat(64) });
  b.seal();
  const wire = JSON.parse(JSON.stringify(b));
  assert.equal(wire.bundle_digest, b.bundleDigest);
  assert.equal(wire.environment_digest, null);
  assert.equal(wire.kind, "EvidenceBundle");
  assert.equal("bundle_digest" in b.body(), false);
});
