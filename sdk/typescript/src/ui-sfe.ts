import { ContractError, digestOf, digestFromBytes, API_VERSION } from "./contracts.ts";

export const REDACTION_STATES = ["none", "partial", "full"] as const;
export type RedactionState = (typeof REDACTION_STATES)[number];

export interface UiArtifact {
  ref: string;
  digest: string;
  capturedAt: string;
  redaction: RedactionState;
}

export interface UICapture {
  captureScreenshot(): Promise<UiArtifact>;
  captureDOM(): Promise<UiArtifact>;
  captureAlert(message: string): Promise<UiArtifact>;
}

export interface StaticUICaptureInit {
  screenshot: Buffer;
  dom: string;
  alert: string;
}

function newArtifact(ref: string, digest: string): UiArtifact {
  return {
    ref,
    digest,
    capturedAt: new Date().toISOString(),
    redaction: "none",
  };
}

/**
 * Reference UICapture that binds caller-supplied values to digests. It never
 * performs browser execution — capture belongs to the runtime; TEKMOVELA
 * records references and digests only (ADR-0012).
 */
export class StaticUICapture implements UICapture {
  private init: StaticUICaptureInit;

  constructor(init: StaticUICaptureInit) {
    this.init = init;
  }

  async captureScreenshot(): Promise<UiArtifact> {
    return newArtifact("static://screenshot", digestFromBytes(this.init.screenshot));
  }

  async captureDOM(): Promise<UiArtifact> {
    return newArtifact("static://dom", digestOf(this.init.dom));
  }

  async captureAlert(message: string): Promise<UiArtifact> {
    return newArtifact("static://alert", digestOf(message));
  }
}

export class UiRepairFailureEvidence {
  id: string;
  traceSliceRef: string;
  screenshot: UiArtifact;
  dom: UiArtifact;
  alert: UiArtifact;

  constructor(
    id: string,
    traceSliceRef: string,
    screenshot: UiArtifact,
    dom: UiArtifact,
    alert: UiArtifact
  ) {
    this.id = id;
    this.traceSliceRef = traceSliceRef;
    this.screenshot = screenshot;
    this.dom = dom;
    this.alert = alert;
  }

  static async capture(cap: UICapture, id: string, traceSliceRef: string): Promise<UiRepairFailureEvidence> {
    if (!cap || !id || !traceSliceRef) {
      throw new ContractError("ui-repair capture requires a capture seam, id, and trace slice ref");
    }
    const screenshot = await cap.captureScreenshot();
    const dom = await cap.captureDOM();
    const alert = await cap.captureAlert("ui failure observed");
    return new UiRepairFailureEvidence(id, traceSliceRef, screenshot, dom, alert);
  }

  toArtifact(a: UiArtifact): Record<string, string> {
    return { ref: a.ref, digest: a.digest, captured_at: a.capturedAt, redaction: a.redaction };
  }

  body(): Record<string, unknown> {
    return {
      api_version: API_VERSION,
      kind: "FailureEvidence",
      id: this.id,
      profile: "ui-repair/v1",
      symptom: { type: "alert", message: "ui failure observed" },
      trace_slice_ref: this.traceSliceRef,
      ui_repair: {
        screenshot: this.toArtifact(this.screenshot),
        dom: this.toArtifact(this.dom),
        alert: this.toArtifact(this.alert),
      },
    };
  }

  digest(): string {
    return digestOf(this.body());
  }
}
