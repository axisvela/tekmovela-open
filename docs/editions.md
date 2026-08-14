# TEKMOVELA OSS and Enterprise Edition Boundaries

## Principle

The OSS open core must complete a real Requirement/Contract -> Harness -> Run -> Evidence -> Gate chain, not just expose an SDK. Enterprise value comes from scale, governance, continuous operation, certification, and compliance.

| Layer | Open Core (OSS) | Enterprise Edition |
|---|---|---|
| Contract & conformance | Versioned schemas, fixtures, validator, SDK foundations | Certification service, certified harness components, supplier/contributor conformance |
| Verification & harness | VerificationContract, HarnessVersion, local registry, Runner/Adapter/Oracle/Reporter SDK | Federated registry, distributed runner, scheduler, quota, lifecycle governance |
| Testing | Scenario, EvaluationSet, deterministic/property evaluators, local runner | Large-scale statistical eval, private judge governance, human calibration |
| Evidence | EvidenceBundle, TraceEnvelope, generic SFE profile, basic Evidence Explorer | Production assurance, drift/anomaly, retention, SLO, incident integration, advanced attribution |
| Assurance | Claims, coverage, local gate workflow, waiver/attestation on local store | Signed attestation, waiver workflow, policy-driven release gate, audit ledger |
| Collaboration | Role/contract references, contributor gate port, PM adapter SPI | Enterprise connector packs, certified integration, benchmark certification |

## Explicit non-boundaries

- AEGIVELA remains the production authorization authority in both editions.
- Contract schemas and security-critical validation stay in the open core.
- Enterprise Edition never creates a second contract dialect or moves tenant isolation or contract validation out of Core.
- Enterprise modules depend on published OSS ports, never on core `internal/*` packages.

See the private core repository for the authoritative boundary document.
