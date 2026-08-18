import { ContractError, digestOf, digestFromBytes, API_VERSION } from "./contracts.ts";

export interface ArtifactRef {
  uri: string;
  digest: string;
}

export interface UICapture {
  captureScreenshot(): Promise<ArtifactRef>;
  captureDOM(): Promise<ArtifactRef>;
  captureAlert(message: string): Promise<ArtifactRef>;
}

export interface StaticUICaptureInit {
  screenshot: Buffer;
  dom: string;
  alert: string;
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

  async captureScreenshot(): Promise<ArtifactRef> {
    return { uri: "static://screenshot", digest: digestFromBytes(this.init.screenshot) };
  }

  async captureDOM(): Promise<ArtifactRef> {
    return { uri: "static://dom", digest: digestOf(this.init.dom) };
  }

  async captureAlert(message: string): Promise<ArtifactRef> {
    return { uri: "static://alert", digest: digestOf(message) };
  }
}

export class UiRepairFailureEvidence {
  id: string;
  traceSliceRef: string;
  screenshotRef: ArtifactRef;
  domRef: ArtifactRef;
  alertRef: ArtifactRef;

  constructor(
    id: string,
    traceSliceRef: string,
    screenshotRef: ArtifactRef,
    domRef: ArtifactRef,
    alertRef: ArtifactRef
  ) {
    this.id = id;
    this.traceSliceRef = traceSliceRef;
    this.screenshotRef = screenshotRef;
    this.domRef = domRef;
    this.alertRef = alertRef;
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

  body(): Record<string, unknown> {
    return {
      api_version: API_VERSION,
      kind: "FailureEvidence",
      id: this.id,
      profile: "ui-repair/v1",
      symptom: { type: "alert", message: "ui failure observed" },
      trace_slice_ref: this.traceSliceRef,
      screenshot_ref: this.screenshotRef,
      dom_ref: this.domRef,
      alert_ref: this.alertRef,
    };
  }

  digest(): string {
    return digestOf(this.body());
  }
}
