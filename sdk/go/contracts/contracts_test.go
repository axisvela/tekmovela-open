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

func TestCanonicalJSONNoHTMLEscaping(t *testing.T) {
	got, err := CanonicalJSON(map[string]any{"x": "<a>&"})
	if err != nil {
		t.Fatal(err)
	}
	const want = `{"x":"<a>&"}`
	if string(got) != want {
		t.Fatalf("canonical json %q != %q", got, want)
	}
}

func TestGoldenDigestHTMLEscaping(t *testing.T) {
	d, err := ComputeDigest(map[string]any{"x": "<a>&"})
	if err != nil {
		t.Fatal(err)
	}
	const want = "sha256:bf934ce4dc8b33603214f2ff9f2270929a467392495e2b2b3cda082d2f2b990a"
	if d != want {
		t.Fatalf("digest %q != golden %q", d, want)
	}
}

func TestReferenceDigestParity(t *testing.T) {
	ref := Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:" + repeatHex('a', 64)}
	d, err := ComputeDigest(ref)
	if err != nil {
		t.Fatal(err)
	}
	const want = "sha256:380bc7efb07a479b54b2290c4f9f17cd7606c258d33754ae2a142add52a97438"
	if d != want {
		t.Fatalf("digest %q != golden %q", d, want)
	}
}

func TestReferenceValidate(t *testing.T) {
	ref := Reference{Kind: "VerificationContract", ID: "vc.1", Version: "v1", Digest: "sha256:" + repeatHex('c', 64)}
	if err := ref.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := ref
	bad.Digest = "nope"
	if err := bad.Validate(); err == nil {
		t.Fatal("invalid digest must fail validation")
	}
}

func repeatHex(c byte, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = c
	}
	return string(out)
}
