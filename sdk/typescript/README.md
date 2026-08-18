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
import { LocalRunner } from "@axisrobo/tekmovela";

const runner = new LocalRunner(runtime);
const result = await runner.execute(
  { id: "scenario.double_spend", version: "v1", verificationContractRef: "checkout.atomicity", seed: "s1" },
  "cap/v2",
  "run.e0001"
);
result.bundle.verify();
```

## Build and test

```bash
npm install
npm test
npm run build
```

Requires Node >= 23.6 (native TypeScript type-stripping). Version `0.15.0`, in lockstep with the `tekmovela` core release.
