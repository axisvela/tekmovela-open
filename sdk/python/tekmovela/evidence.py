"""EvidenceBundle emission.

A portable, digest-verifiable collection of evidence from a run. Raw telemetry
is not evidence; only records bound to versions, sources, digests, scope, and
claims become evidence.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional

from .contracts import ContractError, digest_of, parse_digest

TRACE = "trace"
EFFECT = "effect"
FAILURE = "failure"
ATTRIBUTION = "attribution"
EXPLANATION = "explanation"
AUTHORITY = "authority"
MISSING = "missing"


@dataclass
class Item:
    kind: str
    uri: str
    digest: str
    note: Optional[str] = None

    def to_dict(self) -> Dict[str, Any]:
        out: Dict[str, Any] = {"kind": self.kind, "uri": self.uri, "digest": self.digest}
        if self.note:
            out["note"] = self.note
        return out


@dataclass
class EvidenceBundle:
    id: str
    run_ref: str
    verification_contract_ref: str
    harness_version_ref: str
    items: List[Item] = field(default_factory=list)
    scope: Dict[str, str] = field(default_factory=dict)
    environment_digest: Optional[str] = None
    bundle_digest: Optional[str] = None

    def add_item(self, item: Item) -> "EvidenceBundle":
        if not item.kind or not item.uri:
            raise ContractError("item kind and uri are required")
        parse_digest(item.digest)
        self.items.append(item)
        return self

    def add_missing(self, uri: str, digest: str, note: str) -> "EvidenceBundle":
        return self.add_item(Item(kind=MISSING, uri=uri, digest=digest, note=note))

    def _body(self) -> Dict[str, Any]:
        return {
            "api_version": "tekmovela.io/v1alpha1",
            "kind": "EvidenceBundle",
            "id": self.id,
            "run_ref": self.run_ref,
            "verification_contract_ref": self.verification_contract_ref,
            "harness_version_ref": self.harness_version_ref,
            "environment_digest": self.environment_digest,
            "scope": self.scope,
            "items": [i.to_dict() for i in self.items],
        }

    def digest(self) -> str:
        return digest_of(self._body())

    def seal(self) -> "EvidenceBundle":
        if not self.items:
            raise ContractError(f"bundle {self.id} cannot seal with no items")
        self.bundle_digest = self.digest()
        return self

    def verify(self) -> None:
        for item in self.items:
            parse_digest(item.digest)
        if not self.bundle_digest:
            raise ContractError(f"bundle {self.id} is not sealed")
        if self.digest() != self.bundle_digest:
            raise ContractError(f"bundle {self.id} digest mismatch")

    def missing(self) -> List[Item]:
        return [i for i in self.items if i.kind == MISSING]
