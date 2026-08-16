"""Contract and digest helpers.

Digests are ``sha256:<hex>`` over canonical JSON. Refs to published objects
carry ``{kind, id, version, digest}``.
"""

from __future__ import annotations

import hashlib
import json
import re
from typing import Any, Dict, Mapping, Optional

API_VERSION = "tekmovela.io/v1alpha1"
_SHA256 = re.compile(r"^sha256:[0-9a-f]{64}$")


class ContractError(ValueError):
    """Raised when a document violates a TEKMOVELA contract."""


def parse_digest(digest: str) -> None:
    if not _SHA256.match(digest):
        raise ContractError(f"malformed digest {digest!r}: want sha256:<64 hex>")


def canonical_json(value: Any) -> bytes:
    return json.dumps(value, sort_keys=False, separators=(",", ":"), ensure_ascii=False).encode("utf-8")


def digest_of(value: Any) -> str:
    return "sha256:" + hashlib.sha256(canonical_json(value)).hexdigest()


def digest_from_bytes(data: bytes) -> str:
    return "sha256:" + hashlib.sha256(data).hexdigest()


class Reference(dict):
    """A validated {kind, id, version, digest} reference."""

    def __init__(self, kind: str, id: str, version: str, digest: str):
        if not (kind and id and version):
            raise ContractError("reference kind, id, and version are required")
        parse_digest(digest)
        super().__init__(kind=kind, id=id, version=version, digest=digest)


class Schema:
    """A minimal structural contract validator.

    The authoritative conformance is the JSON Schema under ``contracts/``; this
    lightweight validator covers the shared shape (api_version/kind const,
    required fields, digest patterns) without external dependencies.
    """

    def __init__(self, doc: Mapping[str, Any]):
        self.doc = doc
        self.properties: Mapping[str, Any] = doc.get("properties", {})
        self.required: list[str] = doc.get("required", [])
        self._kind = _const_of(self.properties.get("kind"))
        self._api_version = _const_of(self.properties.get("api_version"))

    @property
    def kind(self) -> str:
        assert self._kind is not None
        return self._kind

    @property
    def api_version(self) -> str:
        assert self._api_version is not None
        return self._api_version

    def validate(self, instance: Mapping[str, Any]) -> None:
        if instance.get("kind") != self._kind:
            raise ContractError(f"kind must be {self._kind!r}")
        if instance.get("api_version") != self._api_version:
            raise ContractError(f"api_version must be {self._api_version!r}")
        for field in self.required:
            if field not in instance:
                raise ContractError(f"missing required field {field!r}")
        self._validate_digests(instance)

    def _validate_digests(self, instance: Mapping[str, Any]) -> None:
        if "digest" in instance:
            parse_digest(instance["digest"])
        for value in instance.values():
            if isinstance(value, dict):
                if "digest" in value and value.get("kind") and value.get("version"):
                    parse_digest(value["digest"])
                self._validate_digests(value)
            elif isinstance(value, list):
                for item in value:
                    if isinstance(item, dict):
                        self._validate_digests(item)


def _const_of(spec: Optional[Mapping[str, Any]]) -> Optional[str]:
    if not spec:
        return None
    if "const" in spec:
        return spec["const"]
    if "enum" in spec:
        return spec["enum"][0]
    return None


def load_schema(path: str) -> Schema:
    with open(path, "r", encoding="utf-8") as fh:
        return Schema(json.load(fh))


def validate_fixture(schema: Schema, instance_path: str) -> None:
    with open(instance_path, "r", encoding="utf-8") as fh:
        schema.validate(json.load(fh))
