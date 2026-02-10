// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// TestLiveContract_Auth_Unauthorized tests access with invalid API key.
func TestLiveContract_Auth_Unauthorized(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "auth_unauthorized",
		Mode:         testutil.GetRecordMode(),
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	// Use invalid API key
	client := NewClient("invalid_api_key_123", WithHTTPClient(httpClient))
	ctx := context.Background()

	// Try to list monitors with invalid key
	_, err := client.ListMonitors(ctx)
	if err == nil {
		t.Fatal("expected error for invalid API key")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != 401 && apiErr.StatusCode != 403 {
		t.Errorf("expected status code 401 or 403, got %d", apiErr.StatusCode)
	}

	t.Logf("Got expected authentication error: %v", apiErr)
}

// TestLiveContract_Auth_MissingKey tests access without API key.
func TestLiveContract_Auth_MissingKey(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "auth_missing_key",
		Mode:         testutil.GetRecordMode(),
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	// Use empty API key
	client := NewClient("", WithHTTPClient(httpClient))
	ctx := context.Background()

	// Try to list monitors without key
	_, err := client.ListMonitors(ctx)
	if err == nil {
		t.Fatal("expected error for missing API key")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != 401 && apiErr.StatusCode != 403 {
		t.Errorf("expected status code 401 or 403, got %d", apiErr.StatusCode)
	}

	t.Logf("Got expected authentication error: %v", apiErr)
}

// TestLiveContract_RateLimit tests rate limiting behavior.
// NOTE: This test is commented out by default as it may trigger actual rate limits.
// Uncomment to test rate limit handling if you have appropriate permissions.
/*
func TestLiveContract_RateLimit(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "rate_limit",
		Mode:         testutil.GetRecordMode(),
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key"
	}

	client := NewClient(apiKey, WithHTTPClient(httpClient))
	ctx := context.Background()

	// Make many rapid requests to potentially trigger rate limit
	for i := 0; i < 100; i++ {
		_, err := client.ListMonitors(ctx)
		if err != nil {
			apiErr, ok := err.(*APIError)
			if ok && apiErr.StatusCode == 429 {
				t.Logf("Hit rate limit after %d requests", i+1)
				t.Logf("Rate limit error: %v", apiErr)
				t.Logf("Retry-After: %d seconds", apiErr.RetryAfter)
				return
			}
		}
	}

	t.Log("Did not hit rate limit (or rate limit is very high)")
}
*/

// TestLiveContract_Auth_ValidAPIKey tests that valid API key works across resources.
func TestLiveContract_Auth_ValidAPIKey(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "auth_valid_key",
		Mode:         testutil.GetRecordMode(),
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key"
	}

	client := NewClient(apiKey, WithHTTPClient(httpClient))
	ctx := context.Background()

	// Test valid key works for monitors
	_, err := client.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed with valid key: %v", err)
	}

	// Test valid key works for incidents
	_, err = client.ListIncidents(ctx)
	if err != nil {
		t.Fatalf("ListIncidents failed with valid key: %v", err)
	}

	// Test valid key works for maintenance
	_, err = client.ListMaintenance(ctx)
	if err != nil {
		t.Fatalf("ListMaintenance failed with valid key: %v", err)
	}

	t.Log("Valid API key works across all resources")
}
