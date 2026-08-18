package evidence

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestBundleSealAndVerify(t *testing.T) {
	b := NewBundle("bundle.e0001", "run.e0001", "checkout.atomicity@v1", "checkout.double_spend@v1")
	if err := b.AddItem(Item{Kind: ItemKindTrace, URI: "trace.1", Digest: "sha256:" + repeat('a', 64)}); err != nil {
		t.Fatal(err)
	}
	if err := b.Seal(); err != nil {
		t.Fatal(err)
	}
	if b.BundleDigest == "" {
		t.Fatal("sealed bundle must carry a digest")
	}
	if err := b.Verify(); err != nil {
		t.Fatal(err)
	}
}

func TestBundleRejectsTampering(t *testing.T) {
	b := NewBundle("bundle.e0002", "run.e0002", "checkout.atomicity@v1", "checkout.double_spend@v1")
	_ = b.AddItem(Item{Kind: ItemKindTrace, URI: "trace.2", Digest: "sha256:" + repeat('b', 64)})
	_ = b.Seal()
	b.Items[0].URI = "tampered"
	if err := b.Verify(); err == nil {
		t.Fatal("tampered bundle must not verify")
	}
}

func TestBundleCannotSealEmpty(t *testing.T) {
	b := NewBundle("bundle.e0003", "run.e0003", "checkout.atomicity@v1", "checkout.double_spend@v1")
	if err := b.Seal(); err == nil {
		t.Fatal("empty bundle must not seal")
	}
}

func TestMissingEvidence(t *testing.T) {
	b := NewBundle("bundle.e0004", "run.e0004", "checkout.atomicity@v1", "checkout.double_spend@v1")
	if err := b.AddMissing("local://authority.jsonl", "sha256:"+repeat('c', 64), "authority perspective not captured"); err != nil {
		t.Fatal(err)
	}
	if err := b.Seal(); err != nil {
		t.Fatal(err)
	}
	if len(b.Missing()) != 1 {
		t.Fatal("missing evidence must be recorded")
	}
}

func TestGoldenDigests(t *testing.T) {
	b := NewBundle("bundle.e0001", "run.1", "checkout.atomicity@v1", "checkout.double_spend@v1")
	_ = b.AddItem(Item{Kind: ItemKindTrace, URI: "trace.1", Digest: "sha256:" + repeat('a', 64)})
	d, err := b.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if d != "sha256:18b4223c8528f03c945964fe0a6590dafcefc77c377034a46bfc3374b4aee470" {
		t.Fatalf("unset-env digest %q != golden", d)
	}
	b.EnvironmentDigest = "sha256:" + repeat('d', 64)
	d, err = b.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if d != "sha256:88e88a573fe31224487abe800f920abf41b31527ba7d5045b86255d8e3fad6e6" {
		t.Fatalf("set-env digest %q != golden", d)
	}
}

func TestScopeRejectsUnknownKeys(t *testing.T) {
	b := NewBundle("bundle.e0005", "run.e0005", "checkout.atomicity@v1", "checkout.double_spend@v1")
	if err := b.SetScope(map[string]string{"intent_ref": "scenario.x", "unknown": "y"}); err == nil {
		t.Fatal("unknown scope key must be rejected")
	}
}

func TestToJSONSealedWireForm(t *testing.T) {
	b := NewBundle("bundle.e0011", "run.11", "checkout.atomicity@v1", "checkout.double_spend@v1")
	_ = b.AddItem(Item{Kind: ItemKindTrace, URI: "trace.1", Digest: "sha256:" + repeat('a', 64)})
	if err := b.Seal(); err != nil {
		t.Fatal(err)
	}
	d, err := b.Digest()
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := b.ToJSON()["bundle_digest"]; !ok || got != d {
		t.Fatalf("sealed ToJSON bundle_digest %v != digest %q", got, d)
	}
	if _, ok := b.Body()["bundle_digest"]; ok {
		t.Fatal("Body() must never contain bundle_digest")
	}
}

func TestToJSONUnsealedOmitsBundleDigest(t *testing.T) {
	b := NewBundle("bundle.e0012", "run.12", "checkout.atomicity@v1", "checkout.double_spend@v1")
	_ = b.AddItem(Item{Kind: ItemKindTrace, URI: "trace.1", Digest: "sha256:" + repeat('a', 64)})
	if _, ok := b.ToJSON()["bundle_digest"]; ok {
		t.Fatal("unsealed ToJSON must omit bundle_digest")
	}
}

func TestNilScopeCanonicalParity(t *testing.T) {
	b := &EvidenceBundle{
		ID:                      "bundle.e0001",
		RunRef:                  "run.1",
		VerificationContractRef: "checkout.atomicity@v1",
		HarnessVersionRef:       "checkout.double_spend@v1",
		Items:                   []Item{{Kind: ItemKindTrace, URI: "trace.1", Digest: "sha256:" + repeat('a', 64)}},
	}
	d, err := b.Digest()
	if err != nil {
		t.Fatal(err)
	}
	const want = "sha256:18b4223c8528f03c945964fe0a6590dafcefc77c377034a46bfc3374b4aee470"
	if d != want {
		t.Fatalf("nil-scope digest %q != golden %q", d, want)
	}
	if _, ok := b.Body()["scope"].(map[string]string); !ok {
		t.Fatalf("Body() scope must be a map, got %T", b.Body()["scope"])
	}
}

func TestSetScopeNilNormalizesEmpty(t *testing.T) {
	b := NewBundle("bundle.e0013", "run.13", "checkout.atomicity@v1", "checkout.double_spend@v1")
	if err := b.SetScope(nil); err != nil {
		t.Fatal(err)
	}
	if b.Scope == nil {
		t.Fatal("SetScope(nil) must set an empty map, not nil")
	}
}

func TestAddItemRejectsUnknownKind(t *testing.T) {
	b := NewBundle("bundle.e0014", "run.14", "checkout.atomicity@v1", "checkout.double_spend@v1")
	if err := b.AddItem(Item{Kind: "bogus", URI: "t", Digest: "sha256:" + repeat('a', 64)}); err == nil {
		t.Fatal("unknown item kind must be rejected")
	}
}

func TestMissingReturnsEmptySlice(t *testing.T) {
	b := NewBundle("bundle.e0015", "run.15", "checkout.atomicity@v1", "checkout.double_spend@v1")
	_ = b.AddItem(Item{Kind: ItemKindTrace, URI: "trace.1", Digest: "sha256:" + repeat('a', 64)})
	if got := b.Missing(); got == nil || len(got) != 0 {
		t.Fatalf("Missing() must return a non-nil empty slice, got %#v", got)
	}
}

func TestMarshalJSONWireForm(t *testing.T) {
	b := NewBundle("bundle.e0016", "run.16", "checkout.atomicity@v1", "checkout.double_spend@v1")
	_ = b.AddItem(Item{Kind: ItemKindTrace, URI: "trace.1", Digest: "sha256:" + repeat('a', 64)})
	if err := b.Seal(); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	if m["kind"] != "EvidenceBundle" || m["id"] != "bundle.e0016" || m["run_ref"] != "run.16" {
		t.Fatalf("wire form must use snake_case keys, got %s", raw)
	}
	if m["bundle_digest"] != b.BundleDigest {
		t.Fatalf("wire form must include bundle_digest when sealed, got %s", raw)
	}
}

func repeat(c byte, n int) string {
	return strings.Repeat(string(c), n)
}
