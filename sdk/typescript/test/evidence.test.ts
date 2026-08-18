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
