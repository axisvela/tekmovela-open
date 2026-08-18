// Command parity prints digests for cross-SDK parity checks.
// Usage: go run ./internal/parity <contracts|reference>
package main

import (
	"fmt"
	"os"

	"github.com/axisrobo/tekmovela-open/sdk/go/contracts"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: parity <contracts|reference>")
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
