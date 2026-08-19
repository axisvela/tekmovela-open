"""Harness authoring helpers.

A harness is a first-class engineering asset: Runner/Adapter/Oracle/Reporter
components frozen into an immutable HarnessVersion with a digest.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, List, Mapping, Optional

from .contracts import API_VERSION, ContractError, digest_of, parse_digest

COMPONENT_RUNNER = "runner"
COMPONENT_ADAPTER = "adapter"
COMPONENT_ORACLE = "oracle"
COMPONENT_REPORTER = "reporter"


@dataclass
class HarnessVersion:
    id: str
    version: str
    verification_contract_ref: Mapping[str, str]
    component_lock: Dict[str, Dict[str, str]]
    evidence_schema: str
    determinism_mode: str = "deterministic"
    seed: Optional[str] = None
    digest: Optional[str] = None

    def __post_init__(self) -> None:
        for field_name in ("kind", "id", "version"):
            if not self.verification_contract_ref.get(field_name):
                raise ContractError(f"verification_contract_ref.{field_name} is required")
        parse_digest(self.verification_contract_ref.get("digest", ""))
        for role in (COMPONENT_RUNNER, COMPONENT_ADAPTER, COMPONENT_ORACLE, COMPONENT_REPORTER):
            ref = self.component_lock.get(role)
            if not ref:
                raise ContractError(f"component_lock.{role} is required")
            parse_digest(ref["digest"])

    def _body(self) -> Dict[str, Any]:
        body: Dict[str, Any] = {
            "api_version": API_VERSION,
            "kind": "HarnessVersion",
            "id": self.id,
            "version": self.version,
            "verification_contract_ref": dict(self.verification_contract_ref),
            "component_lock": self.component_lock,
            "evidence_schema": self.evidence_schema,
            "determinism_profile": {"mode": self.determinism_mode},
            "isolation_profile": {},
        }
        if self.seed is not None:
            body["environment_contract"] = {"seed": self.seed}
        return body

    def canonical_digest(self) -> str:
        return digest_of(self._body())

    def publish(self) -> "HarnessVersion":
        self.digest = self.canonical_digest()
        return self


def component(kind: str, id: str, version: str, digest: str) -> Dict[str, str]:
    if kind not in (COMPONENT_RUNNER, COMPONENT_ADAPTER, COMPONENT_ORACLE, COMPONENT_REPORTER):
        raise ContractError(f"unknown component kind {kind!r}")
    parse_digest(digest)
    return {"kind": "HarnessComponent", "id": id, "version": version, "digest": digest}
