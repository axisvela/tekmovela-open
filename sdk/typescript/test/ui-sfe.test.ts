import { test } from "node:test";
import assert from "node:assert/strict";
import { digestOf } from "../src/contracts.ts";
import { StaticUICapture, UiRepairFailureEvidence } from "../src/ui-sfe.ts";

test("static ui capture returns schema-shaped artifacts", async () => {
  const cap = new StaticUICapture({
    screenshot: new Uint8Array([112, 110, 103, 45, 98, 121, 116, 101, 115]),
    dom: "data:html",
    alert: "something went wrong",
  });
  const shot = await cap.captureScreenshot();
  assert.equal(shot.ref, "static://screenshot");
  assert.match(shot.digest, /^sha256:[0-9a-f]{64}$/);
  assert.equal(typeof shot.capturedAt, "string");
  assert.equal(shot.redaction, "none");
  const dom = await cap.captureDOM();
  assert.equal(dom.ref, "static://dom");
  assert.equal(dom.redaction, "none");
});

test("ui repair failure evidence emits the ui_repair schema shape", async () => {
  const cap = new StaticUICapture({
    screenshot: new Uint8Array([115]),
    dom: "<html>",
    alert: "failure",
  });
  const fe = await UiRepairFailureEvidence.capture(cap, "failure.ui.e0001", "local://trace/run.jsonl");
  const body = fe.body();
  assert.equal(body.profile, "ui-repair/v1");
  assert.equal(body.symptom.type, "alert");
  const repair = body.ui_repair as Record<string, Record<string, string>>;
  for (const name of ["screenshot", "dom", "alert"]) {
    const artifact = repair[name];
    assert.ok(artifact.ref);
    assert.match(artifact.digest, /^sha256:[0-9a-f]{64}$/);
    assert.equal(typeof artifact.captured_at, "string");
    assert.ok(["none", "partial", "full"].includes(artifact.redaction));
  }
  assert.match(fe.digest(), /^sha256:[0-9a-f]{64}$/);
  assert.equal(fe.digest(), digestOf(body));
});
