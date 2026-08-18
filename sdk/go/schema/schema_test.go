// Package schema asserts that the Go SDK's emitted wire forms satisfy the
// mirrored TEKMOVELA JSON Schemas. The Go SDK is stdlib-only (no JSON Schema
// validator), so conformance is checked with structural assertions against the
// schema-required keys and digest patterns, reading the mirrored schemas from
// the repository contracts/ directory so a schema change fails the test.
package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/axisrobo/tekmovela-open/sdk/go/contracts"
	"github.com/axisrobo/tekmovela-open/sdk/go/evidence"
	"github.com/axisrobo/tekmovela-open/sdk/go/harness"
)

const (
	evidenceBundleSchema = "evidence/tekmovela.evidence-bundle.v1alpha1.schema.json"
	harnessVersionSchema = "harness/tekmovela.harness-version.v1alpha1.schema.json"
)

var sha256Pattern = regexp.MustCompile(`^sha256:[0-9a-f]{64}$`)

// schemaRoot returns the repository contracts/ directory (where the mirrored
// schemas live), found by walking up from the test package directory. The walk
// skips the Go contracts package directory, which has no evidence/harness
// schema subdirectories.
func schemaRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		candidate := filepath.Join(dir, "contracts")
		if schemaSubdir(candidate, "evidence") && schemaSubdir(candidate, "harness") {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("contracts/ with evidence and harness schemas not found above %s", dir)
		}
		dir = parent
	}
}

func schemaSubdir(root, name string) bool {
	st, err := os.Stat(filepath.Join(root, name))
	return err == nil && st.IsDir()
}

// decodeSchema loads and decodes a JSON Schema document.
func decodeSchema(path string) (map[string]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("schema %s: %w", path, err)
	}
	return doc, nil
}

// schemaRequired returns the top-level required array of the schema at path.
func schemaRequired(path string) ([]string, error) {
	doc, err := decodeSchema(path)
	if err != nil {
		return nil, err
	}
	return schemaRequiredAt(path, doc)
}

// schemaRequiredAt returns the required array found by descending keys
// (e.g. properties/items/items) from the schema root.
func schemaRequiredAt(path string, doc map[string]any, keys ...string) ([]string, error) {
	cur, err := descend(path, doc, keys...)
	if err != nil {
		return nil, err
	}
	m, ok := cur.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema %s: %v does not resolve to an object", path, keys)
	}
	raw, ok := m["required"].([]any)
	if !ok {
		return nil, fmt.Errorf("schema %s: no required array at %v", path, keys)
	}
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		s, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("schema %s: non-string entry in required at %v", path, keys)
		}
		out = append(out, s)
	}
	return out, nil
}

// schemaObjectKeysAt returns the property keys of the object found by
// descending keys (e.g. properties/scope/properties), sorted for determinism.
func schemaObjectKeysAt(path string, doc map[string]any, keys ...string) ([]string, error) {
	cur, err := descend(path, doc, keys...)
	if err != nil {
		return nil, err
	}
	m, ok := cur.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("schema %s: %v does not resolve to an object", path, keys)
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out, nil
}

func descend(path string, doc map[string]any, keys ...string) (any, error) {
	var cur any = doc
	for _, k := range keys {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("schema %s: path %v breaks at %q (not an object)", path, keys, k)
		}
		cur, ok = m[k]
		if !ok {
			return nil, fmt.Errorf("schema %s: missing key %q on path %v", path, keys, k)
		}
	}
	return cur, nil
}

