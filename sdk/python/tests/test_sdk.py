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
GOLDEN_HARNESS_VERSION_DIGEST = "sha256:bc8520405b1d90daea66d17a62c415f0edee8ecb4a52b9ed1460bc6a5a9b0499"


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
                "api_version": {"const": "tekmovela.io/v1"},
                "kind": {"const": "AcceptanceCriterion"},
            },
            "required": ["id", "digest", "testable"],
        })
        with self.assertRaises(ContractError):
            schema.validate({"api_version": "tekmovela.io/v1", "kind": "Other", "id": "x"})


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

    def test_golden_digests_match_other_sdks(self):
        b = EvidenceBundle("bundle.e0001", "run.1", "checkout.atomicity@v1", "checkout.double_spend@v1")
        b.add_item(Item(kind=TRACE, uri="trace.1", digest="sha256:" + "a" * 64))
        self.assertEqual(b.digest(), "sha256:a8ca24195e5dba0cc990b696dd54259243c0c1bab095eab66ea5865b766537ca")
        b.environment_digest = "sha256:" + "d" * 64
        self.assertEqual(b.digest(), "sha256:20e7916b6b50cb2d81489db84eb5331255de068e2e3551538f908760676162d0")


class HarnessTests(unittest.TestCase):
    def _sample(self, seed=None):
        return HarnessVersion(
            id="checkout.double_spend",
            version="v1",
            verification_contract_ref=Reference("VerificationContract", "checkout.atomicity", "v1", DIGEST),
            component_lock={
                "runner": component("runner", "runner.deterministic", "v1", "sha256:" + "c" * 64),
                "adapter": component("adapter", "adapter.praxovela", "v1", "sha256:" + "d" * 64),
                "oracle": component("oracle", "oracle.effect_ledger", "v1", "sha256:" + "e" * 64),
                "reporter": component("reporter", "reporter.evidence", "v1", "sha256:" + "f" * 64),
            },
            evidence_schema="tekmovela.evidence-bundle.v1.schema.json",
            seed=seed,
        )

    def test_publish_digest(self):
        hv = self._sample()
        hv.publish()
        self.assertTrue(hv.digest)

    def test_golden_digest_matches_typescript_sdk(self):
        hv = self._sample()
        hv.publish()
        self.assertEqual(hv.digest, GOLDEN_HARNESS_VERSION_DIGEST)

    def test_body_emits_reference_object_and_isolation_profile(self):
        hv = self._sample()
        body = hv._body()
        self.assertEqual(body["verification_contract_ref"], {
            "kind": "VerificationContract",
            "id": "checkout.atomicity",
            "version": "v1",
            "digest": DIGEST,
        })
        self.assertEqual(body["isolation_profile"], {})
        self.assertNotIn("environment_contract", body)

    def test_body_emits_environment_contract_seed_when_set(self):
        hv = self._sample(seed="s1")
        self.assertEqual(hv._body()["environment_contract"], {"seed": "s1"})

    def test_rejects_malformed_verification_contract_ref(self):
        with self.assertRaises(ContractError):
            HarnessVersion(
                id="x", version="v1",
                verification_contract_ref={"kind": "VerificationContract", "id": "c", "version": "v1", "digest": "nope"},
                component_lock={"runner": component("runner", "r", "v1", DIGEST)},
                evidence_schema="s",
            )

    def test_incomplete_lock(self):
        with self.assertRaises(ContractError):
            HarnessVersion(
                id="x", version="v1",
                verification_contract_ref=Reference("VerificationContract", "c", "v1", DIGEST),
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

    def test_execute_binds_run_ref_to_runtime_handle(self):
        runner = LocalRunner(FakeRuntime())
        scenario = Scenario(id="scenario.double_spend", version="v1", verification_contract_ref="checkout.atomicity", seed="s1")
        result = runner.execute(scenario, sut_ref="cap/v2", run_id="run.e0001")
        self.assertEqual(result.bundle.run_ref, "scenario.double_spend/run-s1")
        self.assertNotEqual(result.bundle.run_ref, "run.e0001")
        self.assertEqual(result.run_id, result.bundle.run_ref)

    def test_validate_contract_fixture(self):
        with tempfile.TemporaryDirectory() as tmp:
            schema_path = os.path.join(tmp, "schema.json")
            fixture_path = os.path.join(tmp, "fixture.json")
            with open(schema_path, "w", encoding="utf-8") as fh:
                json.dump({"properties": {"kind": {"const": "AcceptanceCriterion"}, "api_version": {"const": "tekmovela.io/v1"}}, "required": ["kind", "id"]}, fh)
            with open(fixture_path, "w", encoding="utf-8") as fh:
                json.dump({"kind": "AcceptanceCriterion", "api_version": "tekmovela.io/v1", "id": "checkout.atomicity"}, fh)
            with open(schema_path, "r", encoding="utf-8") as fh:
                schema_doc = json.load(fh)
            validate_fixture(Schema(schema_doc), fixture_path)


class SchemaConformanceTests(unittest.TestCase):
    """Wire forms must satisfy the mirrored contract JSON Schemas under contracts/."""

    _REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(os.path.abspath(__file__)), "..", "..", ".."))
    EVIDENCE_BUNDLE_SCHEMA = os.path.join(_REPO_ROOT, "contracts", "evidence", "tekmovela.evidence-bundle.v1.schema.json")
    HARNESS_VERSION_SCHEMA = os.path.join(_REPO_ROOT, "contracts", "harness", "tekmovela.harness-version.v1.schema.json")

    def _load_schema(self, path):
        self.assertTrue(os.path.isfile(path), f"schema not found: {path}")
        with open(path, "r", encoding="utf-8") as fh:
            return Schema(json.load(fh))

    def test_sealed_evidence_bundle_satisfies_schema(self):
        b = EvidenceBundle("bundle.conformance", "run.conformance", "checkout.atomicity@v1", "checkout.double_spend@v1")
        b.add_item(Item(kind=TRACE, uri="local://trace/exec.jsonl", digest=DIGEST))
        b.add_item(Item(kind=EFFECT, uri="local://effects/ledger.jsonl", digest=DIGEST))
        b.seal()
        wire = {**b._body(), "bundle_digest": b.digest()}
        schema = self._load_schema(self.EVIDENCE_BUNDLE_SCHEMA)
        self.assertIn("bundle_digest", schema.required)
        parse_digest(wire["bundle_digest"])
        schema.validate(wire)

    def test_harness_version_body_satisfies_schema(self):
        hv = HarnessTests._sample(self)
        schema = self._load_schema(self.HARNESS_VERSION_SCHEMA)
        schema.validate(hv._body())


if __name__ == "__main__":
    unittest.main()
