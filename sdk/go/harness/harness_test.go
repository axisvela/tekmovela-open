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
		sampleLock(),
		"tekmovela.evidence-bundle.v1.schema.json",
		DeterminismDeterministic,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	d, err := hv.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	if d != "sha256:bc8520405b1d90daea66d17a62c415f0edee8ecb4a52b9ed1460bc6a5a9b0499" {
		t.Fatalf("harness version digest %q != golden", d)
	}
}

func TestGoldenHarnessVersionDigestWithSeed(t *testing.T) {
	hv, err := NewHarnessVersion(
		"checkout.double_spend",
		"v1",
		contracts.Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:" + repeat('a', 64)},
		sampleLock(),
		"tekmovela.evidence-bundle.v1.schema.json",
		DeterminismDeterministic,
		"s1",
	)
	if err != nil {
		t.Fatal(err)
	}
	ec, ok := hv.Body()["environment_contract"]
	if !ok {
		t.Fatal("seed must emit environment_contract")
	}
	if got := ec.(map[string]any)["seed"]; got != "s1" {
		t.Fatalf("environment_contract.seed = %v, want s1", got)
	}
	d, err := hv.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	if d != "sha256:ce4ba19c8f426de72786ca7b4206af833f487ee557e3444a056649023f110cd1" {
		t.Fatalf("seed harness version digest %q != golden", d)
	}
}

func TestGoldenHarnessVersionDigestStatistical(t *testing.T) {
	hv, err := NewHarnessVersion(
		"checkout.double_spend",
		"v1",
		contracts.Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:" + repeat('a', 64)},
		sampleLock(),
		"tekmovela.evidence-bundle.v1.schema.json",
		DeterminismStatistical,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	d, err := hv.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	if d != "sha256:ffc4b2e076bbb7588ac2923fcb6dc3f623016f022ab78daf0cc561e87046a519" {
		t.Fatalf("statistical harness version digest %q != golden", d)
	}
}

func TestPublishBindsDigest(t *testing.T) {
	hv, err := NewHarnessVersion(
		"checkout.double_spend",
		"v1",
		contracts.Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:" + repeat('a', 64)},
		sampleLock(),
		"tekmovela.evidence-bundle.v1.schema.json",
		DeterminismDeterministic,
		"",
	)
	if err != nil {
		t.Fatal(err)
	}
	want, err := hv.CanonicalDigest()
	if err != nil {
		t.Fatal(err)
	}
	if err := hv.Publish(); err != nil {
		t.Fatal(err)
	}
	if hv.Digest != want {
		t.Fatalf("Publish digest %q != canonical %q", hv.Digest, want)
	}
}

func TestNewHarnessVersionRejectsInvalidDeterminismMode(t *testing.T) {
	_, err := NewHarnessVersion(
		"checkout.double_spend",
		"v1",
		contracts.Reference{Kind: "VerificationContract", ID: "checkout.atomicity", Version: "v1", Digest: "sha256:" + repeat('a', 64)},
		sampleLock(),
		"tekmovela.evidence-bundle.v1.schema.json",
		DeterminismMode("random"),
		"",
	)
	if err == nil {
		t.Fatal("invalid determinism mode must fail")
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
		DeterminismDeterministic,
		"",
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
	if result.Status != string(RunStatusPassed) {
		t.Fatalf("run status %q != passed", result.Status)
	}
}

func TestExecuteDefaultsHarnessVersionRef(t *testing.T) {
	runner := NewLocalRunner(&stubRuntime{})
	result, err := runner.Execute(
		Scenario{ID: "scenario.double_spend", Version: "v1", VerificationContractRef: "checkout.atomicity", Seed: "s1"},
		"cap/v2",
		"run.e0002",
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Bundle.HarnessVersionRef != "local@v1" {
		t.Fatalf("harness version ref %q != local@v1", result.Bundle.HarnessVersionRef)
	}
}

func TestExecuteRecordsFailureStatus(t *testing.T) {
	runner := NewLocalRunner(&stubRuntime{status: RunStatusFailed})
	result, err := runner.Execute(
		Scenario{ID: "scenario.double_spend", Version: "v1", VerificationContractRef: "checkout.atomicity", Seed: "s1"},
		"cap/v2",
		"run.e0003",
	)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != string(RunStatusFailed) {
		t.Fatalf("run status %q != failed", result.Status)
	}
	if result.Bundle.BundleDigest == "" {
		t.Fatal("bundle must still be sealed on a failed run")
	}
	if err := result.Bundle.Verify(); err != nil {
		t.Fatal(err)
	}
}

type stubRuntime struct {
	status RunStatus
}

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
func (r stubRuntime) Status(RunHandle) (RunStatus, error) {
	if r.status == "" {
		return RunStatusPassed, nil
	}
	return r.status, nil
}
func (stubRuntime) Probe(ProbeRequest) (ProbeResult, error) {
	return ProbeResult{EvidenceRef: "ev/probe"}, nil
}

func sampleLock() ComponentLock {
	return ComponentLock{
		ComponentRunner:   {Kind: "HarnessComponent", ID: "runner.deterministic", Version: "v1", Digest: "sha256:" + repeat('c', 64)},
		ComponentAdapter:  {Kind: "HarnessComponent", ID: "adapter.praxovela", Version: "v1", Digest: "sha256:" + repeat('d', 64)},
		ComponentOracle:   {Kind: "HarnessComponent", ID: "oracle.effect_ledger", Version: "v1", Digest: "sha256:" + repeat('e', 64)},
		ComponentReporter: {Kind: "HarnessComponent", ID: "reporter.evidence", Version: "v1", Digest: "sha256:" + repeat('f', 64)},
	}
}

func repeat(c byte, n int) string {
	return strings.Repeat(string(c), n)
}
