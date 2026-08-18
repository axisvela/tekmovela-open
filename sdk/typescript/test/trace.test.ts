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
