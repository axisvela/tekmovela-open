import { ContractError, parseDigest, API_VERSION } from "./contracts.ts";

export const TracePerspective = {
  Execution: "execution",
  Ppes: "ppes",
  Effect: "effect",
  Authority: "authority",
  WorldState: "world_state",
} as const;
export type TracePerspective = (typeof TracePerspective)[keyof typeof TracePerspective];

export interface TracePerspectiveRef {
  perspective_type: string;
  uri: string;
  digest: string;
}

export class TraceEnvelope {
  id: string;
  runRef: string;
  perspectives: TracePerspectiveRef[];
  tenantId?: string;
  traceId?: string;

  constructor(id: string, runRef: string, tenantId?: string, traceId?: string) {
    this.id = id;
    this.runRef = runRef;
    this.perspectives = [];
    this.tenantId = tenantId;
    this.traceId = traceId;
  }

  addPerspective(type: TracePerspective, uri: string, digest: string): this {
    parseDigest(digest);
    this.perspectives.push({ perspective_type: type, uri, digest });
    return this;
  }

  body(): Record<string, unknown> {
    const out: Record<string, unknown> = {
      api_version: API_VERSION,
      kind: "TraceEnvelope",
      id: this.id,
      run_ref: this.runRef,
      perspectives: this.perspectives,
    };
    if (this.tenantId) out.tenant_id = this.tenantId;
    if (this.traceId) out.trace_id = this.traceId;
    return out;
  }

  verify(): void {
    for (const p of this.perspectives) {
      parseDigest(p.digest);
    }
  }
}
