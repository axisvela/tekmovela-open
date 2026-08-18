import { ContractError, parseDigest, API_VERSION } from "./contracts.ts";

export const TracePerspective = {
  Execution: "execution",
  Ppes: "ppes",
  Effect: "effect",
  Authority: "authority",
  WorldState: "world_state",
} as const;
export type TracePerspective = (typeof TracePerspective)[keyof typeof TracePerspective];

const PERSPECTIVE_TYPES: readonly TracePerspective[] = Object.values(TracePerspective);

function validatePerspectiveRef(ref: TracePerspectiveRef): void {
  if (!PERSPECTIVE_TYPES.includes(ref.perspective_type)) {
    throw new ContractError(`perspective_type ${ref.perspective_type} is not supported`);
  }
  if (!ref.uri) {
    throw new ContractError("perspective uri is required");
  }
  parseDigest(ref.digest);
}

export interface TracePerspectiveRef {
  perspective_type: TracePerspective;
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
    const ref: TracePerspectiveRef = { perspective_type: type, uri, digest };
    validatePerspectiveRef(ref);
    this.perspectives.push(ref);
    return this;
  }

  body(): Record<string, unknown> {
    const out: Record<string, unknown> = {
      api_version: API_VERSION,
      kind: "TraceEnvelope",
      id: this.id,
      run_ref: this.runRef,
      perspectives: this.perspectives.map((p) => ({ ...p })),
    };
    if (this.tenantId) out.tenant_id = this.tenantId;
    if (this.traceId) out.trace_id = this.traceId;
    return out;
  }

  verify(): void {
    if (this.perspectives.length === 0) {
      throw new ContractError(`envelope ${this.id} has no perspectives`);
    }
    for (const p of this.perspectives) {
      validatePerspectiveRef(p);
    }
  }
}
