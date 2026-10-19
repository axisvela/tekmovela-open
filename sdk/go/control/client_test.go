package control

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetGateDecision(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/gate-decisions/g5.pass" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Errorf("authorization = %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"api_version":"tekmovela.io/v1","kind":"GateDecision","id":"g5.pass","gate":"G5_evidence","decision":"pass","release_candidate_ref":"rc.1","policy_version":"v1"}`))
	}))
	defer server.Close()
	client := New(server.URL, "token")
	decision, err := client.GetGateDecision(context.Background(), "g5.pass")
	if err != nil {
		t.Fatal(err)
	}
	if decision.Decision != "pass" || decision.ID != "g5.pass" {
		t.Fatalf("decision = %+v", decision)
	}
}

func TestPublishReleaseCandidate(t *testing.T) {
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/release-candidates" || r.Method != http.MethodPost {
			t.Errorf("request = %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"api_version":"tekmovela.io/v1","kind":"ReleaseCandidate","id":"coordivela.work.item_1","capability_ref":"work-item:1","version":"v2","artifact_digest":"sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}`))
	}))
	defer server.Close()
	client := New(server.URL, "")
	candidate := &ReleaseCandidate{
		APIVersion: "tekmovela.io/v1", Kind: "ReleaseCandidate",
		ID: "coordivela.work.item_1", CapabilityRef: "work-item:1", Version: "v2",
		ArtifactDigest: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	}
	published, err := client.PublishReleaseCandidate(context.Background(), candidate)
	if err != nil {
		t.Fatal(err)
	}
	if published.ID != candidate.ID || body["capability_ref"] != "work-item:1" {
		t.Fatalf("published = %+v, body = %+v", published, body)
	}
}

func TestErrorMapping(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	client := New(server.URL, "")
	if _, err := client.GetGateDecision(context.Background(), "missing"); err != ErrNotFound {
		t.Fatalf("err = %v", err)
	}
}
