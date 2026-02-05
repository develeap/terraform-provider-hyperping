// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// TestLiveContract_Maintenance_CRUD tests maintenance create, read, update, delete operations.
func TestLiveContract_Maintenance_CRUD(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "maintenance_crud",
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

	// Get a monitor to associate maintenance with
	monitors, err := client.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}

	var monitorUUIDs []string
	if len(monitors) > 0 {
		monitorUUIDs = []string{monitors[0].UUID}
	}

	// Test Create - schedule maintenance in the future
	startTime := time.Now().Add(24 * time.Hour).UTC()
	endTime := startTime.Add(2 * time.Hour)

	createReq := CreateMaintenanceRequest{
		Name:      "vcr-test-maintenance",
		Title:     LocalizedText{En: "VCR Test Maintenance Window"},
		Text:      LocalizedText{En: "This is a test maintenance window"},
		StartDate: startTime.Format(time.RFC3339),
		EndDate:   endTime.Format(time.RFC3339),
		Monitors:  monitorUUIDs,
	}

	maintenance, err := client.CreateMaintenance(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateMaintenance failed: %v", err)
	}

	t.Logf("Created maintenance: %s", maintenance.UUID)
	responseJSON, _ := json.MarshalIndent(maintenance, "", "  ")
	t.Logf("Maintenance response:\n%s", string(responseJSON))

	if maintenance.UUID == "" {
		t.Error("expected UUID to be set")
	}

	// Test Read
	readMaintenance, err := client.GetMaintenance(ctx, maintenance.UUID)
	if err != nil {
		t.Fatalf("GetMaintenance failed: %v", err)
	}
	if readMaintenance.UUID != maintenance.UUID {
		t.Errorf("expected UUID %q, got %q", maintenance.UUID, readMaintenance.UUID)
	}

	// Test Update
	newName := "vcr-test-maintenance-updated"
	updateReq := UpdateMaintenanceRequest{Name: &newName}
	updatedMaintenance, err := client.UpdateMaintenance(ctx, maintenance.UUID, updateReq)
	if err != nil {
		t.Fatalf("UpdateMaintenance failed: %v", err)
	}
	if updatedMaintenance.Name != newName {
		t.Errorf("expected updated name %q, got %q", newName, updatedMaintenance.Name)
	}

	// Test Delete
	if err = client.DeleteMaintenance(ctx, maintenance.UUID); err != nil {
		t.Fatalf("DeleteMaintenance failed: %v", err)
	}

	t.Log("Successfully completed maintenance CRUD cycle")
}

// TestLiveContract_Maintenance_List tests listing all maintenance windows.
func TestLiveContract_Maintenance_List(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "maintenance_list",
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

	maintenances, err := client.ListMaintenance(ctx)
	if err != nil {
		t.Fatalf("ListMaintenance failed: %v", err)
	}

	t.Logf("Found %d maintenance windows", len(maintenances))
	for i, m := range maintenances {
		data, _ := json.MarshalIndent(m, "", "  ")
		t.Logf("Maintenance %d:\n%s", i, string(data))
	}
}
