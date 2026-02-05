// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// These tests record real API interactions using go-vcr.
// Run with RECORD_MODE=true HYPERPING_API_KEY=xxx to record new cassettes.
// Run without RECORD_MODE to replay from existing cassettes.

func TestLiveContract_Healthcheck_CRUD(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "healthcheck_crud",
		Mode:         testutil.GetRecordMode(),
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	// Create client with VCR transport
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key" // Used during replay
	}

	client := NewClient(apiKey,
		WithHTTPClient(httpClient),
	)

	ctx := context.Background()

	// Test Create
	periodValue := 60
	periodType := "seconds"
	createReq := CreateHealthcheckRequest{
		Name:             "vcr-test-healthcheck",
		PeriodValue:      &periodValue,
		PeriodType:       &periodType,
		GracePeriodValue: 300,
		GracePeriodType:  "seconds",
	}

	healthcheck, err := client.CreateHealthcheck(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateHealthcheck failed: %v", err)
	}

	t.Logf("Created healthcheck: %s", healthcheck.UUID)

	// Verify response has expected fields
	if healthcheck.UUID == "" {
		t.Error("expected UUID to be set")
	}
	if healthcheck.Name != "vcr-test-healthcheck" {
		t.Errorf("expected name 'vcr-test-healthcheck', got %q", healthcheck.Name)
	}
	if healthcheck.PingURL == "" {
		t.Error("expected PingURL to be set")
	}

	// Test Read
	readHealthcheck, err := client.GetHealthcheck(ctx, healthcheck.UUID)
	if err != nil {
		t.Fatalf("GetHealthcheck failed: %v", err)
	}

	if readHealthcheck.UUID != healthcheck.UUID {
		t.Errorf("expected UUID %q, got %q", healthcheck.UUID, readHealthcheck.UUID)
	}

	// Test Update
	updateReq := UpdateHealthcheckRequest{
		Name:             strPtr("vcr-test-healthcheck-updated"),
		PeriodValue:      intPtr(120),
		PeriodType:       strPtr("seconds"),
		GracePeriodValue: intPtr(600),
		GracePeriodType:  strPtr("seconds"),
	}

	updatedHealthcheck, err := client.UpdateHealthcheck(ctx, healthcheck.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateHealthcheck failed: %v", err)
	}

	if updatedHealthcheck.Name != "vcr-test-healthcheck-updated" {
		t.Errorf("expected updated name, got %q", updatedHealthcheck.Name)
	}

	// Test Delete
	err = client.DeleteHealthcheck(ctx, healthcheck.UUID)
	if err != nil {
		t.Fatalf("DeleteHealthcheck failed: %v", err)
	}

	t.Log("Successfully completed healthcheck CRUD cycle")
}

func TestLiveContract_Monitor_CRUD(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "monitor_crud",
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

	client := NewClient(apiKey,
		WithHTTPClient(httpClient),
	)

	ctx := context.Background()

	// Test Create
	createReq := CreateMonitorRequest{
		Name:               "vcr-test-monitor",
		URL:                "https://httpstat.us/200",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     60,
		Regions:            []string{"london", "frankfurt"},
		FollowRedirects:    boolPtr(true),
		ExpectedStatusCode: "200",
		Paused:             true, // Create paused to avoid actual checks
	}

	monitor, err := client.CreateMonitor(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateMonitor failed: %v", err)
	}

	t.Logf("Created monitor: %s", monitor.UUID)

	// Log the full response for schema discovery
	responseJSON, _ := json.MarshalIndent(monitor, "", "  ")
	t.Logf("Monitor response:\n%s", string(responseJSON))

	// Verify core fields
	if monitor.UUID == "" {
		t.Error("expected UUID to be set")
	}
	if monitor.Name != "vcr-test-monitor" {
		t.Errorf("expected name 'vcr-test-monitor', got %q", monitor.Name)
	}

	// Test Read
	readMonitor, err := client.GetMonitor(ctx, monitor.UUID)
	if err != nil {
		t.Fatalf("GetMonitor failed: %v", err)
	}

	if readMonitor.UUID != monitor.UUID {
		t.Errorf("expected UUID %q, got %q", monitor.UUID, readMonitor.UUID)
	}

	// Test Delete
	err = client.DeleteMonitor(ctx, monitor.UUID)
	if err != nil {
		t.Fatalf("DeleteMonitor failed: %v", err)
	}

	t.Log("Successfully completed monitor CRUD cycle")
}

