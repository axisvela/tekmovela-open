# TEKMOVELA Open Python SDK

Harness authoring, fixtures, oracles, a local deterministic runner, and an
evidence emitter for the TEKMOVELA engineering assurance control plane
(Apache-2.0). This SDK is the public M1 deliverable: it mirrors the core's
contract shapes so Python harness authors can build against the same frozen
schemas.

## Modules

- `tekmovela.contracts` — digest helpers, validated references, and a minimal
  structural schema validator (the authoritative conformance lives in
  `contracts/`).
- `tekmovela.evidence` — digest-sealed EvidenceBundle and Items; missing
  evidence is itself evidence.
- `tekmovela.harness` — HarnessVersion authoring with a frozen component lock.
- `tekmovela.runner` — a deterministic local runner that executes a Scenario
  against a `RuntimeAdapter` and emits a sealed EvidenceBundle.

## Usage

```python
from tekmovela.runner import LocalRunner, Scenario

result = LocalRunner(runtime).execute(
    Scenario(id="scenario.double_spend", version="v1",
             verification_contract_ref="checkout.atomicity", seed="s1"),
    sut_ref="cap/v2",
    run_id="run.001",
)
result.bundle.verify()
```

## Tests

```bash
PYTHONPATH=sdk/python python -m unittest discover -s sdk/python/tests -p "test_*.py"
```

Version `0.15.0`, in lockstep with the `tekmovela` core release.
