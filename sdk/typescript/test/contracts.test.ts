import { test } from "node:test";
import assert from "node:assert/strict";
import { parseDigest, digestOf, canonicalJSON, Reference, API_VERSION, digestFromBytes } from "../src/contracts.ts";

test("API_VERSION is tekmovela.io/v1", () => {
  assert.equal(API_VERSION, "tekmovela.io/v1");
});

test("parseDigest accepts sha256:<64 hex>", () => {
  assert.equal(parseDigest("sha256:" + "a".repeat(64)), "sha256:" + "a".repeat(64));
});

test("parseDigest rejects malformed", () => {
  assert.throws(() => parseDigest("md5:abc"));
  assert.throws(() => parseDigest("sha256:xyz"));
  assert.throws(() => parseDigest(""));
});

test("digestOf is stable across key order", () => {
  const a = digestOf({ a: 1, b: 2 });
  const b = digestOf({ b: 2, a: 1 });
  assert.equal(a, b);
});

test("digestOf matches the shared golden digest", () => {
  assert.equal(digestOf({ a: 1, b: 2 }), "sha256:43258cff783fe7036d8a43033f830adfc60ec037382473548ac742b888292777");
});

test("digestOf is sha256 prefixed", () => {
  assert.match(digestOf({ id: "x" }), /^sha256:[0-9a-f]{64}$/);
});

test("canonicalJSON produces compact deterministic JSON", () => {
  const out = canonicalJSON({ b: 2, a: 1 });
  assert.equal(out, '{"a":1,"b":2}');
});

test("Reference validates kind/id/version/digest", () => {
  const ref = new Reference("EvidenceBundle", "b.1", "v1", "sha256:" + "b".repeat(64));
  assert.equal(ref.kind, "EvidenceBundle");
  assert.throws(() => new Reference("", "b.1", "v1", "sha256:" + "b".repeat(64)));
  assert.throws(() => new Reference("EvidenceBundle", "b.1", "v1", "nope"));
});

test("digestFromBytes accepts Uint8Array and matches a Buffer input", () => {
  const bytes = new Uint8Array([1, 2, 3, 4]);
  assert.match(digestFromBytes(bytes), /^sha256:[0-9a-f]{64}$/);
  assert.equal(digestFromBytes(bytes), digestFromBytes(Buffer.from([1, 2, 3, 4])));
});
