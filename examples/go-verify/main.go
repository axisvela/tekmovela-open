// Command go-verify demonstrates the TEKMOVELA Go SDK reference/digest rules
// against a contract fixture. It parses an EvidenceBundle fixture and verifies
// every digest it references.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/axisrobo/tekmovela-open/sdk/go/reference"
)

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

	var bundle struct {
		APIVersion string `json:"api_version"`
		Kind       string `json:"kind"`
		ID         string `json:"id"`
		Bundle     string `json:"bundle_digest"`
		Items      []struct {
			Kind   string `json:"kind"`
			Digest string `json:"digest"`
		} `json:"items"`
	}
	if err := json.Unmarshal(doc, &bundle); err != nil {
		fmt.Fprintln(os.Stderr, "invalid JSON:", err)
		os.Exit(1)
	}

	if bundle.APIVersion != "tekmovela.io/v1alpha1" || bundle.Kind != "EvidenceBundle" {
		fmt.Fprintf(os.Stderr, "not an EvidenceBundle: %s/%s\n", bundle.APIVersion, bundle.Kind)
		os.Exit(1)
	}

	failures := 0
	check := func(name, digest string) {
		if _, err := reference.ParseDigest(digest); err != nil {
			fmt.Printf("  FAIL %s: %v\n", name, err)
			failures++
		} else {
			fmt.Printf("  ok   %s: %s\n", name, digest[:16]+"…")
		}
	}

	fmt.Printf("EvidenceBundle %s\n", bundle.ID)
	check("bundle_digest", bundle.Bundle)
	for _, it := range bundle.Items {
		check("item["+it.Kind+"]", it.Digest)
	}

	if failures > 0 {
		fmt.Printf("FAILED: %d digest check(s)\n", failures)
		os.Exit(1)
	}
	fmt.Println("PASSED: all digests well-formed")
}
