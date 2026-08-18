// Package harness provides HarnessVersion authoring and a deterministic local
// runner for TEKMOVELA SDK consumers. A harness is a first-class engineering
// asset: Runner/Adapter/Oracle/Reporter components frozen into an immutable
// HarnessVersion with a digest. The canonical form matches the Python and
// TypeScript SDKs (verification_contract_ref as a reference object,
// isolation_profile, determinism_profile).
package harness

import (
	"fmt"

	"github.com/axisrobo/tekmovela-open/sdk/go/contracts"
	"github.com/axisrobo/tekmovela-open/sdk/go/evidence"
)

// ComponentKind is a harness component role.
type ComponentKind string

const (
	ComponentRunner   ComponentKind = "runner"
	ComponentAdapter  ComponentKind = "adapter"
	ComponentOracle   ComponentKind = "oracle"
	ComponentReporter ComponentKind = "reporter"
)

var componentKinds = []ComponentKind{ComponentRunner, ComponentAdapter, ComponentOracle, ComponentReporter}

// ComponentRef references a frozen harness component.
type ComponentRef struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

// ComponentLock freezes the four component roles.
type ComponentLock map[ComponentKind]ComponentRef

// DeterminismMode is the determinism profile.
type DeterminismMode string

const (
	DeterminismDeterministic DeterminismMode = "deterministic"
	DeterminismStatistical   DeterminismMode = "statistical"
	DeterminismHybrid        DeterminismMode = "hybrid"
)

// HarnessVersion is an immutable, digest-bound harness.
type HarnessVersion struct {
	ID                      string
	Version                 string
	VerificationContractRef contracts.Reference
	ComponentLock           ComponentLock
	EvidenceSchema          string
	DeterminismMode         DeterminismMode
	Seed                    string
	Digest                  string
}

// NewHarnessVersion builds and validates a harness version.
func NewHarnessVersion(id, version string, vcRef contracts.Reference, lock ComponentLock, evidenceSchema string, determinismMode DeterminismMode, seed string) (*HarnessVersion, error) {
	if id == "" || version == "" {
		return nil, fmt.Errorf("id and version are required")
	}
	if err := vcRef.Validate(); err != nil {
		return nil, err
	}
	switch determinismMode {
	case DeterminismDeterministic, DeterminismStatistical, DeterminismHybrid:
	default:
		return nil, fmt.Errorf("unknown determinism mode %q", determinismMode)
	}
	lockCopy := make(ComponentLock, len(lock))
	for kind, ref := range lock {
		lockCopy[kind] = ref
	}
	for _, kind := range componentKinds {
		ref, ok := lockCopy[kind]
		if !ok {
			return nil, fmt.Errorf("component_lock.%s is required", kind)
		}
		if ref.Kind != "HarnessComponent" {
			return nil, fmt.Errorf("component_lock.%s.kind must be HarnessComponent", kind)
		}
		if _, err := contracts.ParseDigest(ref.Digest); err != nil {
			return nil, err
		}
	}
	return &HarnessVersion{
		ID:                      id,
		Version:                 version,
		VerificationContractRef: vcRef,
		ComponentLock:           lockCopy,
		EvidenceSchema:          evidenceSchema,
		DeterminismMode:         determinismMode,
		Seed:                    seed,
	}, nil
}

// Body returns the canonical form for digest computation.
func (hv *HarnessVersion) Body() map[string]any {
	body := map[string]any{
		"api_version": contracts.API_VERSION,
		"kind":        "HarnessVersion",
		"id":          hv.ID,
		"version":     hv.Version,
		"verification_contract_ref": map[string]any{
			"kind":    hv.VerificationContractRef.Kind,
			"id":      hv.VerificationContractRef.ID,
			"version": hv.VerificationContractRef.Version,
			"digest":  hv.VerificationContractRef.Digest,
		},
		"component_lock":      hv.lockJSON(),
		"evidence_schema":     hv.EvidenceSchema,
		"isolation_profile":   map[string]any{},
		"determinism_profile": map[string]any{"mode": string(hv.DeterminismMode)},
	}
	if hv.Seed != "" {
		body["environment_contract"] = map[string]any{"seed": hv.Seed}
	}
	return body
}

// lockJSON renders the component lock as plain maps so canonical JSON sorts
// nested keys like the Python (sort_keys=True) and TypeScript SDKs. Marshaling
// the ComponentRef structs directly would emit keys in declaration order and
// break cross-SDK digest parity.
func (hv *HarnessVersion) lockJSON() map[string]map[string]string {
	out := make(map[string]map[string]string, len(componentKinds))
	for _, kind := range componentKinds {
		ref := hv.ComponentLock[kind]
		out[string(kind)] = map[string]string{
			"kind":    ref.Kind,
			"id":      ref.ID,
			"version": ref.Version,
			"digest":  ref.Digest,
		}
	}
	return out
}

// CanonicalDigest computes the digest of the canonical form.
func (hv *HarnessVersion) CanonicalDigest() (string, error) {
	return contracts.ComputeDigest(hv.Body())
}

