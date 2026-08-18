# TEKMOVELA Open SDKs

Public SDK references for consuming TEKMOVELA. All SDKs validate against the mirrored contracts in `../contracts/`.

| SDK | Module | Status |
|---|---|---|
| Go | [`go/`](go/) | Implemented: contracts, evidence bundle, harness authoring, local runner |
| Python | [`python/`](python/) | Implemented: contracts, evidence bundle, harness authoring, local deterministic runner |
| TypeScript | [`typescript/`](typescript/) | Implemented: contracts, evidence, harness, UI-SFE capture, trace |

## Go

```bash
cd go
go test ./...
```

See [`go/README.md`](go/README.md) for module documentation and usage.

## Python

```bash
cd python
PYTHONPATH=. python -m unittest discover -s tests -p "test_*.py"
```

See [`python/README.md`](python/README.md) for module documentation and usage.
