import { createHash } from "node:crypto";

export const API_VERSION = "tekmovela.io/v1alpha1";

const SHA256 = /^sha256:[0-9a-f]{64}$/;

export class ContractError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ContractError";
  }
}

export function parseDigest(digest: string): string {
  if (!SHA256.test(digest)) {
    throw new ContractError(`malformed digest ${digest}: want sha256:<64 hex>`);
  }
  return digest;
}

export function canonicalJSON(value: unknown): string {
  const sorted = sortKeys(value);
  return JSON.stringify(sorted);
}

function sortKeys(value: unknown): unknown {
  if (Array.isArray(value)) {
    return value.map(sortKeys);
  }
  if (value && typeof value === "object") {
    const out: Record<string, unknown> = {};
    for (const key of Object.keys(value).sort()) {
      out[key] = sortKeys((value as Record<string, unknown>)[key]);
    }
    return out;
  }
  return value;
}

export function digestOf(value: unknown): string {
  return "sha256:" + createHash("sha256").update(canonicalJSON(value)).digest("hex");
}

export function digestFromBytes(data: Buffer): string {
  return "sha256:" + createHash("sha256").update(data).digest("hex");
}

export class Reference {
  kind: string;
  id: string;
  version: string;
  digest: string;

  constructor(kind: string, id: string, version: string, digest: string) {
    if (!kind || !id || !version) {
      throw new ContractError("reference kind, id, and version are required");
    }
    parseDigest(digest);
    this.kind = kind;
    this.id = id;
    this.version = version;
    this.digest = digest;
  }

  toJSON(): { kind: string; id: string; version: string; digest: string } {
    return { kind: this.kind, id: this.id, version: this.version, digest: this.digest };
  }
}