func TestLiveContract_Outage_CRUD(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "outage_crud",
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

	client := NewClient(apiKey,
		WithHTTPClient(httpClient),
	)

	ctx := context.Background()

	// First, we need a monitor to create an outage for
	// Skip if no monitors available (would need to create one first)
	monitors, err := client.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}

	if len(monitors) == 0 {
		t.Skip("No monitors available for outage test")
	}

	monitorUUID := monitors[0].UUID
	t.Logf("Using monitor %s for outage test", monitorUUID)

	// Test Create
	endDate := "2026-01-01T01:00:00Z"
	createReq := CreateOutageRequest{
		MonitorUUID: monitorUUID,
		StartDate:   "2026-01-01T00:00:00Z",
		EndDate:     &endDate,
		StatusCode:  500,
		Description: "Manual outage for contract testing",
		OutageType:  "manual",
	}

	outage, err := client.CreateOutage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateOutage failed: %v", err)
	}

	t.Logf("Created outage: %s", outage.UUID)

	// Log full response for schema discovery
	responseJSON, _ := json.MarshalIndent(outage, "", "  ")
	t.Logf("Outage response:\n%s", string(responseJSON))

	// Verify fields
	if outage.UUID == "" {
		t.Error("expected UUID to be set")
	}

	// Test Read
	readOutage, err := client.GetOutage(ctx, outage.UUID)
	if err != nil {
		t.Fatalf("GetOutage failed: %v", err)
	}

	if readOutage.UUID != outage.UUID {
		t.Errorf("expected UUID %q, got %q", outage.UUID, readOutage.UUID)
	}

	// Test Delete
	err = client.DeleteOutage(ctx, outage.UUID)
	if err != nil {
		t.Fatalf("DeleteOutage failed: %v", err)
	}

	t.Log("Successfully completed outage CRUD cycle")
}

// TestLiveContract_DiscoverFields runs through all resources and logs
// the complete API response to help discover undocumented fields.
func TestLiveContract_DiscoverFields(t *testing.T) {
	if os.Getenv("DISCOVER_FIELDS") != "true" {
		t.Skip("Set DISCOVER_FIELDS=true to run field discovery")
	}

	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "discover_fields",
		Mode:         testutil.ModeRecord, // Always record for discovery
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	apiKey := os.Getenv("HYPERPING_API_KEY")
	client := NewClient(apiKey,
		WithHTTPClient(httpClient),
	)

	ctx := context.Background()

	// List all resources and log their complete structure
	t.Run("Monitors", func(t *testing.T) {
		monitors, err := client.ListMonitors(ctx)
		if err != nil {
			t.Fatalf("ListMonitors failed: %v", err)
		}
		for i, m := range monitors {
			data, _ := json.MarshalIndent(m, "", "  ")
			t.Logf("Monitor %d:\n%s", i, string(data))
		}
	})

	t.Run("Healthchecks", func(t *testing.T) {
		healthchecks, err := client.ListHealthchecks(ctx)
		if err != nil {
			t.Fatalf("ListHealthchecks failed: %v", err)
		}
		for i, h := range healthchecks {
			data, _ := json.MarshalIndent(h, "", "  ")
			t.Logf("Healthcheck %d:\n%s", i, string(data))
		}
	})

	t.Run("StatusPages", func(t *testing.T) {
		pagesResp, err := client.ListStatusPages(ctx, nil, nil)
		if err != nil {
			t.Fatalf("ListStatusPages failed: %v", err)
		}
		for i, p := range pagesResp.StatusPages {
			data, _ := json.MarshalIndent(p, "", "  ")
			t.Logf("StatusPage %d:\n%s", i, string(data))
		}
	})

	t.Run("Incidents", func(t *testing.T) {
		incidents, err := client.ListIncidents(ctx)
		if err != nil {
			t.Fatalf("ListIncidents failed: %v", err)
		}
		for i, inc := range incidents {
			data, _ := json.MarshalIndent(inc, "", "  ")
			t.Logf("Incident %d:\n%s", i, string(data))
		}
	})

	t.Run("Maintenance", func(t *testing.T) {
		windows, err := client.ListMaintenance(ctx)
		if err != nil {
			t.Fatalf("ListMaintenance failed: %v", err)
		}
		for i, w := range windows {
			data, _ := json.MarshalIndent(w, "", "  ")
			t.Logf("Maintenance %d:\n%s", i, string(data))
		}
	})
}
