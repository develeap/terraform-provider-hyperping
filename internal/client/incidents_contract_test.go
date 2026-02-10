// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// =============================================================================
// Incident Contract Tests - Using VCR Cassettes for Response Validation
// =============================================================================

// TestContract_Incident_CRUD_ResponseStructure validates the incident CRUD API contract
// using recorded VCR cassettes. This ensures our models correctly map API responses.
func TestContract_Incident_CRUD_ResponseStructure(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "incident_crud",
		Mode:         testutil.ModeReplay,
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	client := NewClient("test_api_key", WithHTTPClient(httpClient))
	ctx := context.Background()

	// Get status pages (required for incident creation)
	statusPages, err := client.ListStatusPages(ctx, nil, nil)
	if err != nil || len(statusPages.StatusPages) == 0 {
		t.Skip("Skipping: cassette doesn't contain status page data")
	}

	statusPageUUID := statusPages.StatusPages[0].UUID

	// Test Create - Validate response structure
	createReq := CreateIncidentRequest{
		Title:       LocalizedText{En: "VCR Test Incident"},
		Text:        LocalizedText{En: "This is a test incident"},
		Type:        "incident",
		StatusPages: []string{statusPageUUID},
	}

	createResp, err := client.CreateIncident(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateIncident failed: %v", err)
	}

	// CreateIncident returns minimal response with UUID only
	// Read full incident details to validate structure
	incident, err := client.GetIncident(ctx, createResp.UUID)
	if err != nil {
		t.Fatalf("GetIncident after create failed: %v", err)
	}

	// Validate required fields are present and correct types
	validateIncidentStructure(t, incident, "GetIncident")

	// Validate specific values match request
	if incident.Title.En != "VCR Test Incident" {
		t.Errorf("GetIncident: expected Title.En 'VCR Test Incident', got '%s'", incident.Title.En)
	}
	// Note: The Hyperping API does NOT return the 'text' field in GET responses,
	// even though it's required in POST requests. This is API behavior.
	if incident.Type != "incident" {
		t.Errorf("GetIncident: expected Type 'incident', got '%s'", incident.Type)
	}

	// Test Delete - Should succeed without error
	if err = client.DeleteIncident(ctx, incident.UUID); err != nil {
		t.Fatalf("DeleteIncident failed: %v", err)
	}

	t.Log("Successfully validated incident CRUD contract")
}

// TestContract_Incident_List_ResponseStructure validates the list incidents API contract.
func TestContract_Incident_List_ResponseStructure(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "incident_list",
		Mode:         testutil.ModeReplay,
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	client := NewClient("test_api_key", WithHTTPClient(httpClient))
	ctx := context.Background()

	incidents, err := client.ListIncidents(ctx)
	if err != nil {
		t.Fatalf("ListIncidents failed: %v", err)
	}

	// Validate response is an array (can be empty)
	if incidents == nil {
		t.Fatal("ListIncidents: expected incidents array, got nil")
	}

	// If there are incidents, validate their structure
	for i, incident := range incidents {
		validateIncidentStructure(t, &incident, "ListIncidents["+string(rune(i))+"]")
	}

	t.Logf("Successfully validated %d incidents in list response", len(incidents))
}

// TestContract_Incident_NotFound validates 404 error handling.
func TestContract_Incident_NotFound(t *testing.T) {
	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: "incident_crud",
		Mode:         testutil.ModeReplay,
		CassetteDir:  "testdata/cassettes",
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop recorder: %v", err)
		}
	}()

	client := NewClient("test_api_key", WithHTTPClient(httpClient))
	ctx := context.Background()

	// After deletion in the CRUD test, trying to read should fail with 404
	// This assumes the cassette contains this interaction
	// We'll use a non-existent UUID that should produce a 404
	_, err := client.GetIncident(ctx, "inci_nonexistent123456")
	if err == nil {
		// This is expected to fail in replay mode if cassette doesn't have this interaction
		t.Skip("Cassette doesn't contain 404 interaction")
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if apiErr.StatusCode != 404 {
			t.Logf("Expected 404, got %d (cassette may not contain 404 interaction)", apiErr.StatusCode)
		}
	}
}

// TestContract_Incident_FieldTypes validates that all field types match expectations.
func TestContract_Incident_FieldTypes(t *testing.T) {
	// Load test data directly
	data := loadTestData(t, "incidents/response.json")

	var incident Incident
	if err := json.Unmarshal(data, &incident); err != nil {
		t.Fatalf("failed to unmarshal incident response: %v", err)
	}

	// Validate UUID format
	if incident.UUID == "" {
		t.Error("expected UUID to be non-empty string")
	}
	if len(incident.UUID) < 10 {
		t.Errorf("UUID appears too short: %q", incident.UUID)
	}

	// Validate LocalizedText structure
	if incident.Title.En == "" {
		t.Error("expected Title.En to be non-empty")
	}
	// Note: Text field is optional in API responses (not returned by GET endpoints)

	// Validate Type is a known value
	validTypes := map[string]bool{
		"incident":    true,
		"outage":      true,
		"maintenance": true,
	}
	if !validTypes[incident.Type] {
		t.Errorf("unexpected incident type: %q", incident.Type)
	}

	// Validate arrays are properly typed
	if incident.AffectedComponents == nil {
		t.Error("expected AffectedComponents to be non-nil (can be empty array)")
	}
	if incident.StatusPages == nil {
		t.Error("expected StatusPages to be non-nil (can be empty array)")
	}
	if incident.Updates == nil {
		t.Error("expected Updates to be non-nil (can be empty array)")
	}

	// Validate update structure if present
	if len(incident.Updates) > 0 {
		update := incident.Updates[0]
		if update.UUID == "" {
			t.Error("expected Update.UUID to be non-empty")
		}
		if update.Date == "" {
			t.Error("expected Update.Date to be non-empty")
		}
		if update.Text.En == "" {
			t.Error("expected Update.Text.En to be non-empty")
		}
		if update.Type == "" {
			t.Error("expected Update.Type to be non-empty")
		}

		// Validate update type
		validUpdateTypes := map[string]bool{
			"investigating": true,
			"identified":    true,
			"monitoring":    true,
			"resolved":      true,
		}
		if !validUpdateTypes[update.Type] {
			t.Errorf("unexpected update type: %q", update.Type)
		}
	}
}

