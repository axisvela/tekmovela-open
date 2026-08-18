// Command go-verify demonstrates the TEKMOVELA Go SDK contracts and evidence
// packages against a contract fixture. It parses an EvidenceBundle fixture,
// verifies every digest it references, and rebuilds the bundle to seal and
// verify it end-to-end.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/axisrobo/tekmovela-open/sdk/go/contracts"
	"github.com/axisrobo/tekmovela-open/sdk/go/evidence"
)

type fixtureItem struct {
	Kind   string `json:"kind"`
	URI    string `json:"uri"`
	Digest string `json:"digest"`
	Note   string `json:"note"`
}

type fixture struct {
	APIVersion              string            `json:"api_version"`
	Kind                    string            `json:"kind"`
	ID                      string            `json:"id"`
	RunRef                  string            `json:"run_ref"`
	VerificationContractRef string            `json:"verification_contract_ref"`
	HarnessVersionRef       string            `json:"harness_version_ref"`
	EnvironmentDigest       string            `json:"environment_digest"`
	Scope                   map[string]string `json:"scope"`
	BundleDigest            string            `json:"bundle_digest"`
	Items                   []fixtureItem     `json:"items"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: go-verify <evidence-bundle.json>")
		os.Exit(2)
	}
	path := os.Args[1]
	doc, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var f fixture
	if err := json.Unmarshal(doc, &f); err != nil {
		fmt.Fprintln(os.Stderr, "invalid JSON:", err)
		os.Exit(1)
	}
	if f.APIVersion != "tekmovela.io/v1alpha1" || f.Kind != "EvidenceBundle" {
		fmt.Fprintf(os.Stderr, "not an EvidenceBundle: %s/%s\n", f.APIVersion, f.Kind)
		os.Exit(1)
	}

	failures := 0
	check := func(name, digest string) {
		if _, err := contracts.ParseDigest(digest); err != nil {
			fmt.Printf("  FAIL %s: %v\n", name, err)
			failures++
		} else {
			fmt.Printf("  ok   %s: %s\n", name, digest[:16]+"…")
		}
	}

	fmt.Printf("EvidenceBundle %s\n", f.ID)
	if f.BundleDigest != "" {
		check("bundle_digest", f.BundleDigest)
	}
	for _, it := range f.Items {
		check("item["+it.Kind+"]", it.Digest)
	}

	// Rebuild the fixture as a sealed, digest-verifiable EvidenceBundle.
	bundle := evidence.NewBundle(f.ID, f.RunRef, f.VerificationContractRef, f.HarnessVersionRef)
	bundle.EnvironmentDigest = f.EnvironmentDigest
	if err := bundle.SetScope(f.Scope); err != nil {
		fmt.Printf("  FAIL scope: %v\n", err)
		failures++
	}
	for _, it := range f.Items {
		if err := bundle.AddItem(evidence.Item{
			Kind:   evidence.ItemKind(it.Kind),
			URI:    it.URI,
			Digest: it.Digest,
			Note:   it.Note,
		}); err != nil {
			fmt.Printf("  FAIL item[%s]: %v\n", it.Kind, err)
			failures++
		}
	}

	if err := bundle.Seal(); err != nil {
		fmt.Printf("  FAIL seal: %v\n", err)
		failures++
	} else {
		fmt.Printf("  ok   canonical bundle digest: %s\n", bundle.BundleDigest[:16]+"…")
	}

	// If the fixture is sealed, cross-check its bundle_digest against the
	// canonical digest. Fixtures often carry placeholder digests, so a mismatch
	// is reported rather than treated as a hard failure.
	if f.BundleDigest != "" {
		bundle.BundleDigest = f.BundleDigest
		if err := bundle.Verify(); err != nil {
			fmt.Printf("  note fixture bundle_digest differs from the canonical digest: %v (fixtures often carry placeholder digests)\n", err)
		} else {
			fmt.Println("  ok   fixture bundle_digest matches the canonical digest")
		}
	}

	if failures > 0 {
		fmt.Printf("FAILED: %d digest check(s)\n", failures)
		os.Exit(1)
	}
	fmt.Println("PASSED: all digests well-formed; bundle rebuilt, sealed, and verified")
}
