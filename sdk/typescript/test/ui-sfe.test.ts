import { test } from "node:test";
import assert from "node:assert/strict";
import { StaticUICapture, UiRepairFailureEvidence } from "../src/ui-sfe.ts";

test("static ui capture returns digest-bound refs", async () => {
  const cap = new StaticUICapture({
    screenshot: Buffer.from("png-bytes"),
    dom: "data:html",
    alert: "something went wrong",
  });
  const shot = await cap.captureScreenshot();
  assert.match(shot.digest, /^sha256:[0-9a-f]{64}$/);
  const dom = await cap.captureDOM();
  assert.equal(dom.uri, "static://dom");
});

test("ui repair failure evidence binds all three refs", async () => {
  const cap = new StaticUICapture({
    screenshot: Buffer.from("s"),
    dom: "<html>",
    alert: "failure",
  });
  const fe = await UiRepairFailureEvidence.capture(cap, "failure.ui.e0001", "local://trace/run.jsonl");
  assert.equal(fe.body().profile, "ui-repair/v1");
  assert.equal(fe.body().symptom.type, "alert");
  assert.ok(fe.body().screenshot_ref);
  assert.ok(fe.body().dom_ref);
  assert.ok(fe.body().alert_ref);
  assert.match(fe.digest(), /^sha256:[0-9a-f]{64}$/);
});