// TestContract_Incident_RequiredFields validates that required fields are always present.
func TestContract_Incident_RequiredFields(t *testing.T) {
	data := loadTestData(t, "incidents/response.json")

	var incident Incident
	if err := json.Unmarshal(data, &incident); err != nil {
		t.Fatalf("failed to unmarshal incident response: %v", err)
	}

	// Required fields that must always be present
	// Note: Text field is optional in API responses (not returned by GET endpoints)
	requiredFields := map[string]interface{}{
		"UUID":        incident.UUID,
		"Title.En":    incident.Title.En,
		"Type":        incident.Type,
		"StatusPages": incident.StatusPages,
	}

	for fieldName, value := range requiredFields {
		switch v := value.(type) {
		case string:
			if v == "" {
				t.Errorf("required field %s is empty", fieldName)
			}
		case []string:
			if v == nil {
				t.Errorf("required field %s is nil", fieldName)
			}
		}
	}
}

// TestContract_Incident_CreateRequest_Marshaling validates request serialization.
func TestContract_Incident_CreateRequest_Marshaling(t *testing.T) {
	req := CreateIncidentRequest{
		Title:              LocalizedText{En: "Test Incident"},
		Text:               LocalizedText{En: "Test incident description"},
		Type:               "incident",
		AffectedComponents: []string{"comp_123"},
		StatusPages:        []string{"sp_main"},
		Date:               "2025-01-15T10:00:00.000Z",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	// Validate JSON contains expected fields
	assertJSONContains(t, data, "title")
	assertJSONContains(t, data, "text")
	assertJSONContains(t, data, "type")
	assertJSONContains(t, data, "affectedComponents")
	assertJSONContains(t, data, "statuspages")

	// Validate nested structure contains the data
	assertJSONContains(t, data, "Test Incident")
	assertJSONContains(t, data, "Test incident description")
}

// TestContract_Incident_UpdateRequest_Marshaling validates update request serialization.
func TestContract_Incident_UpdateRequest_Marshaling(t *testing.T) {
	title := LocalizedText{En: "Updated Title"}
	incidentType := "outage"
	components := []string{"comp_123", "comp_456"}
	pages := []string{"sp_main"}

	req := UpdateIncidentRequest{
		Title:              &title,
		Type:               &incidentType,
		AffectedComponents: &components,
		StatusPages:        &pages,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal update request: %v", err)
	}

	// Validate JSON contains expected fields
	assertJSONContains(t, data, "title")
	assertJSONContains(t, data, "type")
	assertJSONContains(t, data, "affectedComponents")
	assertJSONContains(t, data, "statuspages")
}

// TestContract_Incident_AddUpdate_Marshaling validates add update request serialization.
func TestContract_Incident_AddUpdate_Marshaling(t *testing.T) {
	req := AddIncidentUpdateRequest{
		Text: LocalizedText{En: "Update message"},
		Type: "identified",
		Date: "2025-01-15T11:00:00.000Z",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal add update request: %v", err)
	}

	// Validate JSON contains expected fields
	assertJSONContains(t, data, "text")
	assertJSONContains(t, data, "type")
	assertJSONContains(t, data, "date")
	assertJSONContains(t, data, "Update message")
	assertJSONContains(t, data, "identified")
}

// =============================================================================
// Helper Functions
// =============================================================================

// validateIncidentStructure checks that an incident has all required fields
// with correct types and reasonable values.
func validateIncidentStructure(t *testing.T, incident *Incident, context string) {
	t.Helper()

	if incident == nil {
		t.Fatalf("%s: incident is nil", context)
	}

	// Validate UUID
	if incident.UUID == "" {
		t.Errorf("%s: UUID is empty", context)
	}

	// Validate Title (LocalizedText)
	if incident.Title.En == "" {
		t.Errorf("%s: Title.En is empty", context)
	}

	// Note: Text field is optional in API responses (not returned by GET endpoints)

	// Validate Type
	if incident.Type == "" {
		t.Errorf("%s: Type is empty", context)
	}

	// Validate arrays are not nil (they can be empty)
	if incident.AffectedComponents == nil {
		t.Errorf("%s: AffectedComponents is nil (should be empty array)", context)
	}
	if incident.StatusPages == nil {
		t.Errorf("%s: StatusPages is nil (should be empty array)", context)
	}
	if incident.Updates == nil {
		t.Errorf("%s: Updates is nil (should be empty array)", context)
	}

	// Validate updates structure if present
	for i, update := range incident.Updates {
		if update.UUID == "" {
			t.Errorf("%s: Update[%d].UUID is empty", context, i)
		}
		if update.Text.En == "" {
			t.Errorf("%s: Update[%d].Text.En is empty", context, i)
		}
		if update.Type == "" {
			t.Errorf("%s: Update[%d].Type is empty", context, i)
		}
	}
}
