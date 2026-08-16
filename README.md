# TEKMOVELA — Autonomous-System Engineering Assurance Control Plane

**TEKMOVELA** is the **AxisRobo engineering assurance control plane for autonomous systems**. It turns intent and acceptance criteria into verification contracts and governed harnesses, executes reproducible closed-loop testing of autonomous systems, diagnoses failures with trace-linked evidence, gates releases, and continuously triggers re-verification from production behavior.

> **Read this page in [中文（简体）](docs/zh-CN/README.md) · English is the primary documentation language.**

---

## What problem does TEKMOVELA solve?

Autonomous systems fail in ways that are hard to prove absent. TEKMOVELA makes assurance **reproducible, evidence-backed, and continuously re-checkable** rather than a one-time checklist.

| Problem | TEKMOVELA's answer |
|---|---|
| Intent is vague and untestable | **Requirement Bridge** — acceptance criteria, constraints, and requirement trace links with source digests and testability verification. |
| Testing is ad hoc and unrepeatable | **Verification Contracts & Governed Harnesses** — frozen, digest-locked contracts that bind scenarios to a Runner/Adapter/Oracle/Reporter component lock. |
| Agent runs are hard to reproduce | **Agentic Testing** — scenarios, evaluation sets, durable `TestControlLoop` with interrupt/budget semantics, and deterministic replay classification. |
| Failures lack evidence | **Evidence & Diagnosis** — trace envelopes, evidence bundles, structured failure evidence, failure clusters, causal attribution, and decision explanation. |
| Releases are "trust me" | **Assurance & Release** — scoped, expiring claims; coverage statements; multi-layer G0–G5 gates; waivers; signed attestations. |
| Production drift is invisible | **Runtime Assurance** — behavioral bounds, anomaly signals, and minimal re-verification sets triggered from production behavior. |
| Engineering scale is chaotic | **Engineering Collaboration** — role/contract references, state links, and a zero-trust contributor gate. |

## Product features

- **Contract-first** — versioned JSON Schema for Intent / Verification / Harness / Testing / Evidence / Assurance / Runtime object families, with positive/negative fixtures and a compatibility policy. Published contracts are immutable.
- **Verification Kernel** — `VerificationContract` versions, lifecycle, canonical digests; Requirement Gate semantics; the `tek` CLI.
- **Harness Core** — Runner/Adapter/Oracle/Reporter component model, `HarnessVersion` component locks, registry.
- **Agentic Test Lab** — Scenario, EvaluationSet, durable `TestControlLoop`, UI-SFE failure profile, replay classification, failure clusters, regression adoption, closed-loop regression.
- **Assurance Control Plane** — assurance cases, gate policies (G0–G5), gate orchestration, waivers, digest-verifiable append-only records, a tenant-aware HTTP control plane.
- **Runtime Assurance** — behavioral bounds, anomaly lifecycle, minimal reverification sets, per-run accountability records (AAR).
- **Evidence-authoritative** — raw telemetry is never evidence; only records bound to versions, sources, digests, scope, and claims become assurance evidence. Never a score or projection over a signed contract.
- **Cross-product adapters** — stable contracts with PRAXOVELA, PEIRAVELA, AEGIVELA, MODUREGIS, LENS, and other AxisRobo products.

---

# TEKMOVELA Open

This repository is the **public, Apache-2.0 face** of the TEKMOVELA ecosystem. It publishes what external teams, researchers, and integrators need without exposing the AGPL core or proprietary enterprise modules:

- **Contracts** — a mirror of the authoritative, versioned public contracts, pinned to a released version.
- **SDKs** — Go, Python, and TypeScript SDK references.
- **Examples** — integration and usage examples (e.g. `examples/go-verify`).
- **Documentation** — API reference, architecture overview, OSS/EE boundary, and roadmap.
- **Core binary releases** — linked from GitHub Releases of the core repository.

The **core implementation** lives in a separate (private, AGPL-3.0) repository; **enterprise extensions** live in another private repository. Internal architecture, design, and delivery-planning documents are not duplicated here.

## Repository layout

```
tekmovela-open/        ← You are here. Public homepage, docs, contracts, SDK, examples.
├── README.md
├── LICENSE            (Apache 2.0)
├── docs/              Public architecture, API, roadmap, OSS/EE boundary (+ zh-CN)
├── contracts/         Mirrored public JSON Schema contracts and fixtures
├── sdk/go/            Go SDK module
├── sdk/python/        Python SDK
├── sdk/typescript/    TypeScript SDK
├── examples/          Integration and usage examples
└── scripts/           Build and release tooling
tekmovela/             Core Go implementation (AGPL-3.0, private)
tekmovela-ee/          Enterprise extensions (Enterprise License, private)
```

## What this repo is NOT

No core runtime source (AGPL), no enterprise features (proprietary), and no architecture / roadmap / planning documents. The source of truth for core code, the product plan, and architecture is the private core repository. Published contracts here are a mirror of the authoritative core set, pinned to a released version.

## Ecosystem

| Repo | License | Visibility | Contents |
|---|---|---|---|
| `tekmovela` | AGPL-3.0 | Private | Core implementation, authoritative contracts, product plan |
| `tekmovela-open` | Apache-2.0 | **Public** | SDK/API docs, contracts, examples, Core binary releases |
| `tekmovela-ee` | Proprietary | Private | Enterprise modules, internal architecture/design/planning |

## Quick start for integrators

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

**English (primary):**

- [API reference](docs/api.md)
- [Architecture overview](docs/architecture/overview.md)
- [Roadmap](docs/roadmap.md)
- [OSS/EE boundary](docs/editions.md)
- [License](LICENSE)

**中文（简体）：**

- [TEKMOVELA 文档导航](docs/zh-CN/README.md)

## License

The TEKMOVELA Open Core (`tekmovela-open`) is licensed under the [Apache License 2.0](LICENSE).

| Repository | License |
|---|---|
| `tekmovela-open` (this repo) | Apache 2.0 |
| `tekmovela` (core implementation) | AGPL-3.0 |
| `tekmovela-ee` (enterprise extensions) | Enterprise License |
