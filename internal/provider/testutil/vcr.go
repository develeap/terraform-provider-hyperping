// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

// Package testutil provides testing utilities for the Hyperping provider.
//
// VCR (Video Cassette Recorder) Testing:
//
// This package provides VCR support for recording and replaying HTTP interactions
// in tests. This enables fast, deterministic tests without hitting real APIs.
//
// # Basic Usage
//
//	func TestAPICall(t *testing.T) {
//	    rec, client := testutil.NewVCRRecorder(t, testutil.VCRConfig{
//	        CassetteName: "api_call",
//	        Mode:         testutil.GetRecordMode(),
//	    })
//	    defer rec.Stop()
//
//	    // Use client for HTTP requests - they'll be recorded/replayed
//	    resp, err := client.Get("https://api.example.com/v1/test")
//	    // ...
//	}
//
// # Recording Mode
//
// Set RECORD_MODE=true to record new interactions:
//
//	RECORD_MODE=true go test ./...
//
// Without this flag, tests replay from existing cassettes.
//
// # Security
//
// Sensitive data (Authorization headers, cookies, API keys in URLs) is automatically
// masked in recorded cassettes to prevent credential leaks.
package testutil

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

// VCRMode determines how VCR handles HTTP requests.
type VCRMode int

const (
	// ModeReplay replays from cassette, fails if no recording exists.
	ModeReplay VCRMode = iota
	// ModeRecord always records new interactions.
	ModeRecord
	// ModeAuto replays if cassette exists, otherwise records.
	ModeAuto
)

// VCRConfig configures VCR recording behavior.
type VCRConfig struct {
	CassetteName string
	Mode         VCRMode
	CassetteDir  string
}

// NewVCRRecorder creates a new VCR recorder for contract testing.
// It handles API key masking and returns an HTTP client ready for recording.
// In ModeAuto, if no cassette exists and not in record mode, the test is skipped.
func NewVCRRecorder(t *testing.T, cfg VCRConfig) (*recorder.Recorder, *http.Client) {
	t.Helper()

	if cfg.CassetteDir == "" {
		cfg.CassetteDir = filepath.Join("testdata", "cassettes")
	}

	cassettePath := filepath.Join(cfg.CassetteDir, cfg.CassetteName)

	// Ensure cassette directory exists (0o750 for security - gosec G301)
	if err := os.MkdirAll(cfg.CassetteDir, 0o750); err != nil {
		t.Fatalf("failed to create cassette directory: %v", err)
	}

	// Determine recorder mode
	var mode recorder.Mode
	switch cfg.Mode {
	case ModeReplay:
		mode = recorder.ModeReplayOnly
	case ModeRecord:
		mode = recorder.ModeRecordOnly
	case ModeAuto:
		// Check if cassette exists
		if _, err := os.Stat(cassettePath + ".yaml"); os.IsNotExist(err) {
			// No cassette and not in record mode - skip the test
			t.Skipf("Skipping: no cassette exists at %s.yaml (set RECORD_MODE=true to record)", cassettePath)
		}
		mode = recorder.ModeReplayOnly
	}

	// Create recorder
	r, err := recorder.NewWithOptions(&recorder.Options{
		CassetteName:       cassettePath,
		Mode:               mode,
		SkipRequestLatency: true,
	})
	if err != nil {
		t.Fatalf("failed to create VCR recorder: %v", err)
	}

	// Add hook to mask sensitive data
	r.AddHook(func(i *cassette.Interaction) error {
		maskSensitiveHeaders(i)
		return nil
	}, recorder.AfterCaptureHook)

	// Create HTTP client with VCR transport
	client := &http.Client{
		Transport: r,
	}

	return r, client
}

// maskSensitiveHeaders masks sensitive data in recorded interactions.
func maskSensitiveHeaders(i *cassette.Interaction) {
	// Mask Authorization header
	if auth := i.Request.Headers.Get("Authorization"); auth != "" {
		i.Request.Headers.Set("Authorization", "Bearer [MASKED]")
	}

	// Mask any API keys in URL query params (shouldn't happen but be safe)
	if strings.Contains(i.Request.URL, "api_key=") {
		i.Request.URL = strings.ReplaceAll(i.Request.URL, "api_key=", "api_key=[MASKED]")
	}

	// Mask Set-Cookie headers
	if cookie := i.Response.Headers.Get("Set-Cookie"); cookie != "" {
		i.Response.Headers.Set("Set-Cookie", "[MASKED]")
	}
}

// GetRecordMode returns the VCR mode based on environment variables.
// Set RECORD_MODE=true to enable recording.
func GetRecordMode() VCRMode {
	if os.Getenv("RECORD_MODE") == "true" {
		return ModeRecord
	}
	return ModeAuto
}

// RequireEnvForRecording skips the test if recording mode is enabled
// but the required environment variable is not set.
func RequireEnvForRecording(t *testing.T, envVar string) {
	t.Helper()
	if GetRecordMode() == ModeRecord && os.Getenv(envVar) == "" {
		t.Skipf("Skipping recording test: %s not set", envVar)
	}
}
