package reference

import "testing"

func TestParseDigest(t *testing.T) {
	good := "sha256:0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	if _, err := ParseDigest(good); err != nil {
		t.Fatalf("good digest rejected: %v", err)
	}
	for _, bad := range []string{"", "md5:0123456789abcdef", "sha256:01234567", "SHA256:" + "a", "sha256:" + "A"} {
		if _, err := ParseDigest(bad); err == nil {
			t.Fatalf("bad digest %q accepted", bad)
		}
	}
}

func TestNewReferenceRequiresAllFields(t *testing.T) {
	if _, err := NewReference("", "a", "v1", "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"); err == nil {
		t.Fatal("empty kind must fail")
	}
	if _, err := NewReference("EvidenceBundle", "run.001", "v1", "md5:deadbeef"); err == nil {
		t.Fatal("bad digest must fail")
	}
}

func TestReferenceValidate(t *testing.T) {
	ref := Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"}
	if err := ref.Validate(); err != nil {
		t.Fatal(err)
	}
	bad := ref
	bad.Digest = "nope"
	if err := bad.Validate(); err == nil {
		t.Fatal("invalid digest must fail validation")
	}
}

func TestVerifyHex(t *testing.T) {
	if !VerifyHex("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") {
		t.Fatal("64 hex chars must verify")
	}
	if VerifyHex("zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz") {
		t.Fatal("non-hex must fail")
	}
}
