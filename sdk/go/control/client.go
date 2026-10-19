// Package control is a typed client for the TEKMOVELA control API.
package control

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrNotFound     = errors.New("tekmovela: not found")
	ErrUnauthorized = errors.New("tekmovela: unauthenticated")
	ErrConflict     = errors.New("tekmovela: conflict")
	ErrUnavailable  = errors.New("tekmovela: unavailable")
)

// GateDecision is a published G0..G5 decision.
type GateDecision struct {
	APIVersion          string   `json:"api_version"`
	Kind                string   `json:"kind"`
	ID                  string   `json:"id"`
	Gate                string   `json:"gate"`
	Decision            string   `json:"decision"`
	ReleaseCandidateRef string   `json:"release_candidate_ref"`
	PolicyVersion       string   `json:"policy_version"`
	BlockingReasons     []string `json:"blocking_reasons,omitempty"`
	EvidenceRefs        []string `json:"evidence_refs,omitempty"`
}

// ReleaseCandidate is the subject of assurance.
type ReleaseCandidate struct {
	APIVersion      string   `json:"api_version"`
	Kind            string   `json:"kind"`
	ID              string   `json:"id"`
	CapabilityRef   string   `json:"capability_ref"`
	Version         string   `json:"version"`
	ArtifactDigest  string   `json:"artifact_digest"`
	IntentScopeRefs []string `json:"intent_scope_refs,omitempty"`
	GateDecisions   []string `json:"gate_decisions,omitempty"`
}

// Client talks to the TEKMOVELA control API.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// New returns a control client. The base URL may include a path prefix.
func New(baseURL, token string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   strings.TrimSpace(token),
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// WithHTTPClient overrides the HTTP client.
func (c *Client) WithHTTPClient(hc *http.Client) *Client {
	if hc != nil {
		c.http = hc
	}
	return c
}

// GetGateDecision fetches a published gate decision by id.
func (c *Client) GetGateDecision(ctx context.Context, id string) (*GateDecision, error) {
	var out GateDecision
	if err := c.do(ctx, http.MethodGet, "/v1/gate-decisions/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// PublishReleaseCandidate publishes a release candidate.
func (c *Client) PublishReleaseCandidate(ctx context.Context, candidate *ReleaseCandidate) (*ReleaseCandidate, error) {
	if candidate == nil {
		return nil, errors.New("tekmovela: release candidate is required")
	}
	var out ReleaseCandidate
	if err := c.do(ctx, http.MethodPost, "/v1/release-candidates", candidate, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetReleaseCandidate fetches a published release candidate by id.
func (c *Client) GetReleaseCandidate(ctx context.Context, id string) (*ReleaseCandidate, error) {
	var out ReleaseCandidate
	if err := c.do(ctx, http.MethodGet, "/v1/release-candidates/"+url.PathEscape(id), nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out == nil {
			return nil
		}
		return json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(out)
	}
	payload, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	switch resp.StatusCode {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusUnauthorized, http.StatusForbidden:
		return ErrUnauthorized
	case http.StatusConflict:
		return ErrConflict
	case http.StatusServiceUnavailable:
		return ErrUnavailable
	default:
		return fmt.Errorf("%w: status %d: %s", ErrUnavailable, resp.StatusCode, strings.TrimSpace(string(payload)))
	}
}
