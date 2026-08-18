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

export interface ComponentRef {
  id: string;
  version: string;
  digest: string;
}

export interface HarnessVersionInit {
  id: string;
  version: string;
  verificationContractRef: string;
  componentLock: Record<ComponentKind, ComponentRef>;
  evidenceSchema: string;
  determinismMode?: string;
  seed?: string;
}

export class HarnessVersion {
  id: string;
  version: string;
  verificationContractRef: string;
  componentLock: Record<ComponentKind, ComponentRef>;
  evidenceSchema: string;
  determinismMode: string;
  seed?: string;
  digest?: string;

  constructor(init: HarnessVersionInit) {
    this.id = init.id;
    this.version = init.version;
    this.verificationContractRef = init.verificationContractRef;
    this.componentLock = init.componentLock;
    this.evidenceSchema = init.evidenceSchema;
    this.determinismMode = init.determinismMode ?? "deterministic";
    this.seed = init.seed;
    for (const kind of COMPONENT_KINDS) {
      const ref = this.componentLock[kind];
      if (!ref) {
        throw new ContractError(`component_lock.${kind} is required`);
      }
      parseDigest(ref.digest);
    }
  }

  body(): Record<string, unknown> {
    return {
      api_version: API_VERSION,
      kind: "HarnessVersion",
      id: this.id,
      version: this.version,
      verification_contract_ref: this.verificationContractRef,
      component_lock: this.componentLock,
      evidence_schema: this.evidenceSchema,
      determinism_profile: { mode: this.determinismMode },
    };
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
    const handle = await this.runtime.startRun({
      scenarioRef: scenario.id,
      harnessVersionRef: scenario.verificationContractRef,
      sutRef,
      seed: scenario.seed,
    });
    const effects = await this.runtime.fetchEffects(handle);
    const bundle = new EvidenceBundle({
      id: `bundle.${runId}`,
      runRef: runId,
      verificationContractRef: scenario.verificationContractRef,
      harnessVersionRef: scenario.verificationContractRef,
    });
    for (const e of effects) {
      bundle.addItem({ kind: ItemKind.Effect, uri: `${e.action}/${e.resource}`, digest: e.argsDigest });
    }
    bundle.seal();
    return { runId: handle.runRef, bundle, effects };
  }
}
