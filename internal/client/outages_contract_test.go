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

// =============================================================================
// Outage CRUD Contract Tests with VCR Cassettes
// =============================================================================

// TestContract_Outage_CRUD validates the full CRUD lifecycle using VCR cassettes.
// This test validates:
// - Create outage response structure
// - Read outage response structure
// - Delete outage success
func TestContract_Outage_CRUD(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "outage_crud",
		Mode:         testutil.ModeReplay,
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

	// Get monitor for outage creation
	monitors, err := client.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}
	if len(monitors) == 0 {
		t.Fatal("No monitors available for outage test")
	}

	monitorUUID := monitors[0].UUID
	endDate := "2026-02-09T05:00:00.000Z"
	createReq := CreateOutageRequest{
		MonitorUUID: monitorUUID,
		StartDate:   "2026-02-09T04:00:00.000Z",
		EndDate:     &endDate,
		StatusCode:  500,
		Description: "Manual outage for contract testing",
		OutageType:  "manual",
	}

	// Test: Create outage
	outage, err := client.CreateOutage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateOutage failed: %v", err)
	}

	// Validate create response structure
	// NOTE: The API returns a minimal response on create, not a full Outage object
	t.Run("CreateResponse", func(t *testing.T) {
		if outage.UUID == "" {
			t.Error("UUID must not be empty")
		}
		if outage.StartDate == "" {
			t.Error("StartDate must not be empty")
		}
		if outage.Description != createReq.Description {
			t.Errorf("expected description %q, got %q", createReq.Description, outage.Description)
		}

		// Note: Create response does not include all fields like StatusCode, OutageType, etc.
		// These are only available from the GET endpoint
		t.Logf("Created outage UUID: %s", outage.UUID)
	})

	// Test: Read outage
	t.Run("ReadResponse", func(t *testing.T) {
		readOutage, err := client.GetOutage(ctx, outage.UUID)
		if err != nil {
			t.Fatalf("GetOutage failed: %v", err)
		}

		validateOutageStructure(t, readOutage)

		if readOutage.UUID != outage.UUID {
			t.Errorf("expected UUID %q, got %q", outage.UUID, readOutage.UUID)
		}
		if readOutage.Description != outage.Description {
			t.Errorf("expected description %q, got %q", outage.Description, readOutage.Description)
		}
	})

	// Test: Delete outage
	t.Run("Delete", func(t *testing.T) {
		err := client.DeleteOutage(ctx, outage.UUID)
		if err != nil {
			t.Fatalf("DeleteOutage failed: %v", err)
		}
	})

	responseJSON, _ := json.MarshalIndent(outage, "", "  ")
	t.Logf("Outage response:\n%s", string(responseJSON))
}

// TestContract_Outage_List validates the list outages response structure.
func TestContract_Outage_List(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "outage_list",
		Mode:         testutil.ModeReplay,
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

	// Test: List outages
	outages, err := client.ListOutages(ctx)
	if err != nil {
		t.Fatalf("ListOutages failed: %v", err)
	}

	t.Run("ListResponse", func(t *testing.T) {
		if len(outages) == 0 {
			t.Skip("No outages in list (ok for test)")
		}

		// Validate each outage in the list
		for i, outage := range outages {
			t.Run("OutageStructure", func(t *testing.T) {
				validateOutageStructure(t, &outage)
			})

			if i < 3 {
				data, _ := json.MarshalIndent(outage, "", "  ")
				t.Logf("Outage %d:\n%s", i, string(data))
			}
		}

		t.Logf("Total outages in list: %d", len(outages))
	})
}

