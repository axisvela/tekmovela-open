package harness

import (
	"strings"
	"testing"

	"github.com/axisrobo/tekmovela-open/sdk/go/contracts"
)

func TestGoldenHarnessVersionDigest(t *testing.T) {
	hv, err := NewHarnessVersion(
		"checkout.double_spend",
		"v1",
		contracts.Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:" + repeat('a', 64)},
		ComponentLock{
			ComponentRunner:   {Kind: "HarnessComponent", ID: "runner.deterministic", Version: "v1", Digest: "sha256:" + repeat('c', 64)},
			ComponentAdapter:  {Kind: "HarnessComponent", ID: "adapter.praxovela", Version: "v1", Digest: "sha256:" + repeat('d', 64)},
			ComponentOracle:   {Kind: "HarnessComponent", ID: "oracle.effect_ledger", Version: "v1", Digest: "sha256:" + repeat('e', 64)},
			ComponentReporter: {Kind: "HarnessComponent", ID: "reporter.evidence", Version: "v1", Digest: "sha256:" + repeat('f', 64)},
		},
		"tekmovela.evidence-bundle.v1alpha1.schema.json",
	)
	if err != nil {
		t.Fatal(err)
	}
	d, err := hv.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	if d != "sha256:c20430a95d5d42570cdd2434bf25ac9bbc3a846d9e6ffa1307a8ade5164b0557" {
		t.Fatalf("harness version digest %q != golden", d)
	}
}

func TestHarnessVersionRejectsMissingComponent(t *testing.T) {
	_, err := NewHarnessVersion(
		"checkout.double_spend",
		"v1",
		contracts.Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:" + repeat('c', 64)},
		ComponentLock{
			ComponentRunner: {Kind: "HarnessComponent", ID: "r.1", Version: "v1", Digest: "sha256:" + repeat('c', 64)},
		},
		"evidence.bundle@v1",
	)
	if err == nil {
		t.Fatal("missing component lock entries must fail")
	}
}

func TestLocalRunnerSealsBundle(t *testing.T) {
	runner := NewLocalRunner(&stubRuntime{})
	result, err := runner.Execute(
		Scenario{ID: "scenario.double_spend", Version: "v1", VerificationContractRef: "checkout.atomicity", Seed: "s1"},
		"cap/v2",
		"run.e0001",
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Bundle.BundleDigest == "" {
		t.Fatal("run bundle must be sealed")
	}
	if err := result.Bundle.Verify(); err != nil {
		t.Fatal(err)
	}
	if result.RunID != result.Bundle.RunRef {
		t.Fatalf("run id %q != bundle run ref %q", result.RunID, result.Bundle.RunRef)
	}
}

type stubRuntime struct{}

func (stubRuntime) StartRun(req RunRequest) (RunHandle, error) {
	return RunHandle{RunRef: req.ScenarioRef + "/run-" + req.Seed}, nil
}
func (stubRuntime) FetchEffects(handle RunHandle) ([]Effect, error) {
	return []Effect{{Action: "debit", ArgsDigest: "sha256:" + repeat('a', 64), Resource: "ledger", StateDelta: "-1"}}, nil
}
func (stubRuntime) CancelRun(RunHandle) error { return nil }
func (stubRuntime) CheckpointRun(RunHandle) (CheckpointResult, error) {
	return CheckpointResult{CheckpointRef: "ckp", EvidenceRef: "ev", Digest: "sha256:" + repeat('b', 64)}, nil
}
func (stubRuntime) Status(RunHandle) (RunStatus, error) { return RunStatusPassed, nil }
func (stubRuntime) Probe(ProbeRequest) (ProbeResult, error) {
	return ProbeResult{EvidenceRef: "ev/probe"}, nil
}

func repeat(c byte, n int) string {
	return strings.Repeat(string(c), n)
}
