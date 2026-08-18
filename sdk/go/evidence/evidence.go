// Package evidence provides a digest-sealed EvidenceBundle for TEKMOVELA SDK
// consumers. Raw telemetry is not evidence; only records bound to versions,
// sources, digests, scope, and claims become evidence. The canonical form and
// wire form match the Python and TypeScript SDKs (missing evidence is itself
// evidence; environment_digest is omitted when unset; bundle_digest appears in
// the sealed wire form but not in the canonical digest body).
package evidence

import (
	"encoding/json"
	"fmt"

	"github.com/axisrobo/tekmovela-open/sdk/go/contracts"
)

// ItemKind is the kind of evidence item.
type ItemKind string

const (
	ItemKindTrace       ItemKind = "trace"
	ItemKindEffect      ItemKind = "effect"
	ItemKindFailure     ItemKind = "failure"
	ItemKindAttribution ItemKind = "attribution"
	ItemKindExplanation ItemKind = "explanation"
	ItemKindAuthority   ItemKind = "authority"
	ItemKindMissing     ItemKind = "missing"
)

// Item is a single evidence item.
type Item struct {
	Kind   ItemKind `json:"kind"`
	URI    string   `json:"uri"`
	Digest string   `json:"digest"`
	Note   string   `json:"note,omitempty"`
}

// allowedScopeKeys are the schema-constrained scope keys.
var allowedScopeKeys = map[string]bool{
	"intent_ref":     true,
	"capability_ref": true,
	"time_window":    true,
}

// EvidenceBundle is a portable, digest-verifiable collection of evidence.
type EvidenceBundle struct {
	ID                      string            `json:"id"`
	RunRef                  string            `json:"run_ref"`
	VerificationContractRef string            `json:"verification_contract_ref"`
	HarnessVersionRef       string            `json:"harness_version_ref"`
	Items                   []Item            `json:"items"`
	Scope                   map[string]string `json:"scope"`
	EnvironmentDigest       string            `json:"environment_digest,omitempty"`
	BundleDigest            string            `json:"bundle_digest,omitempty"`
}

// NewBundle builds an empty evidence bundle.
func NewBundle(id, runRef, verificationContractRef, harnessVersionRef string) *EvidenceBundle {
	return &EvidenceBundle{
		ID:                      id,
		RunRef:                  runRef,
		VerificationContractRef: verificationContractRef,
		HarnessVersionRef:       harnessVersionRef,
		Items:                   []Item{},
		Scope:                   map[string]string{},
	}
}

// SetScope replaces the scope, rejecting unknown keys (schema-constrained).
// A nil scope normalizes to an empty map so the canonical form stays `{}`.
func (b *EvidenceBundle) SetScope(scope map[string]string) error {
	for k := range scope {
		if !allowedScopeKeys[k] {
			return fmt.Errorf("unknown scope key %q (allowed: intent_ref, capability_ref, time_window)", k)
		}
	}
	if scope == nil {
		scope = map[string]string{}
	}
	b.Scope = scope
	return nil
}

// AddItem validates and appends an evidence item.
func (b *EvidenceBundle) AddItem(item Item) error {
	switch item.Kind {
	case ItemKindTrace, ItemKindEffect, ItemKindFailure, ItemKindAttribution, ItemKindExplanation, ItemKindAuthority, ItemKindMissing:
	default:
		return fmt.Errorf("unknown item kind %q", item.Kind)
	}
	if item.URI == "" {
		return fmt.Errorf("item uri is required")
	}
	if _, err := contracts.ParseDigest(item.Digest); err != nil {
		return err
	}
	b.Items = append(b.Items, item)
	return nil
}

// AddMissing records that evidence is missing — missing evidence is evidence.
func (b *EvidenceBundle) AddMissing(uri, digest, note string) error {
	return b.AddItem(Item{Kind: ItemKindMissing, URI: uri, Digest: digest, Note: note})
}

// Body returns the canonical form (without bundle_digest; environment_digest
// omitted when unset). A nil scope normalizes to an empty map so the canonical
// form stays `{}`, matching the Python and TypeScript SDKs. Used for digest
// computation.
func (b *EvidenceBundle) Body() map[string]any {
	scope := b.Scope
	if scope == nil {
		scope = map[string]string{}
	}
	body := map[string]any{
		"api_version":               contracts.API_VERSION,
		"kind":                      "EvidenceBundle",
		"id":                        b.ID,
		"run_ref":                   b.RunRef,
		"verification_contract_ref": b.VerificationContractRef,
		"harness_version_ref":       b.HarnessVersionRef,
		"scope":                     scope,
		"items":                     b.itemsJSON(),
	}
	if b.EnvironmentDigest != "" {
		body["environment_digest"] = b.EnvironmentDigest
	}
	return body
}

func (b *EvidenceBundle) itemsJSON() []map[string]any {
	out := make([]map[string]any, 0, len(b.Items))
	for _, i := range b.Items {
		item := map[string]any{"kind": string(i.Kind), "uri": i.URI, "digest": i.Digest}
		if i.Note != "" {
			item["note"] = i.Note
		}
		out = append(out, item)
	}
	return out
}

// Digest computes the canonical digest.
func (b *EvidenceBundle) Digest() (string, error) {
	return contracts.ComputeDigest(b.Body())
}

// Seal binds the bundle digest; a bundle must have at least one item.
func (b *EvidenceBundle) Seal() error {
	if len(b.Items) == 0 {
		return fmt.Errorf("bundle %s cannot seal with no items", b.ID)
	}
	d, err := b.Digest()
	if err != nil {
		return err
	}
	b.BundleDigest = d
	return nil
}

// Verify rejects tampering or an unsealed bundle.
func (b *EvidenceBundle) Verify() error {
	for _, item := range b.Items {
		if _, err := contracts.ParseDigest(item.Digest); err != nil {
			return err
		}
	}
	if b.BundleDigest == "" {
		return fmt.Errorf("bundle %s is not sealed", b.ID)
	}
	d, err := b.Digest()
	if err != nil {
		return err
	}
	if d != b.BundleDigest {
		return fmt.Errorf("bundle %s digest mismatch", b.ID)
	}
	return nil
}

// Missing returns the missing-evidence items.
func (b *EvidenceBundle) Missing() []Item {
	out := []Item{}
	for _, i := range b.Items {
		if i.Kind == ItemKindMissing {
			out = append(out, i)
		}
	}
	return out
}

// ToJSON returns the schema-valid sealed wire form (includes bundle_digest).
func (b *EvidenceBundle) ToJSON() map[string]any {
	body := b.Body()
	if b.BundleDigest != "" {
		body["bundle_digest"] = b.BundleDigest
	}
	return body
}

// MarshalJSON emits the sealed wire form so direct json.Marshal calls produce
// the schema-valid snake_case document.
func (b *EvidenceBundle) MarshalJSON() ([]byte, error) {
	return json.Marshal(b.ToJSON())
}
