import { test } from "node:test";
import assert from "node:assert/strict";
import { TraceEnvelope, TracePerspective } from "../src/trace.ts";

test("trace envelope emits and verifies perspectives", () => {
  const env = new TraceEnvelope("env.e0001", "run.e0001");
  env.addPerspective(TracePerspective.Execution, "local://trace/exec.jsonl", "sha256:" + "a".repeat(64));
  env.addPerspective(TracePerspective.Effect, "local://trace/effects.jsonl", "sha256:" + "b".repeat(64));
  env.verify();
  assert.equal(env.body().perspectives.length, 2);
});

test("trace envelope rejects bad perspective digest", () => {
  const env = new TraceEnvelope("env.e0002", "run.e0002");
  assert.throws(() => env.addPerspective(TracePerspective.Ppes, "t", "bad"));
});

test("trace envelope rejects a non-enum perspective type", () => {
  const env = new TraceEnvelope("env.e0003", "run.e0003");
  assert.throws(() =>
    env.addPerspective("not-a-perspective" as TracePerspective, "t", "sha256:" + "a".repeat(64))
  );
});

test("trace envelope rejects an empty uri", () => {
  const env = new TraceEnvelope("env.e0004", "run.e0004");
  assert.throws(() => env.addPerspective(TracePerspective.Execution, "", "sha256:" + "a".repeat(64)));
});

test("trace envelope verify rejects an empty perspective set", () => {
  const env = new TraceEnvelope("env.e0005", "run.e0005");
  assert.throws(() => env.verify());
});

test("trace envelope verify rejects a non-enum perspective type", () => {
  const env = new TraceEnvelope("env.e0006", "run.e0006");
  env.perspectives.push({ perspective_type: "nope" as TracePerspective, uri: "t", digest: "sha256:" + "a".repeat(64) });
  assert.throws(() => env.verify());
});

test("trace envelope verify rejects an empty uri", () => {
  const env = new TraceEnvelope("env.e0007", "run.e0007");
  env.perspectives.push({ perspective_type: TracePerspective.Authority, uri: "", digest: "sha256:" + "a".repeat(64) });
  assert.throws(() => env.verify());
});

test("trace envelope body omits tenant_id and trace_id when unset", () => {
  const env = new TraceEnvelope("env.e0008", "run.e0008");
  const body = env.body();
  assert.equal("tenant_id" in body, false);
  assert.equal("trace_id" in body, false);
});

test("trace envelope body carries tenant_id and trace_id when set", () => {
  const env = new TraceEnvelope("env.e0009", "run.e0009", "tenant.t1", "trace.t1");
  const body = env.body();
  assert.equal(body.tenant_id, "tenant.t1");
  assert.equal(body.trace_id, "trace.t1");
});

test("trace envelope body returns a fresh perspectives copy", () => {
  const env = new TraceEnvelope("env.e0010", "run.e0010");
  env.addPerspective(TracePerspective.Execution, "local://trace/exec.jsonl", "sha256:" + "a".repeat(64));
  const first = env.body();
  const second = env.body();
  assert.notEqual(first.perspectives, second.perspectives);
  assert.deepEqual(first.perspectives, second.perspectives);
});