// Publish binds the canonical digest.
func (hv *HarnessVersion) Publish() error {
	d, err := hv.CanonicalDigest()
	if err != nil {
		return err
	}
	hv.Digest = d
	return nil
}

// RunStatus is the lifecycle status of a run.
type RunStatus string

const (
	RunStatusPending      RunStatus = "pending"
	RunStatusRunning      RunStatus = "running"
	RunStatusInterrupted  RunStatus = "interrupted"
	RunStatusPassed       RunStatus = "passed"
	RunStatusFailed       RunStatus = "failed"
	RunStatusInconclusive RunStatus = "inconclusive"
)

// RunRequest asks a runtime to execute a scenario.
type RunRequest struct {
	ScenarioRef       string `json:"scenario_ref"`
	HarnessVersionRef string `json:"harness_version_ref"`
	SUTRef            string `json:"sut_ref"`
	Seed              string `json:"seed"`
}

// RunHandle identifies an in-flight run.
type RunHandle struct {
	RunRef string `json:"run_ref"`
}

// Effect is a recorded external effect of a run.
type Effect struct {
	Action     string `json:"action"`
	ArgsDigest string `json:"args_digest"`
	Resource   string `json:"resource"`
	StateDelta string `json:"state_delta"`
}

// CheckpointResult is the outcome of a bounded checkpoint request.
type CheckpointResult struct {
	CheckpointRef string `json:"checkpoint_ref"`
	EvidenceRef   string `json:"evidence_ref"`
	Digest        string `json:"digest"`
}

// ProbeRequest is a bounded observation request.
type ProbeRequest struct {
	RunRef           string `json:"run_ref"`
	ControlLoopRef   string `json:"control_loop_ref"`
	Seed             string `json:"seed"`
	ObservationBound int    `json:"observation_bound"`
}

// ProbeResult returns effects observed within a bound plus evidence.
type ProbeResult struct {
	Effects     []Effect `json:"effects"`
	EvidenceRef string   `json:"evidence_ref"`
}

// RuntimeAdapter drives an external agent runtime. TEKMOVELA sends bounded
// test/checkpoint/status/probe requests and consumes effects; it never
// schedules execution or owns a run loop.
type RuntimeAdapter interface {
	StartRun(req RunRequest) (RunHandle, error)
	FetchEffects(handle RunHandle) ([]Effect, error)
	CancelRun(handle RunHandle) error
	CheckpointRun(handle RunHandle) (CheckpointResult, error)
	Status(handle RunHandle) (RunStatus, error)
	Probe(req ProbeRequest) (ProbeResult, error)
}

// Scenario is a test scenario to execute.
type Scenario struct {
	ID                      string
	Version                 string
	VerificationContractRef string
	Seed                    string
	HarnessVersionRef       string
}

// RunResult is the outcome of a local run.
type RunResult struct {
	RunID   string
	Status  string
	Bundle  *evidence.EvidenceBundle
	Effects []Effect
}

// LocalRunner executes scenarios against a RuntimeAdapter and emits sealed
// evidence.
type LocalRunner struct {
	runtime RuntimeAdapter
}

// NewLocalRunner builds a local runner.
func NewLocalRunner(rt RuntimeAdapter) *LocalRunner {
	return &LocalRunner{runtime: rt}
}

// Execute runs a scenario and seals the resulting evidence bundle. The bundle
// run ref binds the authoritative runtime handle ref.
func (lr *LocalRunner) Execute(scenario Scenario, sutRef, runID string) (*RunResult, error) {
	hvRef := scenario.HarnessVersionRef
	if hvRef == "" {
		hvRef = "local@v1"
	}
	handle, err := lr.runtime.StartRun(RunRequest{
		ScenarioRef:       scenario.ID,
		HarnessVersionRef: hvRef,
		SUTRef:            sutRef,
		Seed:              scenario.Seed,
	})
	if err != nil {
		return nil, err
	}
	effects, err := lr.runtime.FetchEffects(handle)
	if err != nil {
		return nil, err
	}
	status, err := lr.runtime.Status(handle)
	if err != nil {
		return nil, err
	}
	b := evidence.NewBundle(runID+".evidence", handle.RunRef, scenario.VerificationContractRef, hvRef)
	if err := b.SetScope(map[string]string{"intent_ref": scenario.ID, "capability_ref": sutRef}); err != nil {
		return nil, err
	}
	traceDigest, err := digestOf(handle.RunRef)
	if err != nil {
		return nil, err
	}
	if err := b.AddItem(evidence.Item{Kind: evidence.ItemKindTrace, URI: "local://trace/exec.jsonl", Digest: traceDigest}); err != nil {
		return nil, err
	}
	for _, e := range effects {
		if err := b.AddItem(evidence.Item{Kind: evidence.ItemKindEffect, URI: e.Action + "/" + e.Resource, Digest: e.ArgsDigest}); err != nil {
			return nil, err
		}
	}
	if err := b.Seal(); err != nil {
		return nil, err
	}
	return &RunResult{RunID: handle.RunRef, Status: string(status), Bundle: b, Effects: effects}, nil
}

func digestOf(v any) (string, error) {
	return contracts.ComputeDigest(v)
}
