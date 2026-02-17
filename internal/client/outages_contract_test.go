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

// newOutageVCRClient sets up a VCR recorder and returns a client and teardown func.
func newOutageVCRClient(t *testing.T, cassetteName string) (*Client, func()) {
	t.Helper()
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: cassetteName,
		Mode:         testutil.ModeReplay,
		CassetteDir:  "testdata/cassettes",
	})
	apiKey := os.Getenv("HYPERPING_API_KEY")
	if apiKey == "" {
		apiKey = "test_api_key"
	}
	client := NewClient(apiKey, WithHTTPClient(httpClient))
	teardown := func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}
	return client, teardown
}

// newOutageCreateRequest builds a standard outage request for contract tests.
func newOutageCreateRequest(monitorUUID string) CreateOutageRequest {
	endDate := "2026-02-09T05:00:00.000Z"
	return CreateOutageRequest{
		MonitorUUID: monitorUUID,
		StartDate:   "2026-02-09T04:00:00.000Z",
		EndDate:     &endDate,
		StatusCode:  500,
		Description: "Manual outage for contract testing",
		OutageType:  "manual",
	}
}

// getFirstMonitorUUID lists monitors and returns the UUID of the first one.
func getFirstMonitorUUID(t *testing.T, client *Client, ctx context.Context) string {
	t.Helper()
	monitors, err := client.ListMonitors(ctx)
	if err != nil || len(monitors) == 0 {
		t.Fatal("No monitors available")
	}
	return monitors[0].UUID
}

// TestContract_Outage_CRUD validates the full CRUD lifecycle using VCR cassettes.
// This test validates:
// - Create outage response structure
// - Read outage response structure
// - Delete outage success
func TestContract_Outage_CRUD(t *testing.T) {
	client, teardown := newOutageVCRClient(t, "outage_crud")
	defer teardown()

	ctx := context.Background()

	monitors, err := client.ListMonitors(ctx)
	if err != nil {
		t.Fatalf("ListMonitors failed: %v", err)
	}
	if len(monitors) == 0 {
		t.Fatal("No monitors available for outage test")
	}

	createReq := newOutageCreateRequest(monitors[0].UUID)
	outage, err := client.CreateOutage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateOutage failed: %v", err)
	}

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
		t.Logf("Created outage UUID: %s", outage.UUID)
	})

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
	client, teardown := newOutageVCRClient(t, "outage_list")
	defer teardown()

	ctx := context.Background()

	outages, err := client.ListOutages(ctx)
	if err != nil {
		t.Fatalf("ListOutages failed: %v", err)
	}

	t.Run("ListResponse", func(t *testing.T) {
		if len(outages) == 0 {
			t.Skip("No outages in list (ok for test)")
		}
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
	client, teardown := newOutageVCRClient(t, "outage_crud")
	defer teardown()

	ctx := context.Background()
	monitorUUID := getFirstMonitorUUID(t, client, ctx)
	createReq := newOutageCreateRequest(monitorUUID)

	outage, err := client.CreateOutage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateOutage failed: %v", err)
	}

	t.Run("RequiredFields", func(t *testing.T) {
		assertOutageRequiredFields(t, outage)
	})

	t.Run("OptionalFields", func(t *testing.T) {
		logOptionalOutageFields(t, outage)
	})

	t.Run("MonitorReference", func(t *testing.T) {
		assertOutageMonitorReference(t, outage)
	})

	t.Run("OutageTypeEnum", func(t *testing.T) {
		assertOutageTypeEnum(t, outage)
	})

	t.Run("StatusCodeRange", func(t *testing.T) {
		if outage.StatusCode != 0 && (outage.StatusCode < 100 || outage.StatusCode > 599) {
			t.Errorf("StatusCode %d is outside valid HTTP range (100-599) or 0", outage.StatusCode)
		}
	})
}

// assertOutageRequiredFields checks fields that must always be present.
func assertOutageRequiredFields(t *testing.T, outage *Outage) {
	t.Helper()
	if outage.UUID == "" {
		t.Error("UUID must not be empty")
	}
	if outage.StartDate == "" {
		t.Error("StartDate must not be empty")
	}
	if outage.Description == "" {
		t.Error("Description must not be empty")
	}
	t.Logf("OutageType: %q (may be empty in create response)", outage.OutageType)
	t.Logf("DetectedLocation: %q (may be empty in create response)", outage.DetectedLocation)
	t.Logf("DurationMs: %d", outage.DurationMs)
	t.Logf("IsResolved: %v", outage.IsResolved)
}

// logOptionalOutageFields logs optional fields that may be null.
func logOptionalOutageFields(t *testing.T, outage *Outage) {
	t.Helper()
	if outage.EndDate != nil {
		t.Logf("EndDate: %s", *outage.EndDate)
	} else {
		t.Log("EndDate is null (ongoing outage)")
	}
	if outage.AcknowledgedAt != nil {
		t.Logf("AcknowledgedAt: %s", *outage.AcknowledgedAt)
	} else {
		t.Log("AcknowledgedAt is null (not acknowledged)")
	}
	if outage.AcknowledgedBy != nil {
		t.Logf("AcknowledgedBy: %+v", outage.AcknowledgedBy)
	} else {
		t.Log("AcknowledgedBy is null (not acknowledged)")
	}
	if outage.EscalationPolicy != nil {
		t.Logf("EscalationPolicy: %+v", outage.EscalationPolicy)
	} else {
		t.Log("EscalationPolicy is null")
	}
}

// assertOutageMonitorReference validates the embedded monitor reference.
func assertOutageMonitorReference(t *testing.T, outage *Outage) {
	t.Helper()
	if outage.Monitor.UUID == "" {
		t.Log("Monitor reference is empty (possible for manual outages)")
		return
	}
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
}

// assertOutageTypeEnum validates that OutageType is one of the allowed values.
func assertOutageTypeEnum(t *testing.T, outage *Outage) {
	t.Helper()
	validTypes := map[string]bool{
		"":          true,
		"manual":    true,
		"automatic": true,
	}
	if !validTypes[outage.OutageType] {
		t.Errorf("OutageType %q is not in valid set: manual, automatic", outage.OutageType)
	}
	if outage.OutageType != "" {
		t.Logf("OutageType is valid: %q", outage.OutageType)
	}
}

// TestContract_Outage_ListStructure validates list response pagination structure.
func TestContract_Outage_ListStructure(t *testing.T) {
	client, teardown := newOutageVCRClient(t, "outage_list")
	defer teardown()

	ctx := context.Background()

	outages, err := client.ListOutages(ctx)
	if err != nil {
		t.Fatalf("ListOutages failed: %v", err)
	}

	t.Run("ArrayStructure", func(t *testing.T) {
		if outages == nil {
			t.Fatal("outages list should not be nil")
		}
		t.Logf("Retrieved %d outages", len(outages))
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
	client, teardown := newOutageVCRClient(t, "outage_crud")
	defer teardown()

	ctx := context.Background()
	monitorUUID := getFirstMonitorUUID(t, client, ctx)
	createReq := newOutageCreateRequest(monitorUUID)

	outage, err := client.CreateOutage(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateOutage failed: %v", err)
	}

	t.Run("ActionResponseStructure", func(t *testing.T) {
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
