# TEKMOVELA Open SDKs

Public SDK references for consuming TEKMOVELA. All SDKs validate against the mirrored contracts in `../contracts/`.

| SDK | Module | Status |
|---|---|---|
| Go | [`go/`](go/) | Implemented: `reference` package (digest/reference rules) |
| Python | [`python/`](python/) | Implemented: contracts, evidence bundle, harness authoring, local deterministic runner |
| TypeScript | [`typescript/`](typescript/) | Skeleton; UI SFE capture + trace adapter planned |

## Go

```bash
cd go
go test ./...
```

The `reference` package implements the `{kind, id, version, digest}` reference and `sha256:<hex>` digest rules shared by all TEKMOVELA contracts.

## Python

```bash
cd python
PYTHONPATH=. python -m unittest discover -s tests -p "test_*.py"
```

See [`python/README.md`](python/README.md) for module documentation and usage.
