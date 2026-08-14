# TEKMOVELA Open

**TEKMOVELA** is the AxisRobo Autonomous-System Engineering Assurance Control Plane — the engineering assurance control plane that turns intent and acceptance criteria into verification contracts and governed harnesses, executes reproducible closed-loop testing of autonomous systems, diagnoses failures with trace-linked evidence, gates releases, and continuously triggers re-verification from production behavior.

This is the **public open-source face** of the TEKMOVELA ecosystem, published under the **Apache License 2.0**. It contains the project overview, documentation, versioned public contracts, SDKs, integration examples, and links to Core binary releases. The core implementation lives in a separate repository.

## What TEKMOVELA Does

TEKMOVELA connects stakeholder intent, verification substrate, agentic testing, failure diagnosis, engineering accountability, release decision, and runtime re-verification into an **Engineering Assurance Chain** that is enterprise-verifiable, issuable, revocable, and continuously updatable.

### Open Core Capabilities

- **Contracts** — Versioned JSON Schema for the Intent/Verification/Harness/Evidence/Assurance/Runtime object families with positive/negative fixtures and a compatibility policy.
- **Verification Kernel** — VerificationContract versions, lifecycle, and canonical digests; Requirement Gate semantics.
- **Harness Core** — Runner/Adapter/Oracle/Reporter component model, HarnessVersion component locks, local registry.
- **Agentic Test Lab (planned)** — Scenario, EvaluationSet, TestControlLoop with interrupt/budget semantics, StructuredFailureEvidence.
- **Assurance (planned)** — AssuranceClaim, CoverageStatement, GateDecision, Waiver, Attestation on local workflow.
- **SDKs** — Go, Python, and TypeScript SDK references.
- **Core binary** — Released from the core repository under GitHub Releases.

## Repository Layout

```
tekmovela-open/        ← You are here. Public homepage, docs, contracts, SDK, examples.
├── README.md
├── LICENSE            (Apache 2.0)
├── docs/              Public architecture, API, roadmap, OSS/EE boundary
├── contracts/         Mirrored public JSON Schema contracts and fixtures
├── sdk/go/            Go SDK module
├── sdk/python/        Python SDK
├── sdk/typescript/    TypeScript SDK
├── examples/          Integration and usage examples
└── scripts/           Build and release tooling
tekmovela/             Core Go implementation (AGPL-3.0, private)
tekmovela-ee/          Enterprise extensions (Enterprise License, private)
```

## What This Repo is NOT

No core runtime source (AGPL), no enterprise features (proprietary), and no architecture / roadmap / planning documents. The source of truth for core code, the product plan, and architecture is the private core repository; internal architecture, design, and planning documents live in the private EE repository. Published contracts here are a mirror of the authoritative core set, pinned to a released version.

## Ecosystem

| Repo | License | Visibility | Contents |
|---|---|---|---|
| `tekmovela` | AGPL-3.0 | Private | Core implementation, authoritative contracts, product plan |
| `tekmovela-open` | Apache-2.0 | **Public** | SDK/API docs, contracts, examples, Core binary releases |
| `tekmovela-ee` | Proprietary | Private | Enterprise modules, internal architecture/design/planning |

## Quick Start for Integrators

1. **Validate the contracts**:
   ```bash
   cd contracts
   npm install
   npm run test:contracts
   ```
2. **Read the contract reference** — start with `contracts/verification/tekmovela.verification-contract.v1alpha1.schema.json` and the intent family under `contracts/intent/`.
3. **Use the Go SDK** — `cd sdk/go && go test ./...`.
4. **Download the Core binary** from GitHub Releases when published.

## Documentation

- [API reference](docs/api.md)
- [Architecture overview](docs/architecture/overview.md)
- [Roadmap](docs/roadmap.md)
- [OSS/EE boundary](docs/editions.md)
- [License](LICENSE)

## License

The TEKMOVELA Open Core (`tekmovela-open`) is licensed under the [Apache License 2.0](LICENSE).

| Repository | License |
|---|---|
| `tekmovela-open` (this repo) | Apache 2.0 |
| `tekmovela` (core implementation) | AGPL-3.0 |
| `tekmovela-ee` (enterprise extensions) | Enterprise License |
