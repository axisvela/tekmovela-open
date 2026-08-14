# TEKMOVELA API Reference

The TEKMOVELA control plane exposes REST (and planned gRPC) surfaces for the core object families. Request/response shapes are defined by the [contracts](../contracts/) and are the source of truth. This page documents the public surface for integrators.

## Core resources

| Family | Resources | Notes |
|---|---|---|
| Intents | intents, acceptance-criteria, constraints, requirement-trace-links | References with digests; source authority is IRIS/external |
| Verification | verification-contracts | Immutable versions; digest-bound refs |
| Harness | harness-definitions, harness-versions, harness-components, coverage-debts | Component lock model |
| Testing | scenarios, evaluation-sets, runs, test-control-loops | Interrupt/budget semantics |
| Evidence | trace-envelopes, evidence-bundles, failure-evidence, assurance-evidence | Digest-verifiable |
| Assurance | release-candidates, claims, coverage, gates, waivers, attestations | Scoped + expiring |
| Runtime | behavioral-bounds, anomaly-signals, reverification-requests | Trigger provenance |

## CLI

The `tek` CLI mirrors the API:

```text
tek verify     compile a VerificationContract and run its harness
tek harness    register/inspect harnesses and components
tek test       run a Scenario / EvaluationSet
tek explain    open failure evidence and attribution
tek assure     build an AssuranceClaim
tek gate       evaluate a gate decision
tek reverify   trigger targeted re-verification
```

## Digests and references

Every reference to a published object carries `{kind, id, version, digest}` with `sha256:<hex>` digests. API clients must verify digests on evidence payloads. See [COMPATIBILITY_POLICY](../contracts/COMPATIBILITY_POLICY.md).

Versioned OpenAPI specifications will be published under this directory as the M1 API is defined.
