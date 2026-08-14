// Package reference implements TEKMOVELA reference and digest rules for SDK
// consumers. Every reference to a published object carries a sha256 digest so
// evidence, harnesses, and gates bind to an exact artifact.
package reference

import (
	"encoding/hex"
	"fmt"
	"regexp"
)

var sha256Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Reference points to a published TEKMOVELA object.
type Reference struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

// ParseDigest validates a digest string of the form "sha256:<hex>".
func ParseDigest(s string) (string, error) {
	if len(s) != 7+64 || s[:7] != "sha256:" || !sha256Hex.MatchString(s[7:]) {
		return "", fmt.Errorf("malformed digest %q: want sha256:<64 hex>", s)
	}
	return s, nil
}

// NewReference builds a validated reference. version and digest are required.
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

// Validate reports whether every digest field in the reference is well-formed.
func (r Reference) Validate() error {
	if _, err := ParseDigest(r.Digest); err != nil {
		return fmt.Errorf("reference %s/%s@%s: %w", r.Kind, r.ID, r.Version, err)
	}
	return nil
}

// VerifyHex checks a raw hex string without the sha256: prefix (e.g. for
// artifact content digests stored separately).
func VerifyHex(s string) bool {
	if len(s) != 64 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}
