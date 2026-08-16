# TEKMOVELA Contracts

Versioned JSON Schema contracts and fixtures for the TEKMOVELA engineering assurance control plane. Published contract versions are immutable; behavioral changes create a new version.

## Layout

```
contracts/
├── intent/          IntentSpecRef, AcceptanceCriterionRef, ConstraintRef, RequirementTraceLink
├── verification/    VerificationContract
├── harness/         HarnessDefinition, HarnessVersion, HarnessComponent, CoverageDebt
├── testing/         Scenario, EvaluationSet, TestRun, TestControlLoop
├── evidence/        TraceEnvelope, EvidenceBundle, FailureEvidence, AssuranceEvidence
├── assurance/       AssuranceClaim, CoverageStatement, GateDecision, Attestation, ReleaseCandidate
└── runtime/         BehavioralBound, AnomalySignal, ReverificationRequest
```

Each schema directory has `fixtures/valid/` (must validate) and `fixtures/invalid/` (must fail) fixtures that serve as the conformance reference for any SDK or adapter.

## Validation

```bash
npm install
npm run test:contracts
```

The validator discovers schemas by `api_version` + `kind`, then asserts that every valid fixture passes and every invalid fixture fails.

## Reference Model

Every contract carries `api_version` (`tekmovela.io/v1alpha1`) and a fixed `kind`. Refs to published objects are `{kind, id, version, digest}` with `sha256:` digests. See [COMPATIBILITY_POLICY.md](COMPATIBILITY_POLICY.md) for immutability and compatibility rules.
