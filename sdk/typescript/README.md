# TEKMOVELA TypeScript SDK

Dependency-free TypeScript SDK for the TEKMOVELA engineering assurance control
plane: Node agent harness, UI structured-failure-evidence capture, and trace
emission. Apache-2.0, version-locked to the `tekmovela` core release.

## Modules

- `contracts` — canonical-JSON digests, `parseDigest`, validated `Reference`.
- `evidence` — digest-sealed `EvidenceBundle`; missing evidence is evidence.
- `harness` — `HarnessVersion` authoring + a deterministic `LocalRunner` over a `RuntimeAdapter`.
- `ui-sfe` — `ui-repair/v1` failure evidence; `UICapture` is a caller-supplied seam.
- `trace` — `TraceEnvelope` perspective emission.

## Usage

```ts
import { LocalRunner, Reference } from "@axisrobo/tekmovela";

const runner = new LocalRunner(runtime);
const result = await runner.execute(
  { id: "scenario.double_spend", version: "v1", verificationContractRef: "checkout.atomicity", seed: "s1" },
  "cap/v2",
  "run.e0001"
);
result.bundle.verify();

const version = new HarnessVersion({
  id: "checkout.double_spend",
  version: "v1",
  verificationContractRef: new Reference("VerificationContract", "checkout.atomicity", "v1", "sha256:" + "a".repeat(64)),
  componentLock: {
    runner: { kind: "HarnessComponent", id: "runner.deterministic", version: "v1", digest: "sha256:" + "c".repeat(64) },
    adapter: { kind: "HarnessComponent", id: "adapter.praxovela", version: "v1", digest: "sha256:" + "d".repeat(64) },
    oracle: { kind: "HarnessComponent", id: "oracle.effect_ledger", version: "v1", digest: "sha256:" + "e".repeat(64) },
    reporter: { kind: "HarnessComponent", id: "reporter.evidence", version: "v1", digest: "sha256:" + "f".repeat(64) },
  },
  evidenceSchema: "tekmovela.evidence-bundle.v1.schema.json",
});
version.publish();
```

## Build and test

```bash
npm install
npm test
npm run build
```

The published `dist/` runs on Node >= 18. Running `npm test` against the
TypeScript sources needs Node >= 23.6 (native type-stripping). Version `0.17.0`, in lockstep with the `tekmovela` core release.
