# TEKMOVELA Go SDK

Go SDK for TEKMOVELA consumers and AxisRobo runtime integration. Provides
canonical-JSON digests, digest-sealed EvidenceBundle, and harness authoring
with a local runner. Digests match the Python and TypeScript SDKs so evidence
sealed in any SDK verifies in the others.

## Modules

- `contracts` — canonical-JSON digests, `ParseDigest`, `Reference`.
- `evidence` — digest-sealed `EvidenceBundle`; missing evidence is evidence.
- `harness` — `HarnessVersion` authoring + a deterministic `LocalRunner` over a `RuntimeAdapter`.
- `reference` — low-level digest/reference rules.

## Usage

```go
import "github.com/axisrobo/tekmovela-open/sdk/go/harness"

runner := harness.NewLocalRunner(rt)
result, err := runner.Execute(harness.Scenario{
    ID: "scenario.double_spend", Version: "v1",
    VerificationContractRef: "checkout.atomicity", Seed: "s1",
}, "cap/v2", "run.e0001")
```

## Test

```bash
go test ./...
```

Version `0.15.0`, in lockstep with the `tekmovela` core release.