// wireForm returns v marshaled and unmarshaled as a generic JSON object,
// exercising the JSON wire encoding and normalizing nested types.
func wireForm(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func assertRequired(t *testing.T, object string, form map[string]any, required []string) {
	t.Helper()
	for _, key := range required {
		if _, ok := form[key]; !ok {
			t.Errorf("%s wire form missing schema-required field %q", object, key)
		}
	}
}

func TestEvidenceBundleWireForm(t *testing.T) {
	root, err := schemaRoot()
	if err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(root, evidenceBundleSchema)
	required, err := schemaRequired(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := decodeSchema(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	itemRequired, err := schemaRequiredAt(schemaPath, doc, "properties", "items", "items")
	if err != nil {
		t.Fatal(err)
	}
	scopeAllowed, err := schemaObjectKeysAt(schemaPath, doc, "properties", "scope", "properties")
	if err != nil {
		t.Fatal(err)
	}

	b := evidence.NewBundle("bundle.conformance", "run.conformance", "checkout.atomicity@v1", "checkout.double_spend@v1")
	if err := b.AddItem(evidence.Item{Kind: evidence.ItemKindTrace, URI: "local://trace/exec.jsonl", Digest: "sha256:" + strings.Repeat("a", 64)}); err != nil {
		t.Fatal(err)
	}
	if err := b.SetScope(map[string]string{"intent_ref": "scenario.conformance", "capability_ref": "cap/v2"}); err != nil {
		t.Fatal(err)
	}
	if err := b.Seal(); err != nil {
		t.Fatal(err)
	}

	form := wireForm(t, b.ToJSON())
	assertRequired(t, "EvidenceBundle", form, required)
	if form["api_version"] != contracts.API_VERSION {
		t.Errorf("api_version = %v, want %q", form["api_version"], contracts.API_VERSION)
	}
	if form["kind"] != "EvidenceBundle" {
		t.Errorf("kind = %v, want %q", form["kind"], "EvidenceBundle")
	}

	items, ok := form["items"].([]any)
	if !ok {
		t.Fatalf("items has type %T, want []any", form["items"])
	}
	if len(items) == 0 {
		t.Fatal("items must be non-empty")
	}
	for i, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("items[%d] has type %T, want object", i, raw)
		}
		for _, key := range itemRequired {
			if _, ok := item[key]; !ok {
				t.Errorf("items[%d] missing schema-required field %q", i, key)
			}
		}
		if d, _ := item["digest"].(string); !sha256Pattern.MatchString(d) {
			t.Errorf("items[%d].digest %v does not match ^sha256:[0-9a-f]{64}$", i, item["digest"])
		}
	}

	if d, _ := form["bundle_digest"].(string); !sha256Pattern.MatchString(d) {
		t.Errorf("bundle_digest %v does not match ^sha256:[0-9a-f]{64}$", form["bundle_digest"])
	}

	if scope, ok := form["scope"].(map[string]any); ok {
		for key := range scope {
			if !slices.Contains(scopeAllowed, key) {
				t.Errorf("scope key %q not allowed by schema (allowed: %v)", key, scopeAllowed)
			}
		}
	}
}

func TestHarnessVersionWireForm(t *testing.T) {
	root, err := schemaRoot()
	if err != nil {
		t.Fatal(err)
	}
	schemaPath := filepath.Join(root, harnessVersionSchema)
	required, err := schemaRequired(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := decodeSchema(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	refRequired, err := schemaRequiredAt(schemaPath, doc, "$defs", "reference")
	if err != nil {
		t.Fatal(err)
	}
	lockRequired, err := schemaRequiredAt(schemaPath, doc, "properties", "component_lock")
	if err != nil {
		t.Fatal(err)
	}
	determinismRequired, err := schemaRequiredAt(schemaPath, doc, "properties", "determinism_profile")
	if err != nil {
		t.Fatal(err)
	}

	hv, err := harness.NewHarnessVersion(
		"checkout.double_spend",
		"v1",
		contracts.Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:" + strings.Repeat("a", 64)},
		harness.ComponentLock{
			harness.ComponentRunner:   {Kind: "HarnessComponent", ID: "runner.deterministic", Version: "v1", Digest: "sha256:" + strings.Repeat("b", 64)},
			harness.ComponentAdapter:  {Kind: "HarnessComponent", ID: "adapter.praxovela", Version: "v1", Digest: "sha256:" + strings.Repeat("c", 64)},
			harness.ComponentOracle:   {Kind: "HarnessComponent", ID: "oracle.effect_ledger", Version: "v1", Digest: "sha256:" + strings.Repeat("d", 64)},
			harness.ComponentReporter: {Kind: "HarnessComponent", ID: "reporter.evidence", Version: "v1", Digest: "sha256:" + strings.Repeat("e", 64)},
		},
		"tekmovela.evidence-bundle.v1alpha1.schema.json",
		harness.DeterminismDeterministic,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}

	form := wireForm(t, hv.Body())
	assertRequired(t, "HarnessVersion", form, required)
	if form["api_version"] != contracts.API_VERSION {
		t.Errorf("api_version = %v, want %q", form["api_version"], contracts.API_VERSION)
	}
	if form["kind"] != "HarnessVersion" {
		t.Errorf("kind = %v, want %q", form["kind"], "HarnessVersion")
	}

	assertRef := func(field string, ref any) {
		t.Helper()
		m, ok := ref.(map[string]any)
		if !ok {
			t.Errorf("%s has type %T, want a reference object", field, ref)
			return
		}
		for _, key := range refRequired {
			if _, ok := m[key]; !ok {
				t.Errorf("%s missing schema-required field %q", field, key)
			}
		}
		if d, _ := m["digest"].(string); !sha256Pattern.MatchString(d) {
			t.Errorf("%s.digest %v does not match ^sha256:[0-9a-f]{64}$", field, m["digest"])
		}
	}

	assertRef("verification_contract_ref", form["verification_contract_ref"])

	lock, ok := form["component_lock"].(map[string]any)
	if !ok {
		t.Fatalf("component_lock has type %T, want object", form["component_lock"])
	}
	for _, role := range lockRequired {
		ref, ok := lock[role]
		if !ok {
			t.Errorf("component_lock missing schema-required role %q", role)
			continue
		}
		assertRef("component_lock."+role, ref)
	}

	if dp, ok := form["determinism_profile"].(map[string]any); ok {
		for _, key := range determinismRequired {
			if _, ok := dp[key]; !ok {
				t.Errorf("determinism_profile missing schema-required field %q", key)
			}
		}
		if mode, _ := dp["mode"].(string); mode != string(hv.DeterminismMode) {
			t.Errorf("determinism_profile.mode = %q, want %q", mode, hv.DeterminismMode)
		}
	} else {
		t.Errorf("determinism_profile has type %T, want object", form["determinism_profile"])
	}
}