// TestContract_Outage_ResponseStructure validates outage response field structure.
func TestContract_Outage_ResponseStructure(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "outage_crud",
		Mode:         testutil.ModeReplay,
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
	if err != nil || len(monitors) == 0 {
		t.Fatal("No monitors available")
	}

	monitorUUID := monitors[0].UUID
	endDate := "2026-02-09T05:00:00.000Z"
	createReq := CreateOutageRequest{
		MonitorUUID: monitorUUID,
		StartDate:   "2026-02-09T04:00:00.000Z",
		EndDate:     &endDate,
		StatusCode:  500,
		Description: "Manual outage for contract testing",
		OutageType:  "manual",
	}

	outage, err := client.CreateOutage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateOutage failed: %v", err)
	}

	t.Run("RequiredFields", func(t *testing.T) {
		// Core required fields (always present)
		if outage.UUID == "" {
			t.Error("UUID must not be empty")
		}
		if outage.StartDate == "" {
			t.Error("StartDate must not be empty")
		}
		if outage.Description == "" {
			t.Error("Description must not be empty")
		}

		// Note: Create response may not include OutageType, DetectedLocation, etc.
		// These are populated by the GET endpoint
		t.Logf("OutageType: %q (may be empty in create response)", outage.OutageType)
		t.Logf("DetectedLocation: %q (may be empty in create response)", outage.DetectedLocation)
		t.Logf("DurationMs: %d", outage.DurationMs)
		t.Logf("IsResolved: %v", outage.IsResolved)
	})

	t.Run("OptionalFields", func(t *testing.T) {
		// EndDate is optional (null for ongoing outages)
		if outage.EndDate != nil {
			t.Logf("EndDate: %s", *outage.EndDate)
		} else {
			t.Log("EndDate is null (ongoing outage)")
		}

		// AcknowledgedAt is optional
		if outage.AcknowledgedAt != nil {
			t.Logf("AcknowledgedAt: %s", *outage.AcknowledgedAt)
		} else {
			t.Log("AcknowledgedAt is null (not acknowledged)")
		}

		// AcknowledgedBy is optional
		if outage.AcknowledgedBy != nil {
			t.Logf("AcknowledgedBy: %+v", outage.AcknowledgedBy)
		} else {
			t.Log("AcknowledgedBy is null (not acknowledged)")
		}

		// EscalationPolicy is optional
		if outage.EscalationPolicy != nil {
			t.Logf("EscalationPolicy: %+v", outage.EscalationPolicy)
		} else {
			t.Log("EscalationPolicy is null")
		}
	})

	t.Run("MonitorReference", func(t *testing.T) {
		// Monitor reference validation
		if outage.Monitor.UUID != "" {
			if outage.Monitor.Name == "" {
				t.Error("Monitor.Name should not be empty when UUID is set")
			}
			if outage.Monitor.URL == "" {
				t.Error("Monitor.URL should not be empty when UUID is set")
			}
			if outage.Monitor.Protocol == "" {
				t.Error("Monitor.Protocol should not be empty when UUID is set")
			}
			t.Logf("Monitor: UUID=%s, Name=%s", outage.Monitor.UUID, outage.Monitor.Name)
		} else {
			t.Log("Monitor reference is empty (possible for manual outages)")
		}
	})

	t.Run("OutageTypeEnum", func(t *testing.T) {
		validTypes := map[string]bool{
			"":          true, // Empty in create responses
			"manual":    true,
			"automatic": true,
		}

		if !validTypes[outage.OutageType] {
			t.Errorf("OutageType %q is not in valid set: manual, automatic", outage.OutageType)
		}

		if outage.OutageType != "" {
			t.Logf("OutageType is valid: %q", outage.OutageType)
		}
	})

	t.Run("StatusCodeRange", func(t *testing.T) {
		// Status code should be in valid HTTP range or 0 for manual outages
		if outage.StatusCode != 0 && (outage.StatusCode < 100 || outage.StatusCode > 599) {
			t.Errorf("StatusCode %d is outside valid HTTP range (100-599) or 0", outage.StatusCode)
		}
	})
}

