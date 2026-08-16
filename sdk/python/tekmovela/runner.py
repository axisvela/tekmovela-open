"""Local deterministic runner.

Executes a Scenario against a RuntimeAdapter and emits a sealed
EvidenceBundle. Deterministic because everything derives from the scenario seed
and the frozen harness.
"""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Protocol

from .contracts import digest_from_bytes
from .evidence import EFFECT, TRACE, EvidenceBundle, Item


class RuntimeAdapter(Protocol):
    def start_run(self, scenario_ref: str, harness_version_ref: str, sut_ref: str, seed: str) -> str: ...
    def fetch_effects(self, run_ref: str) -> List[Dict[str, str]]: ...


@dataclass
class Scenario:
    id: str
    version: str
    verification_contract_ref: str
    seed: str
    harness_version_ref: str = ""
    focus: Optional[str] = None


@dataclass
class RunResult:
    run_id: str
    status: str
    bundle: EvidenceBundle
    effects: List[Dict[str, str]] = field(default_factory=list)


class LocalRunner:
    def __init__(self, runtime: RuntimeAdapter):
        self.runtime = runtime

    def execute(self, scenario: Scenario, sut_ref: str, run_id: str) -> RunResult:
        run_ref = self.runtime.start_run(
            scenario_ref=scenario.id,
            harness_version_ref=scenario.harness_version_ref or "local@v1",
            sut_ref=sut_ref,
            seed=scenario.seed,
        )
        effects = self.runtime.fetch_effects(run_ref)

        bundle = EvidenceBundle(
            id=f"{run_id}.evidence",
            run_ref=run_id,
            verification_contract_ref=scenario.verification_contract_ref,
            harness_version_ref=scenario.harness_version_ref or "local@v1",
            scope={"intent_ref": scenario.id, "capability_ref": sut_ref},
        )
        bundle.add_item(Item(kind=TRACE, uri="local://trace/exec.jsonl", digest=digest_from_bytes(f"{scenario.seed}|{run_id}".encode())))
        bundle.add_item(Item(kind=EFFECT, uri="local://trace/effects.jsonl", digest=digest_from_bytes(f"{len(effects)} effects".encode())))
        bundle.seal()
        bundle.verify()

        return RunResult(run_id=run_id, status="passed", bundle=bundle, effects=effects)
