# TEKMOVELA Contracts

Versioned JSON Schema contracts and fixtures for the TEKMOVELA engineering assurance control plane. Published contract versions are immutable; behavioral changes create a new version.

This directory is a **mirror** of the authoritative contracts in the core `tekmovela` repository, pinned to released version **v0.11.0**. Refresh it by copying from a released core version; never edit a published schema in place.

## Layout

```
contracts/
├── intent/          IntentSpecRef, AcceptanceCriterion, ConstraintRef, RequirementTraceLink
├── verification/    VerificationContract, Fixture
├── harness/         HarnessDefinition, HarnessVersion, HarnessComponent, HarnessComposition, OracleProfile, CoverageDebt
├── testing/         Scenario, EvaluationSet, EvaluationResult, TestRun, TestControlLoop, TestBudget, OwnershipTransferRecord, CIResult
├── evidence/        TraceEnvelope, EvidenceBundle, FailureEvidence, AssuranceEvidence, CausalAttribution, DecisionExplanation, CognitiveTrace, AccountabilityRecord, FailureCluster
├── assurance/       AssuranceClaim, CoverageStatement, GateDecision, GatePolicy, Attestation, Waiver, ReleaseCandidate, AssuranceCase
├── collaboration/   ContributorEvidence, ContributorGateDecision, DevelopmentRoleProfileRef, DevelopmentContractRef, EngineeringStateLink
├── improvement/     RegressionCandidate, ImprovementAction, IncidentRef
└── runtime/         BehavioralBound, AnomalySignal, ReverificationRequest, AccountabilityRecordRef
```

Each schema directory has `fixtures/valid/` (must validate) and `fixtures/invalid/` (must fail) fixtures that serve as the conformance reference for any SDK or adapter.

## Validation

```bash
npm install
npm run test:ci
```

The validator discovers schemas by `api_version` + `kind`, asserts that every valid fixture passes and every invalid fixture fails, and enforces per-schema coverage (every schema must have at least one valid and one invalid fixture). `check:compatibility` verifies the mirror baseline matches the released core set.

## Reference Model

Every contract carries `api_version` (`tekmovela.io/v1alpha1`) and a fixed `kind`. Refs to published objects are `{kind, id, version, digest}` with `sha256:` digests. See [COMPATIBILITY_POLICY.md](COMPATIBILITY_POLICY.md) for immutability and compatibility rules.
