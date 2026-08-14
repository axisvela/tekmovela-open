# TEKMOVELA Open

This is the OpenCode project rules file for the public TEKMOVELA repository. Claude Code reads the matching `CLAUDE.md`.

Keep this file and `CLAUDE.md` aligned when changing project instructions.

## Harness

Use Superpowers as the project development harness. OpenCode loads the upstream plugin declared in `opencode.json`; `superpowers/` contains project-local harness notes.

## Purpose

`tekmovela-open` is the public, Apache-2.0 face of TEKMOVELA: documentation, versioned public contracts (mirrored from the core repository), SDKs, integration examples, and links to Core binary releases.

## Rules

- No core runtime source (AGPL), no enterprise features, and no architecture / roadmap / planning documents belong here. Point to the private core and EE repositories instead.
- Published contracts in `contracts/` are a mirror of the authoritative set in the core repository, pinned to a released version. Update the mirror by copying from a released core version; never edit a published schema in place.
- Every SDK or example must validate against the mirrored contracts (`npm run test:contracts` from `contracts/`).
- Keep public documentation accurate and current; never invent API shapes not present in the contracts.
- Do not duplicate internal design or delivery-planning documents in this repository.
