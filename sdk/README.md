# TEKMOVELA Open SDKs

Public SDK references for consuming TEKMOVELA. All SDKs validate against the mirrored contracts in `../contracts/`.

| SDK | Module | Status |
|---|---|---|
| Go | [`go/`](go/) | Implemented: `reference` package (digest/reference rules) |
| Python | [`python/`](python/) | Skeleton; harness authoring + evidence emitter planned |
| TypeScript | [`typescript/`](typescript/) | Skeleton; UI SFE capture + trace adapter planned |

## Go

```bash
cd go
go test ./...
```

The `reference` package implements the `{kind, id, version, digest}` reference and `sha256:<hex>` digest rules shared by all TEKMOVELA contracts.
