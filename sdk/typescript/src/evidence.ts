import { ContractError, digestOf, parseDigest, API_VERSION } from "./contracts.ts";

export const ItemKind = {
  Trace: "trace",
  Effect: "effect",
  Failure: "failure",
  Attribution: "attribution",
  Explanation: "explanation",
  Authority: "authority",
  Missing: "missing",
} as const;

export type ItemKind = (typeof ItemKind)[keyof typeof ItemKind];

export interface Item {
  kind: ItemKind;
  uri: string;
  digest: string;
  note?: string;
}

export interface EvidenceBundleInit {
  id: string;
  runRef: string;
  verificationContractRef: string;
  harnessVersionRef: string;
  items?: Item[];
  scope?: Record<string, string>;
  environmentDigest?: string;
}

export class EvidenceBundle {
  id: string;
  runRef: string;
  verificationContractRef: string;
  harnessVersionRef: string;
  items: Item[];
  scope: Record<string, string>;
  environmentDigest?: string;
  bundleDigest?: string;

  constructor(init: EvidenceBundleInit) {
    this.id = init.id;
    this.runRef = init.runRef;
    this.verificationContractRef = init.verificationContractRef;
    this.harnessVersionRef = init.harnessVersionRef;
    this.items = init.items ?? [];
    this.scope = init.scope ?? {};
    this.environmentDigest = init.environmentDigest;
  }

  addItem(item: Item): this {
    if (!item.kind || !item.uri) {
      throw new ContractError("item kind and uri are required");
    }
    parseDigest(item.digest);
    this.items.push(item);
    return this;
  }

  addMissing(uri: string, digest: string, note: string): this {
    return this.addItem({ kind: ItemKind.Missing, uri, digest, note });
  }

  body(): Record<string, unknown> {
    const body: Record<string, unknown> = {
      api_version: API_VERSION,
      kind: "EvidenceBundle",
      id: this.id,
      run_ref: this.runRef,
      verification_contract_ref: this.verificationContractRef,
      harness_version_ref: this.harnessVersionRef,
      scope: this.scope,
      items: this.items.map((i) => ({ kind: i.kind, uri: i.uri, digest: i.digest, ...(i.note ? { note: i.note } : {}) })),
    };
    if (this.environmentDigest) body.environment_digest = this.environmentDigest;
    return body;
  }

  toJSON(): Record<string, unknown> {
    return { ...this.body(), bundle_digest: this.bundleDigest ?? null };
  }

  digest(): string {
    return digestOf(this.body());
  }

  seal(): this {
    if (this.items.length === 0) {
      throw new ContractError(`bundle ${this.id} cannot seal with no items`);
    }
    this.bundleDigest = this.digest();
    return this;
  }

  verify(): void {
    for (const item of this.items) {
      parseDigest(item.digest);
    }
    if (!this.bundleDigest) {
      throw new ContractError(`bundle ${this.id} is not sealed`);
    }
    if (this.digest() !== this.bundleDigest) {
      throw new ContractError(`bundle ${this.id} digest mismatch`);
    }
  }

  missing(): Item[] {
    return this.items.filter((i) => i.kind === ItemKind.Missing);
  }
}
