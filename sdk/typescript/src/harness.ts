import { ContractError, digestOf, parseDigest, API_VERSION } from "./contracts.ts";
import { EvidenceBundle, ItemKind } from "./evidence.ts";

export type RunStatus = "pending" | "running" | "interrupted" | "passed" | "failed" | "inconclusive";

export interface RunRequest {
  scenarioRef: string;
  harnessVersionRef: string;
  sutRef: string;
  seed: string;
}

export interface RunHandle {
  runRef: string;
}

export interface Effect {
  action: string;
  argsDigest: string;
  resource: string;
  stateDelta: string;
  externalEffect?: boolean;
}

export interface CheckpointResult {
  checkpointRef: string;
  evidenceRef: string;
  digest: string;
}

export interface ProbeRequest {
  runRef: string;
  controlLoopRef: string;
  seed: string;
  observationBound: number;
}

export interface ProbeResult {
  effects: Effect[];
  evidenceRef: string;
}

export interface RuntimeAdapter {
  startRun(req: RunRequest): Promise<RunHandle>;
  fetchEffects(handle: RunHandle): Promise<Effect[]>;
  cancelRun(handle: RunHandle): Promise<void>;
  checkpointRun(handle: RunHandle): Promise<CheckpointResult>;
  status(handle: RunHandle): Promise<RunStatus>;
  probe(req: ProbeRequest): Promise<ProbeResult>;
}

export const COMPONENT_KINDS = ["runner", "adapter", "oracle", "reporter"] as const;
export type ComponentKind = (typeof COMPONENT_KINDS)[number];

export const DETERMINISM_MODES = ["deterministic", "statistical", "hybrid"] as const;
export type DeterminismMode = (typeof DETERMINISM_MODES)[number];

export interface ComponentRef {
  kind: string;
  id: string;
  version: string;
  digest: string;
}

export interface ReferenceLike {
  kind: string;
  id: string;
  version: string;
  digest: string;
}

export interface HarnessVersionInit {
  id: string;
  version: string;
  verificationContractRef: ReferenceLike;
  componentLock: Record<ComponentKind, ComponentRef>;
  evidenceSchema: string;
  determinismMode?: DeterminismMode;
  seed?: string;
}

export class HarnessVersion {
  id: string;
  version: string;
  verificationContractRef: ReferenceLike;
  componentLock: Record<ComponentKind, ComponentRef>;
  evidenceSchema: string;
  determinismMode: DeterminismMode;
  seed?: string;
  digest?: string;

  constructor(init: HarnessVersionInit) {
    this.id = init.id;
    this.version = init.version;
    this.verificationContractRef = { ...init.verificationContractRef };
    if (!this.verificationContractRef.kind || !this.verificationContractRef.id || !this.verificationContractRef.version) {
      throw new ContractError("verification_contract_ref kind, id, and version are required");
    }
    parseDigest(this.verificationContractRef.digest);
    this.componentLock = init.componentLock;
    this.evidenceSchema = init.evidenceSchema;
    this.determinismMode = init.determinismMode ?? "deterministic";
    this.seed = init.seed;
    if (!DETERMINISM_MODES.includes(this.determinismMode)) {
      throw new ContractError(`determinism_mode ${this.determinismMode} is not supported`);
    }
    for (const kind of COMPONENT_KINDS) {
      const ref = this.componentLock[kind];
      if (!ref) {
        throw new ContractError(`component_lock.${kind} is required`);
      }
      if (!ref.kind) {
        throw new ContractError(`component_lock.${kind}.kind is required`);
      }
      parseDigest(ref.digest);
    }
  }

  body(): Record<string, unknown> {
    const body: Record<string, unknown> = {
      api_version: API_VERSION,
      kind: "HarnessVersion",
      id: this.id,
      version: this.version,
      verification_contract_ref: this.verificationContractRef,
      component_lock: this.componentLock,
      evidence_schema: this.evidenceSchema,
      determinism_profile: { mode: this.determinismMode },
      isolation_profile: {},
    };
    if (this.seed !== undefined) {
      body.environment_contract = { seed: this.seed };
    }
    return body;
  }

  canonicalDigest(): string {
    return digestOf(this.body());
  }

  publish(): this {
    this.digest = this.canonicalDigest();
    return this;
  }
}

export interface Scenario {
  id: string;
  version: string;
  verificationContractRef: string;
  seed: string;
  harnessVersionRef?: string;
}

export interface RunResult {
  runId: string;
  bundle: EvidenceBundle;
  effects: Effect[];
}

export class LocalRunner {
  runtime: RuntimeAdapter;

  constructor(runtime: RuntimeAdapter) {
    this.runtime = runtime;
  }

  async execute(scenario: Scenario, sutRef: string, runId: string): Promise<RunResult> {
    const harnessVersionRef = scenario.harnessVersionRef ?? "local@v1";
    const handle = await this.runtime.startRun({
      scenarioRef: scenario.id,
      harnessVersionRef,
      sutRef,
      seed: scenario.seed,
    });
    const effects = await this.runtime.fetchEffects(handle);
    const bundle = new EvidenceBundle({
      id: `bundle.${runId}`,
      runRef: handle.runRef,
      verificationContractRef: scenario.verificationContractRef,
      harnessVersionRef,
    });
    for (const e of effects) {
      bundle.addItem({ kind: ItemKind.Effect, uri: `${e.action}/${e.resource}`, digest: e.argsDigest });
    }
    bundle.seal();
    return { runId: handle.runRef, bundle, effects };
  }
}
