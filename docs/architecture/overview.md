# TEKMOVELA Open Architecture Overview

TEKMOVELA is the AxisRobo **Autonomous-System Engineering Assurance Control Plane**. This is the public, high-level architecture for integrators; the authoritative architecture, product plan, and internal design live in the private core and EE repositories.

## Main data chain

```text
Intent -> VerificationContract -> HarnessVersion -> Scenario/TestRun
  -> Trace/FailureEvidence -> EvaluationResult -> AssuranceClaim
  -> GateDecision/Attestation -> RuntimeSignal -> ReverificationRequest
```

## Eight planes

1. **Assurance Specification** — Intent Bridge, Requirement Trace, Acceptance Criteria (IRIS)
2. **Verification Substrate** — Verification Contract, Harness Registry, Component Catalog, Coverage Debt (VESTA)
3. **Agentic Testing** — Scenario, TestControlLoop, Run Coordinator, OTP profile (ATLAS)
4. **Telemetry & Evidence** — Trace Ingest, PPES Adapter, Evidence Store (LENS + AxisRobo runtimes)
5. **Evaluation & Diagnosis** — Oracle/Evaluator Registry, StructuredFailureEvidence, attribution (ATLAS + LENS)
6. **Assurance & Release** — Claims, Coverage, Gate Engine, Waiver, Attestation, Reverification (Core)
7. **Engineering Collaboration** — Role/Contract references, contributor gate, PM sync (MACD)
8. **Operations** — Behavioral bounds, anomaly triage, drift, SLO (LENS)

## Storage layering

Control metadata (relational), harness/contract artifacts (object store + metadata), raw telemetry (columnar/trace store), and assurance evidence (append-only, digest-verifiable) are stored in separate layers. A trace database never doubles as the evidence ledger.

## Integration boundaries

TEKMOVELA verifies but does not execute. It consumes Intent/AcceptanceCriteria and runtime signals and produces verification evidence, GateDecision, ConformanceAttestation, and ReverificationRequest. The full product-by-product integration boundary is documented in the private core repository.
