// Command parity prints digests for cross-SDK parity checks.
// Usage: go run ./internal/parity <contracts|reference|evidence|harness>
package main

import (
	"fmt"
	"os"

	"github.com/axisrobo/tekmovela-open/sdk/go/contracts"
	"github.com/axisrobo/tekmovela-open/sdk/go/evidence"
	"github.com/axisrobo/tekmovela-open/sdk/go/harness"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: parity <contracts|reference|evidence|harness>")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "contracts":
		d, err := contracts.ComputeDigest(map[string]any{"a": 1, "b": 2})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(d)
	case "reference":
		ref, err := contracts.NewReference("VerificationContract", "checkout.atomicity", "v1", "sha256:"+repeat('a', 64))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		d, err := contracts.ComputeDigest(ref)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(d)
	case "evidence":
		b := evidence.NewBundle("bundle.e0001", "run.1", "checkout.atomicity@v1", "checkout.double_spend@v1")
		if err := b.AddItem(evidence.Item{Kind: evidence.ItemKindTrace, URI: "trace.1", Digest: "sha256:" + repeat('a', 64)}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		d, err := b.Digest()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(d)
		b.EnvironmentDigest = "sha256:" + repeat('d', 64)
		d, err = b.Digest()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(d)
	case "harness":
		ref, err := contracts.NewReference("VerificationContract", "checkout.atomicity", "v1", "sha256:"+repeat('a', 64))
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		hv, err := harness.NewHarnessVersion(
			"checkout.double_spend", "v1", ref,
			harness.ComponentLock{
				harness.ComponentRunner:   {Kind: "HarnessComponent", ID: "runner.deterministic", Version: "v1", Digest: "sha256:" + repeat('c', 64)},
				harness.ComponentAdapter:  {Kind: "HarnessComponent", ID: "adapter.praxovela", Version: "v1", Digest: "sha256:" + repeat('d', 64)},
				harness.ComponentOracle:   {Kind: "HarnessComponent", ID: "oracle.effect_ledger", Version: "v1", Digest: "sha256:" + repeat('e', 64)},
				harness.ComponentReporter: {Kind: "HarnessComponent", ID: "reporter.evidence", Version: "v1", Digest: "sha256:" + repeat('f', 64)},
			},
			"tekmovela.evidence-bundle.v1alpha1.schema.json",
			harness.DeterminismDeterministic, "",
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		d, err := hv.CanonicalDigest()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(d)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", os.Args[1])
		os.Exit(2)
	}
}

func repeat(c byte, n int) string {
	out := make([]byte, n)
	for i := range out {
		out[i] = c
	}
	return string(out)
}
