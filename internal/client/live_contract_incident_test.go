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

// TestLiveContract_Incident_List tests listing all incidents.
func TestLiveContract_Incident_List(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "incident_list",
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

	incidents, err := client.ListIncidents(ctx)
	if err != nil {
		t.Fatalf("ListIncidents failed: %v", err)
	}

	t.Logf("Found %d incidents", len(incidents))
	for i, inc := range incidents {
		data, _ := json.MarshalIndent(inc, "", "  ")
		t.Logf("Incident %d:\n%s", i, string(data))
	}
}

// TestLiveContract_Incident_CRUD tests incident create, read, update, delete operations.
// NOTE: Requires status page access - skips if status pages are not available.
func TestLiveContract_Incident_CRUD(t *testing.T) {
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "incident_crud",
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

	// Get a status page to associate the incident with (required)
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: status pages not accessible (required for incident CRUD)")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test Create
	createReq := CreateIncidentRequest{
		Title:       LocalizedText{En: "VCR Test Incident"},
		Text:        LocalizedText{En: "This is a test incident"},
		Type:        "incident",
		StatusPages: []string{statusPageUUID},
	}

	incident, err := client.CreateIncident(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateIncident failed: %v", err)
	}

	t.Logf("Created incident: %s", incident.UUID)
	responseJSON, _ := json.MarshalIndent(incident, "", "  ")
	t.Logf("Incident response:\n%s", string(responseJSON))

	if incident.UUID == "" {
		t.Error("expected UUID to be set")
	}

	// Test Read
	readIncident, err := client.GetIncident(ctx, incident.UUID)
	if err != nil {
		t.Fatalf("GetIncident failed: %v", err)
	}
	if readIncident.UUID != incident.UUID {
		t.Errorf("expected UUID %q, got %q", incident.UUID, readIncident.UUID)
	}

	// Test Delete
	if err = client.DeleteIncident(ctx, incident.UUID); err != nil {
		t.Fatalf("DeleteIncident failed: %v", err)
	}

	t.Log("Successfully completed incident CRUD cycle")
}
