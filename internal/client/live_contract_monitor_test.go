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

// TestLiveContract_Monitor_CRUD tests monitor create, read, delete operations.
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

	client := NewClient(apiKey, WithHTTPClient(httpClient))
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
		Paused:             true,
	}

	monitor, err := client.CreateMonitor(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateMonitor failed: %v", err)
	}

	t.Logf("Created monitor: %s", monitor.UUID)
	responseJSON, _ := json.MarshalIndent(monitor, "", "  ")
	t.Logf("Monitor response:\n%s", string(responseJSON))

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
	if err = client.DeleteMonitor(ctx, monitor.UUID); err != nil {
		t.Fatalf("DeleteMonitor failed: %v", err)
	}

	t.Log("Successfully completed monitor CRUD cycle")
}

// TestLiveContract_Monitor_List tests listing all monitors.
func TestLiveContract_Monitor_List(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "monitor_list",
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

	monitors, err := client.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}

	t.Logf("Found %d monitors", len(monitors))
	for i, m := range monitors {
		data, _ := json.MarshalIndent(m, "", "  ")
		t.Logf("Monitor %d:\n%s", i, string(data))
	}
}
