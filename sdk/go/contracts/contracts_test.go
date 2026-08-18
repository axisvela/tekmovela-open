package contracts

import "testing"

func TestGoldenDigest(t *testing.T) {
	d, err := ComputeDigest(map[string]any{"a": 1, "b": 2})
	if err != nil {
		t.Fatal(err)
	}
	const want = "sha256:43258cff783fe7036d8a43033f830adfc60ec037382473548ac742b888292777"
	if d != want {
		t.Fatalf("digest %q != golden %q", d, want)
	}
}

func TestDigestStableAcrossKeyOrder(t *testing.T) {
	a, _ := ComputeDigest(map[string]any{"a": 1, "b": 2})
	b, _ := ComputeDigest(map[string]any{"b": 2, "a": 1})
	if a != b {
		t.Fatalf("digests must be key-order independent: %q vs %q", a, b)
	}
}

func TestParseDigest(t *testing.T) {
	good := "sha256:" + repeatHex('a', 64)
	if _, err := ParseDigest(good); err != nil {
		t.Fatalf("good digest rejected: %v", err)
	}
	for _, bad := range []string{"", "md5:abc", "sha256:xyz", "sha256:" + repeatHex('A', 64)} {
		if _, err := ParseDigest(bad); err == nil {
			t.Fatalf("bad digest %q accepted", bad)
		}
	}
}

func TestNewReferenceValidates(t *testing.T) {
	if _, err := NewReference("EvidenceBundle", "b.1", "v1", "sha256:"+repeatHex('b', 64)); err != nil {
		t.Fatal(err)
	}
	if _, err := NewReference("", "b.1", "v1", "sha256:"+repeatHex('b', 64)); err == nil {
		t.Fatal("empty kind must fail")
	}
	if _, err := NewReference("EvidenceBundle", "b.1", "v1", "nope"); err == nil {
		t.Fatal("bad digest must fail")
	}
}

func repeatHex(c byte, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = c
	}
	return string(out)
}
