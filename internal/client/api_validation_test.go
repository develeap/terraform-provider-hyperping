// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"os"
	"testing"
)

// TestRealAPI_Cleanup removes any leftover test monitors.
// Run this manually to clean up after failed tests.
func TestRealAPI_Cleanup(t *testing.T) {
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		t.Skip("HYPERPING_API_KEY not set, skipping real API test")
	}

	c := NewClient(apiKey)
	ctx := context.Background()

	monitors, err := c.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}

	for _, m := range monitors {
		if m.Name == "TF Provider Test Monitor" || m.Name == "TF Provider Test Monitor (Updated)" {
			t.Logf("Deleting leftover test monitor: %s (%s)", m.Name, m.UUID)
			if err := c.DeleteMonitor(ctx, m.UUID); err != nil {
				t.Logf("Warning: failed to delete %s: %v", m.UUID, err)
			} else {
				t.Logf("Deleted %s successfully", m.UUID)
			}
		}
	}
}

// TestRealAPI_ListMonitors validates the client against the real Hyperping API.
// Requires HYPERPING_API_KEY environment variable to be set.
func TestRealAPI_ListMonitors(t *testing.T) {
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		t.Skip("HYPERPING_API_KEY not set, skipping real API test")
	}

	c := NewClient(apiKey)
	ctx := context.Background()

	monitors, err := c.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}

	t.Logf("Found %d monitors", len(monitors))
	for i, m := range monitors {
		if i < 3 {
			t.Logf("  - %s: %s (protocol=%s, method=%s, frequency=%d)",
				m.UUID, m.Name, m.Protocol, m.HTTPMethod, m.CheckFrequency)
		}
	}
}

// TestRealAPI_ListIncidents validates incident listing against the real API.
func TestRealAPI_ListIncidents(t *testing.T) {
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		t.Skip("HYPERPING_API_KEY not set, skipping real API test")
	}

	c := NewClient(apiKey)
	ctx := context.Background()

	incidents, err := c.ListIncidents(ctx)
	if err != nil {
		t.Fatalf("ListIncidents failed: %v", err)
	}

	t.Logf("Found %d incidents", len(incidents))
	for i, inc := range incidents {
		if i < 3 {
			t.Logf("  - %s: %s (type=%s)", inc.UUID, inc.Title.En, inc.Type)
		}
	}
}

// TestRealAPI_ListMaintenance validates maintenance listing against the real API.
func TestRealAPI_ListMaintenance(t *testing.T) {
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		t.Skip("HYPERPING_API_KEY not set, skipping real API test")
	}

	c := NewClient(apiKey)
	ctx := context.Background()

	maintenances, err := c.ListMaintenance(ctx)
	if err != nil {
		t.Fatalf("ListMaintenance failed: %v", err)
	}

	t.Logf("Found %d maintenance windows", len(maintenances))
	for i, m := range maintenances {
		if i < 3 {
			t.Logf("  - %s: %s", m.UUID, m.Name)
		}
	}
}

// TestRealAPI_MonitorCRUD validates the full create/read/update/delete cycle.
func TestRealAPI_MonitorCRUD(t *testing.T) {
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		t.Skip("HYPERPING_API_KEY not set, skipping real API test")
	}

	c := NewClient(apiKey)
	ctx := context.Background()

	// Create a test monitor
	// Reliable public test endpoints:
	//   - https://httpbin.org/status/200 - Popular API testing service
	//   - https://mock.httpstatus.io/200 - Simple status code mock
	//   - https://example.com - IANA maintained, always returns 200
	// Using httpbin.org as it's specifically designed for HTTP testing
	t.Log("Creating test monitor...")
	followRedirects := true
	createReq := CreateMonitorRequest{
		Name:               "TF Provider Test Monitor",
		URL:                "https://httpbin.org/status/200",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     300,
		ExpectedStatusCode: "200",
		FollowRedirects:    &followRedirects,
		Regions:            []string{"london", "virginia"},
	}

	monitor, err := c.CreateMonitor(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateMonitor failed: %v", err)
	}
	t.Logf("Created monitor: %s (ID: %s)", monitor.Name, monitor.UUID)

	// Clean up on exit
	defer func() {
		t.Log("Deleting test monitor...")
		if deleteErr := c.DeleteMonitor(ctx, monitor.UUID); deleteErr != nil {
			t.Logf("Warning: Failed to delete test monitor: %v", deleteErr)
		} else {
			t.Log("Test monitor deleted successfully")
		}
	}()

	// Verify creation
	if monitor.Name != createReq.Name {
		t.Errorf("Name mismatch: expected %s, got %s", createReq.Name, monitor.Name)
	}
	if monitor.Protocol != createReq.Protocol {
		t.Errorf("Protocol mismatch: expected %s, got %s", createReq.Protocol, monitor.Protocol)
	}

	// Read the monitor back
	t.Log("Reading monitor back...")
	readMonitor, err := c.GetMonitor(ctx, monitor.UUID)
	if err != nil {
		t.Fatalf("GetMonitor failed: %v", err)
	}
	if readMonitor.UUID != monitor.UUID {
		t.Errorf("UUID mismatch: expected %s, got %s", monitor.UUID, readMonitor.UUID)
	}
	t.Logf("Read monitor: %s", readMonitor.Name)

	// Update the monitor
	t.Log("Updating monitor...")
	newName := "TF Provider Test Monitor (Updated)"
	updateReq := UpdateMonitorRequest{
		Name: &newName,
	}
	updatedMonitor, err := c.UpdateMonitor(ctx, monitor.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateMonitor failed: %v", err)
	}
	if updatedMonitor.Name != newName {
		t.Errorf("Updated name mismatch: expected %s, got %s", newName, updatedMonitor.Name)
	}
	t.Logf("Updated monitor name to: %s", updatedMonitor.Name)

	t.Log("Monitor CRUD test completed successfully!")
}

// TestRealAPI_GetMonitor validates fetching a single monitor.
func TestRealAPI_GetMonitor(t *testing.T) {
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		t.Skip("HYPERPING_API_KEY not set, skipping real API test")
	}

	c := NewClient(apiKey)
	ctx := context.Background()

	// First list monitors to get a valid ID
	monitors, err := c.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}

	if len(monitors) == 0 {
		t.Skip("No monitors found, skipping GetMonitor test")
	}

	// Get the first monitor
	monitorID := monitors[0].UUID
	monitor, err := c.GetMonitor(ctx, monitorID)
	if err != nil {
		t.Fatalf("GetMonitor(%s) failed: %v", monitorID, err)
	}

	t.Logf("Got monitor: %s", monitor.Name)
	t.Logf("  UUID: %s", monitor.UUID)
	t.Logf("  URL: %s", monitor.URL)
	t.Logf("  Protocol: %s", monitor.Protocol)
	t.Logf("  HTTPMethod: %s", monitor.HTTPMethod)
	t.Logf("  CheckFrequency: %d", monitor.CheckFrequency)
	t.Logf("  ExpectedStatusCode: %s", monitor.ExpectedStatusCode)
	t.Logf("  FollowRedirects: %v", monitor.FollowRedirects)
	t.Logf("  Paused: %v", monitor.Paused)
	t.Logf("  Regions: %v", monitor.Regions)
	if len(monitor.RequestHeaders) > 0 {
		t.Logf("  RequestHeaders: %d headers", len(monitor.RequestHeaders))
	}
	if monitor.RequestBody != "" {
		t.Logf("  RequestBody: %s", monitor.RequestBody)
	}
}