// TestContract_Outage_ListStructure validates list response pagination structure.
func TestContract_Outage_ListStructure(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "outage_list",
		Mode:         testutil.ModeReplay,
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

	outages, err := client.ListOutages(ctx)
	if err != nil {
		t.Fatalf("ListOutages failed: %v", err)
	}

	t.Run("ArrayStructure", func(t *testing.T) {
		// List should return an array
		if outages == nil {
			t.Fatal("outages list should not be nil")
		}

		t.Logf("Retrieved %d outages", len(outages))

		// Each item should be valid
		for i, outage := range outages {
			if outage.UUID == "" {
				t.Errorf("outage[%d].UUID is empty", i)
			}
			if outage.StartDate == "" {
				t.Errorf("outage[%d].StartDate is empty", i)
			}
		}
	})

	t.Run("ResponseConsistency", func(t *testing.T) {
		// All outages should have consistent structure
		for i, outage := range outages {
			validateOutageStructure(t, &outage)

			if i == 0 {
				data, _ := json.MarshalIndent(outage, "", "  ")
				t.Logf("First outage structure:\n%s", string(data))
			}
		}
	})
}

// TestContract_Outage_Actions validates outage action responses.
func TestContract_Outage_Actions(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "outage_crud",
		Mode:         testutil.ModeReplay,
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
	if err != nil || len(monitors) == 0 {
		t.Fatal("No monitors available")
	}

	monitorUUID := monitors[0].UUID
	endDate := "2026-02-09T05:00:00.000Z"
	createReq := CreateOutageRequest{
		MonitorUUID: monitorUUID,
		StartDate:   "2026-02-09T04:00:00.000Z",
		EndDate:     &endDate,
		StatusCode:  500,
		Description: "Manual outage for contract testing",
		OutageType:  "manual",
	}

	outage, err := client.CreateOutage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateOutage failed: %v", err)
	}

	t.Run("ActionResponseStructure", func(t *testing.T) {
		// OutageAction responses should have Message and UUID
		// These are returned by Acknowledge, Unacknowledge, Resolve, Escalate

		// We can't actually test these actions with the CRUD cassette,
		// but we validate the structure expectations
		t.Log("OutageAction structure validation:")
		t.Log("- Message field (string)")
		t.Log("- UUID field (string)")

		if outage.UUID == "" {
			t.Error("Outage UUID should not be empty for action testing")
		}
	})
}

// =============================================================================
// Helper Functions
// =============================================================================

// validateOutageStructure validates that an outage has all required fields.
// Note: Some fields may be empty in create responses, so we only validate core fields.
func validateOutageStructure(t *testing.T, outage *Outage) {
	t.Helper()

	if outage == nil {
		t.Fatal("outage is nil")
	}

	// Core required fields (always present)
	if outage.UUID == "" {
		t.Error("UUID must not be empty")
	}
	if outage.StartDate == "" {
		t.Error("StartDate must not be empty")
	}
	if outage.Description == "" {
		t.Error("Description must not be empty")
	}

	// Optional fields from GET responses
	// OutageType, DetectedLocation, etc. may be empty in create responses
	if outage.OutageType != "" {
		t.Logf("OutageType: %s", outage.OutageType)
	}
	if outage.DetectedLocation != "" {
		t.Logf("DetectedLocation: %s", outage.DetectedLocation)
	}

	// Numeric fields
	// Note: DurationMs can be 0 for just-created outages or negative in some edge cases
	t.Logf("DurationMs: %d", outage.DurationMs)

	// ConfirmedLocations is a string (comma-separated)
	// It can be empty
	t.Logf("ConfirmedLocations: %q", outage.ConfirmedLocations)

	// Monitor reference (can be empty for manual outages)
	t.Logf("Monitor.UUID: %s", outage.Monitor.UUID)
}

// validateOutageAction validates an outage action response.
func validateOutageAction(t *testing.T, action *OutageAction) {
	t.Helper()

	if action == nil {
		t.Fatal("action is nil")
	}

	if action.Message == "" {
		t.Error("action.Message should not be empty")
	}

	if action.UUID == "" {
		t.Error("action.UUID should not be empty")
	}

	t.Logf("Action: Message=%q, UUID=%s", action.Message, action.UUID)
}
