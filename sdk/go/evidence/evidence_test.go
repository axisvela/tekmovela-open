package evidence

import (
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

func repeat(c byte, n int) string {
	return strings.Repeat(string(c), n)
}
