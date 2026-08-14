# TEKMOVELA Contract Compatibility Policy

## Principles

1. **Published contract versions are immutable.** A published `v1alpha1` schema is never edited in place. Behavioral changes create a new schema version (`v1alpha2`, `v1beta1`, ...).
2. **Digests bind evidence to versions.** Refs to published objects carry `sha256:` digests. A digest mismatch invalidates the reference.
3. **Lifecycle is not a behavior change.** Contract lifecycle transitions (draft -> published -> superseded/revoked) never change the digest of the frozen content.
4. **Compatibility is asserted, not assumed.** Every change that affects a schema must ship positive fixtures (must validate) and negative fixtures (must fail) together with the change.

## Change Rules

| Change | Rule |
|---|---|
| Add optional field | Compatible; add positive fixture. |
| Tighten a constraint (new required field, narrower enum, stronger pattern) | Breaking; new schema version required. |
| Widen a constraint (drop required field, add enum value) | Backward-compatible for readers; new schema version for new writers. |
| Rename or reinterpret a field | Breaking; new schema version required. |
| Emergency stop of a contract version | Do not edit; mark the published version `revoked` and create a replacement. |

## Governance

- A contract change that alters the reference model, lifecycle rules, or a cross-product adapter contract must also update an ADR under `docs/adr/`.
- The authoritative contract set lives in this core repository (`tekmovela`). The public `tekmovela-open/contracts` mirrors published schemas and is pinned to a released version.
- Any C/D research adoption level promoted to A/B requires a versioned ADR listing empirical evidence, failure modes, backward compatibility, and rollout strategy.
