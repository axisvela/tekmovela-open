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

# --- evidence golden (unset + set env) ---
EV_GO_UNSET="$(cd "$SDK/go" && go run ./internal/parity evidence | head -1)"
EV_GO_SET="$(cd "$SDK/go" && go run ./internal/parity evidence | tail -1)"
EV_PY_UNSET="$(cd "$SDK/python" && PYTHONPATH=. python -c "
from tekmovela.evidence import EvidenceBundle, Item, TRACE
b = EvidenceBundle('bundle.e0001', 'run.1', 'checkout.atomicity@v1', 'checkout.double_spend@v1')
b.add_item(Item(kind=TRACE, uri='trace.1', digest='sha256:' + 'a'*64))
print(b.digest())
")"
EV_PY_SET="$(cd "$SDK/python" && PYTHONPATH=. python -c "
from tekmovela.evidence import EvidenceBundle, Item, TRACE
b = EvidenceBundle('bundle.e0001', 'run.1', 'checkout.atomicity@v1', 'checkout.double_spend@v1')
b.add_item(Item(kind=TRACE, uri='trace.1', digest='sha256:' + 'a'*64))
b.environment_digest = 'sha256:' + 'd'*64
print(b.digest())
")"
EV_TS_UNSET="$(cd "$SDK/typescript" && node -e "
const {EvidenceBundle,ItemKind}=require('./dist/index.js');
const b=new EvidenceBundle({id:'bundle.e0001',runRef:'run.1',verificationContractRef:'checkout.atomicity@v1',harnessVersionRef:'checkout.double_spend@v1'});
b.addItem({kind:ItemKind.Trace,uri:'trace.1',digest:'sha256:'+'a'.repeat(64)});
console.log(b.digest());
")"
EV_TS_SET="$(cd "$SDK/typescript" && node -e "
const {EvidenceBundle,ItemKind}=require('./dist/index.js');
const b=new EvidenceBundle({id:'bundle.e0001',runRef:'run.1',verificationContractRef:'checkout.atomicity@v1',harnessVersionRef:'checkout.double_spend@v1'});
b.addItem({kind:ItemKind.Trace,uri:'trace.1',digest:'sha256:'+'a'.repeat(64)});
b.environmentDigest='sha256:'+'d'.repeat(64);
console.log(b.digest());
")"
[ "$EV_GO_UNSET" = "$EV_PY_UNSET" ] || fail "evidence-unset Go=$EV_GO_UNSET != Python=$EV_PY_UNSET"
[ "$EV_GO_UNSET" = "$EV_TS_UNSET" ] || fail "evidence-unset Go=$EV_GO_UNSET != TS=$EV_TS_UNSET"
[ "$EV_GO_SET" = "$EV_PY_SET" ] || fail "evidence-set Go=$EV_GO_SET != Python=$EV_PY_SET"
[ "$EV_GO_SET" = "$EV_TS_SET" ] || fail "evidence-set Go=$EV_GO_SET != TS=$EV_TS_SET"
echo "evidence goldens OK: unset=$EV_GO_UNSET set=$EV_GO_SET"

# --- harness golden ---
HARN_GO="$(cd "$SDK/go" && go run ./internal/parity harness)"
HARN_PY="$(cd "$SDK/python" && PYTHONPATH=. python -c "
from tekmovela.contracts import Reference
from tekmovela.harness import HarnessVersion, component
hv = HarnessVersion(
    id='checkout.double_spend', version='v1',
    verification_contract_ref=Reference('VerificationContract', 'checkout.atomicity', 'v1', 'sha256:' + 'a'*64),
    component_lock={
        'runner': component('runner', 'runner.deterministic', 'v1', 'sha256:' + 'c'*64),
        'adapter': component('adapter', 'adapter.praxovela', 'v1', 'sha256:' + 'd'*64),
        'oracle': component('oracle', 'oracle.effect_ledger', 'v1', 'sha256:' + 'e'*64),
        'reporter': component('reporter', 'reporter.evidence', 'v1', 'sha256:' + 'f'*64),
    },
    evidence_schema='tekmovela.evidence-bundle.v1.schema.json',
)
print(hv.canonical_digest())
")"
HARN_TS="$(cd "$SDK/typescript" && node -e "
const {HarnessVersion,Reference}=require('./dist/index.js');
const hv=new HarnessVersion({id:'checkout.double_spend',version:'v1',verificationContractRef:new Reference('VerificationContract','checkout.atomicity','v1','sha256:'+'a'.repeat(64)),componentLock:{runner:{kind:'HarnessComponent',id:'runner.deterministic',version:'v1',digest:'sha256:'+'c'.repeat(64)},adapter:{kind:'HarnessComponent',id:'adapter.praxovela',version:'v1',digest:'sha256:'+'d'.repeat(64)},oracle:{kind:'HarnessComponent',id:'oracle.effect_ledger',version:'v1',digest:'sha256:'+'e'.repeat(64)},reporter:{kind:'HarnessComponent',id:'reporter.evidence',version:'v1',digest:'sha256:'+'f'.repeat(64)}},evidenceSchema:'tekmovela.evidence-bundle.v1.schema.json'});
console.log(hv.canonicalDigest());
")"
[ "$HARN_GO" = "$HARN_PY" ] || fail "harness Go=$HARN_GO != Python=$HARN_PY"
[ "$HARN_GO" = "$HARN_TS" ] || fail "harness Go=$HARN_GO != TS=$HARN_TS"
echo "harness golden OK: $HARN_GO"

echo "cross-SDK parity OK"
