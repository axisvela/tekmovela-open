# TEKMOVELA Examples

Integration examples for the TEKMOVELA open core.

| Example | Language | Demonstrates |
|---|---|---|
| `go-verify` | Go | Parsing an EvidenceBundle fixture and verifying every `sha256:` digest with the Go SDK |

## go-verify

```bash
cd go-verify
go run . ../../contracts/evidence/fixtures/valid/run-evidence.json
```

Expect `PASSED: all digests well-formed`. Feed it the invalid fixture to see digest checks fail:

```bash
go run . ../../contracts/evidence/fixtures/invalid/bad-bundle-digest.json
```
