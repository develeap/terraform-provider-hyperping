// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// TestLiveContract_Monitor_Update tests updating a monitor.
func TestLiveContract_Monitor_Update(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "monitor_update",
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

	// Create monitor first
	createReq := CreateMonitorRequest{
		Name:               "vcr-test-monitor-update",
		URL:                "https://httpstat.us/200",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     60,
		Regions:            []string{"london"},
		FollowRedirects:    boolPtr(true),
		ExpectedStatusCode: "200",
		Paused:             true,
	}

	monitor, err := client.CreateMonitor(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateMonitor failed: %v", err)
	}
	defer client.DeleteMonitor(ctx, monitor.UUID)

	// Update monitor
	newName := "vcr-test-monitor-updated"
	newFreq := 120
	newRegions := []string{"london", "frankfurt"}
	updateReq := UpdateMonitorRequest{
		Name:           &newName,
		CheckFrequency: &newFreq,
		Regions:        &newRegions,
	}

	updated, err := client.UpdateMonitor(ctx, monitor.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateMonitor failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(updated, "", "  ")
	t.Logf("Updated monitor:\n%s", string(responseJSON))

	if updated.Name != newName {
		t.Errorf("expected name %q, got %q", newName, updated.Name)
	}
	if updated.CheckFrequency != newFreq {
		t.Errorf("expected frequency %d, got %d", newFreq, updated.CheckFrequency)
	}
}

// TestLiveContract_Monitor_PauseResume tests pausing and resuming a monitor.
func TestLiveContract_Monitor_PauseResume(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "monitor_pause_resume",
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

	// Create monitor (not paused)
	createReq := CreateMonitorRequest{
		Name:               "vcr-test-monitor-pause",
		URL:                "https://httpstat.us/200",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     60,
		Regions:            []string{"london"},
		FollowRedirects:    boolPtr(true),
		ExpectedStatusCode: "200",
		Paused:             false,
	}

	monitor, err := client.CreateMonitor(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateMonitor failed: %v", err)
	}
	defer client.DeleteMonitor(ctx, monitor.UUID)

	// Pause monitor
	paused, err := client.PauseMonitor(ctx, monitor.UUID)
	if err != nil {
		t.Fatalf("PauseMonitor failed: %v", err)
	}
	if !paused.Paused {
		t.Error("expected monitor to be paused")
	}

	// Resume monitor
	resumed, err := client.ResumeMonitor(ctx, monitor.UUID)
	if err != nil {
		t.Fatalf("ResumeMonitor failed: %v", err)
	}
	if resumed.Paused {
		t.Error("expected monitor to be resumed")
	}

	t.Log("Successfully paused and resumed monitor")
}

// TestLiveContract_Monitor_NotFound tests getting a non-existent monitor.
func TestLiveContract_Monitor_NotFound(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "monitor_not_found",
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

	// Try to get non-existent monitor
	_, err := client.GetMonitor(ctx, "mon_nonexistent123456")
	if err == nil {
		t.Fatal("expected error for non-existent monitor")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != 404 {
		t.Errorf("expected status code 404, got %d", apiErr.StatusCode)
	}

	t.Logf("Got expected error: %v", apiErr)
}

// TestLiveContract_Monitor_ValidationError tests creating monitor with invalid data.
func TestLiveContract_Monitor_ValidationError(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "monitor_validation_error",
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

	// Create monitor with invalid URL
	createReq := CreateMonitorRequest{
		Name:               "vcr-test-invalid",
		URL:                "not-a-valid-url",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     60,
		Regions:            []string{"london"},
		ExpectedStatusCode: "200",
	}

	_, err := client.CreateMonitor(ctx, createReq)
	if err == nil {
		t.Fatal("expected validation error for invalid URL")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}

	if apiErr.StatusCode != 400 && apiErr.StatusCode != 422 {
		t.Errorf("expected status code 400 or 422, got %d", apiErr.StatusCode)
	}

	t.Logf("Got expected validation error: %v", apiErr)
}
