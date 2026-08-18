#!/usr/bin/env bash
# Cross-SDK digest parity check.
#
# Computes digests for identical logical documents with the Go, Python, and
# TypeScript SDKs and asserts they all agree. Run from the repository root.
#
# Requires: go, python3, node (>= 23.6 for TS type-stripping).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SDK="$ROOT/sdk"

fail() {
  echo "PARITY FAIL: $1" >&2
  exit 1
}

# --- contracts golden: {"a":1,"b":2} -> 43258cff... ---
CONTRACTS_GO="$(cd "$SDK/go" && go run ./internal/parity contracts)"
CONTRACTS_PY="$(cd "$SDK/python" && PYTHONPATH=. python -c "from tekmovela.contracts import digest_of; print(digest_of({'a': 1, 'b': 2}))")"
CONTRACTS_TS="$(cd "$SDK/typescript" && node -e "const {digestOf}=require('./dist/index.js'); console.log(digestOf({a:1,b:2}))")"
[ "$CONTRACTS_GO" = "$CONTRACTS_PY" ] || fail "contracts Go=$CONTRACTS_GO != Python=$CONTRACTS_PY"
[ "$CONTRACTS_GO" = "$CONTRACTS_TS" ] || fail "contracts Go=$CONTRACTS_GO != TS=$CONTRACTS_TS"
echo "contracts golden OK: $CONTRACTS_GO"

# --- reference golden ---
REF_GO="$(cd "$SDK/go" && go run ./internal/parity reference)"
REF_PY="$(cd "$SDK/python" && PYTHONPATH=. python -c "from tekmovela.contracts import Reference, digest_of; print(digest_of(Reference('VerificationContract','checkout.atomicity','v1','sha256:'+'a'*64)))")"
REF_TS="$(cd "$SDK/typescript" && node -e "const {Reference,digestOf}=require('./dist/index.js'); console.log(digestOf(new Reference('VerificationContract','checkout.atomicity','v1','sha256:'+'a'.repeat(64))))")"
[ "$REF_GO" = "$REF_PY" ] || fail "reference Go=$REF_GO != Python=$REF_PY"
[ "$REF_GO" = "$REF_TS" ] || fail "reference Go=$REF_GO != TS=$REF_TS"
echo "reference golden OK: $REF_GO"

echo "cross-SDK parity OK"
