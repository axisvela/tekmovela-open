// Package contracts provides canonical-JSON digest helpers and validated
// references for TEKMOVELA SDK consumers. Digests are sha256:<hex> over
// canonical JSON (map keys sorted, compact separators), matching the Python and
// TypeScript SDKs so evidence sealed in one SDK verifies in the others.
//
// Digestable-input contract: values are JSON strings, integers, booleans,
// null, arrays, or objects. Floating-point values are not canonicalized and
// must not appear in digestible documents (Go formats 1.0 as "1", Python as
// "1.0").
package contracts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
)

var sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// API_VERSION is the TEKMOVELA API version carried by published objects.
const API_VERSION = "tekmovela.io/v1alpha1"

// Reference points to a published TEKMOVELA object.
type Reference struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

// ParseDigest validates a digest of the form sha256:<64 hex>.
func ParseDigest(s string) (string, error) {
	if len(s) != 7+64 || s[:7] != "sha256:" || !sha256Hex.MatchString(s[7:]) {
		return "", fmt.Errorf("malformed digest %q: want sha256:<64 hex>", s)
	}
	return s, nil
}

// NewReference builds a validated reference.
func NewReference(kind, id, version, digest string) (Reference, error) {
	if kind == "" || id == "" || version == "" {
		return Reference{}, fmt.Errorf("kind, id, and version are required")
	}
	d, err := ParseDigest(digest)
	if err != nil {
		return Reference{}, err
	}
	return Reference{Kind: kind, ID: id, Version: version, Digest: d}, nil
}

// CanonicalJSON marshals v with Go's json.Marshal, which sorts map keys
// alphabetically and emits compact separators — matching Python's
// sort_keys=True and the TS key-sort.
func CanonicalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}

// ComputeDigest returns sha256:<hex> over the canonical JSON of v.
func ComputeDigest(v any) (string, error) {
	b, err := CanonicalJSON(v)
	if err != nil {
		return "", fmt.Errorf("canonical json: %w", err)
	}
	sum := sha256.Sum256(b)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}
