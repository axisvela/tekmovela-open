"""Tests for the TEKMOVELA Python SDK."""

import hashlib
import json
import os
import tempfile
import unittest
from typing import Dict, List

from tekmovela.contracts import ContractError, Reference, Schema, digest_of, parse_digest, validate_fixture
from tekmovela.evidence import EvidenceBundle, Item, MISSING, TRACE, EFFECT
from tekmovela.harness import HarnessVersion, component
from tekmovela.runner import LocalRunner, Scenario

DIGEST = "sha256:" + "a" * 64
GOLDEN_DIGEST = "sha256:43258cff783fe7036d8a43033f830adfc60ec037382473548ac742b888292777"


class ContractTests(unittest.TestCase):
    def test_digest_rules(self):
        parse_digest(DIGEST)
        with self.assertRaises(ContractError):
            parse_digest("md5:abc")
        self.assertEqual(len(digest_of({"a": 1})), 7 + 64)

    def test_golden_digest(self):
        self.assertEqual(digest_of({"a": 1, "b": 2}), GOLDEN_DIGEST)

    def test_canonical_digest_is_key_order_stable(self):
        self.assertEqual(digest_of({"a": 1, "b": 2}), digest_of({"b": 2, "a": 1}))

    def test_reference(self):
        r = Reference("VerificationContract", "checkout.atomicity", "v1", DIGEST)
        self.assertEqual(r["kind"], "VerificationContract")
        with self.assertRaises(ContractError):
            Reference("k", "i", "v1", "bad")

    def test_schema_kind_mismatch(self):
        schema = Schema({
            "properties": {
                "api_version": {"const": "tekmovela.io/v1alpha1"},
                "kind": {"const": "AcceptanceCriterion"},
            },
            "required": ["id", "digest", "testable"],
        })
        with self.assertRaises(ContractError):
            schema.validate({"api_version": "tekmovela.io/v1alpha1", "kind": "Other", "id": "x"})


class EvidenceTests(unittest.TestCase):
    def test_seal_and_verify(self):
        b = EvidenceBundle("bundle.e0001", "run.001", "checkout.atomicity", "h@v1")
        b.add_item(Item(kind=TRACE, uri="s3://t", digest=DIGEST))
        b.add_item(Item(kind=EFFECT, uri="s3://e", digest=DIGEST))
        b.seal()
        b.verify()
        self.assertEqual(len(b.missing()), 0)

    def test_tamper_detected(self):
        b = EvidenceBundle("bundle.e0002", "run.001", "c", "h")
        b.add_item(Item(kind=TRACE, uri="s3://t", digest=DIGEST))
        b.seal()
        b.items[0].uri = "s3://tampered"
        with self.assertRaises(ContractError):
            b.verify()

    def test_missing_is_evidence(self):
        b = EvidenceBundle("bundle.e0003", "run.001", "c", "h")
        b.add_item(Item(kind=TRACE, uri="s3://t", digest=DIGEST))
        b.add_missing("s3://authority.jsonl", DIGEST, "not captured")
        self.assertEqual(len(b.missing()), 1)


class HarnessTests(unittest.TestCase):
    def test_publish_digest(self):
        hv = HarnessVersion(
            id="checkout.double_spend",
            version="v1",
            verification_contract_ref="checkout.atomicity",
            component_lock={
                "runner": component("runner", "runner.deterministic", "v1", DIGEST),
                "adapter": component("adapter", "adapter.praxovela", "v1", DIGEST),
                "oracle": component("oracle", "oracle.effect_ledger", "v1", DIGEST),
                "reporter": component("reporter", "reporter.evidence", "v1", DIGEST),
            },
            evidence_schema="tekmovela.evidence-bundle.v1alpha1.schema.json",
        )
        hv.publish()
        self.assertTrue(hv.digest)

    def test_incomplete_lock(self):
        with self.assertRaises(ContractError):
            HarnessVersion(
                id="x", version="v1", verification_contract_ref="c",
                component_lock={"runner": component("runner", "r", "v1", DIGEST)},
                evidence_schema="s",
            )


class FakeRuntime:
    def start_run(self, scenario_ref, harness_version_ref, sut_ref, seed) -> str:
        return f"{scenario_ref}/run-{seed}"

    def fetch_effects(self, run_ref) -> List[Dict[str, str]]:
        return [{"action": "debit", "resource": "ledger", "state_delta": "-1"}]


class RunnerTests(unittest.TestCase):
    def test_execute_emits_sealed_bundle(self):
        runner = LocalRunner(FakeRuntime())
        scenario = Scenario(id="scenario.double_spend", version="v1", verification_contract_ref="checkout.atomicity", seed="s1")
        result = runner.execute(scenario, sut_ref="cap/v2", run_id="run.e0001")
        self.assertEqual(result.status, "passed")
        result.bundle.verify()
        self.assertEqual(len(result.effects), 1)

    def test_validate_contract_fixture(self):
        with tempfile.TemporaryDirectory() as tmp:
            schema_path = os.path.join(tmp, "schema.json")
            fixture_path = os.path.join(tmp, "fixture.json")
            with open(schema_path, "w", encoding="utf-8") as fh:
                json.dump({"properties": {"kind": {"const": "AcceptanceCriterion"}, "api_version": {"const": "tekmovela.io/v1alpha1"}}, "required": ["kind", "id"]}, fh)
            with open(fixture_path, "w", encoding="utf-8") as fh:
                json.dump({"kind": "AcceptanceCriterion", "api_version": "tekmovela.io/v1alpha1", "id": "checkout.atomicity"}, fh)
            with open(schema_path, "r", encoding="utf-8") as fh:
                schema_doc = json.load(fh)
            validate_fixture(Schema(schema_doc), fixture_path)


if __name__ == "__main__":
    unittest.main()
