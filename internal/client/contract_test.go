// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// loadTestData loads a JSON file from the testdata directory.
func loadTestData(t *testing.T, filename string) []byte {
	t.Helper()
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to load test data %s: %v", filename, err)
	}
	return data
}

// assertJSONContains checks that the JSON contains the expected field name.
func assertJSONContains(t *testing.T, jsonData []byte, fieldName string) {
	t.Helper()
	if !strings.Contains(string(jsonData), `"`+fieldName+`"`) {
		t.Errorf("expected JSON to contain field %q, but it doesn't.\nJSON: %s", fieldName, string(jsonData))
	}
}

// assertJSONNotContains checks that the JSON does NOT contain the given field name.
func assertJSONNotContains(t *testing.T, jsonData []byte, fieldName string) {
	t.Helper()
	if strings.Contains(string(jsonData), `"`+fieldName+`"`) {
		t.Errorf("expected JSON to NOT contain field %q, but it does.\nJSON: %s", fieldName, string(jsonData))
	}
}

// =============================================================================
// Monitor Contract Tests
// =============================================================================

func TestContractMonitor_UnmarshalResponse(t *testing.T) {
	data := loadTestData(t, "monitors/response.json")

	var monitor Monitor
	err := json.Unmarshal(data, &monitor)
	if err != nil {
		t.Fatalf("failed to unmarshal monitor response: %v", err)
	}

	// Verify critical fields are mapped correctly
	if monitor.UUID != "mon_OYKr5fpSDHqbP2" {
		t.Errorf("expected UUID 'mon_OYKr5fpSDHqbP2', got %q", monitor.UUID)
	}
	if monitor.Name != "API Monitor" {
		t.Errorf("expected Name 'API Monitor', got %q", monitor.Name)
	}
	if monitor.Protocol != "http" {
		t.Errorf("expected Protocol 'http', got %q", monitor.Protocol)
	}
	if monitor.HTTPMethod != "GET" {
		t.Errorf("expected HTTPMethod 'GET', got %q", monitor.HTTPMethod)
	}
	if monitor.CheckFrequency != 30 {
		t.Errorf("expected CheckFrequency 30, got %d", monitor.CheckFrequency)
	}
	if !monitor.FollowRedirects {
		t.Error("expected FollowRedirects true")
	}
	if string(monitor.ExpectedStatusCode) != "2xx" {
		t.Errorf("expected ExpectedStatusCode '2xx', got %q", monitor.ExpectedStatusCode)
	}
	if len(monitor.Regions) != 4 {
		t.Errorf("expected 4 regions, got %d", len(monitor.Regions))
	}
	if len(monitor.RequestHeaders) != 2 {
		t.Errorf("expected 2 request headers, got %d", len(monitor.RequestHeaders))
	}
	if len(monitor.RequestHeaders) > 0 {
		if monitor.RequestHeaders[0].Name != "Authorization" {
			t.Errorf("expected first header name 'Authorization', got %q", monitor.RequestHeaders[0].Name)
		}
	}
}

func TestContractMonitor_MarshalCreateRequest(t *testing.T) {
	req := CreateMonitorRequest{
		Name:               "Test Monitor",
		URL:                "https://example.com",
		Protocol:           "http",
		HTTPMethod:         "POST",
		CheckFrequency:     60,
		Regions:            []string{"london", "paris"},
		RequestHeaders:     []RequestHeader{{Name: "Auth", Value: "Bearer token"}},
		RequestBody:        strPtr(`{"test": true}`),
		FollowRedirects:    boolPtr(false),
		ExpectedStatusCode: "201",
		Paused:             false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal create request: %v", err)
	}

	// Must use snake_case field names
	assertJSONContains(t, data, "http_method")
	assertJSONContains(t, data, "check_frequency")
	assertJSONContains(t, data, "request_headers")
	assertJSONContains(t, data, "request_body")
	assertJSONContains(t, data, "follow_redirects")
	assertJSONContains(t, data, "expected_status_code")
	assertJSONContains(t, data, "protocol")

	// Must NOT use old camelCase field names
	assertJSONNotContains(t, data, "method")
	assertJSONNotContains(t, data, "frequency")
	assertJSONNotContains(t, data, "headers")
	assertJSONNotContains(t, data, "body")
	assertJSONNotContains(t, data, "expectedStatus")
	assertJSONNotContains(t, data, "followRedirects")
}

func TestContractMonitor_UnmarshalListResponse(t *testing.T) {
	data := loadTestData(t, "monitors/list_response.json")

	var monitors []Monitor
	err := json.Unmarshal(data, &monitors)
	if err != nil {
		t.Fatalf("failed to unmarshal monitor list: %v", err)
	}

	if len(monitors) != 2 {
		t.Errorf("expected 2 monitors, got %d", len(monitors))
	}

	// Verify first monitor
	if monitors[0].UUID != "mon_OYKr5fpSDHqbP2" {
		t.Errorf("expected first monitor UUID 'mon_OYKr5fpSDHqbP2', got %q", monitors[0].UUID)
	}
	if monitors[0].Protocol != "http" {
		t.Errorf("expected first monitor Protocol 'http', got %q", monitors[0].Protocol)
	}
}

// =============================================================================
// Maintenance Contract Tests
// =============================================================================

func TestContractMaintenance_UnmarshalResponse(t *testing.T) {
	data := loadTestData(t, "maintenance/response.json")

	var maint Maintenance
	err := json.Unmarshal(data, &maint)
	if err != nil {
		t.Fatalf("failed to unmarshal maintenance response: %v", err)
	}

	if maint.UUID != "mw_TY2vFNUbdzdskD" {
		t.Errorf("expected UUID 'mw_TY2vFNUbdzdskD', got %q", maint.UUID)
	}
	if maint.Name != "Scheduled API Infrastructure Upgrade" {
		t.Errorf("expected Name 'Scheduled API Infrastructure Upgrade', got %q", maint.Name)
	}
	if maint.Title.En != "Planned API Gateway Maintenance" {
		t.Errorf("expected Title.En 'Planned API Gateway Maintenance', got %q", maint.Title.En)
	}
	if maint.StartDate == nil {
		t.Error("expected StartDate to be set")
	}
	if maint.EndDate == nil {
		t.Error("expected EndDate to be set")
	}
	if len(maint.Monitors) != 2 {
		t.Errorf("expected 2 monitors, got %d", len(maint.Monitors))
	}
	if len(maint.StatusPages) != 1 {
		t.Errorf("expected 1 status page, got %d", len(maint.StatusPages))
	}
}

func TestContractMaintenance_MarshalCreateRequest(t *testing.T) {
	req := CreateMaintenanceRequest{
		Name:                "Test Maintenance",
		Title:               LocalizedText{En: "Test Title"},
		Text:                LocalizedText{En: "Test Description"},
		StartDate:           "2025-12-20T02:00:00.000Z",
		EndDate:             "2025-12-20T06:00:00.000Z",
		Monitors:            []string{"mon_123"},
		StatusPages:         []string{"sp_456"},
		NotificationOption:  "scheduled",
		NotificationMinutes: intPtr(60),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal maintenance create request: %v", err)
	}

	// Must use correct field names
	assertJSONContains(t, data, "start_date")
	assertJSONContains(t, data, "end_date")
	assertJSONContains(t, data, "monitors")
	assertJSONContains(t, data, "statuspages")
	assertJSONContains(t, data, "notificationOption")
	assertJSONContains(t, data, "notificationMinutes")

	// Must NOT use old field names
	assertJSONNotContains(t, data, "scheduledStart")
	assertJSONNotContains(t, data, "scheduledEnd")
	assertJSONNotContains(t, data, "monitorUuids")
}

// =============================================================================
// Incident Contract Tests
// =============================================================================

func TestContractIncident_UnmarshalResponse(t *testing.T) {
	data := loadTestData(t, "incidents/response.json")

	var incident Incident
	err := json.Unmarshal(data, &incident)
	if err != nil {
		t.Fatalf("failed to unmarshal incident response: %v", err)
	}

	if incident.UUID != "inci_cDAqydvnNnyc8D" {
		t.Errorf("expected UUID 'inci_cDAqydvnNnyc8D', got %q", incident.UUID)
	}
	if incident.Title.En != "API Service Degradation" {
		t.Errorf("expected Title.En 'API Service Degradation', got %q", incident.Title.En)
	}
	if incident.Text.En == "" {
		t.Error("expected Text.En to be set")
	}
	if incident.Type != "incident" {
		t.Errorf("expected Type 'incident', got %q", incident.Type)
	}
	if len(incident.AffectedComponents) != 2 {
		t.Errorf("expected 2 affected components, got %d", len(incident.AffectedComponents))
	}
	if len(incident.StatusPages) != 2 {
		t.Errorf("expected 2 status pages, got %d", len(incident.StatusPages))
	}
	if len(incident.Updates) != 2 {
		t.Errorf("expected 2 updates, got %d", len(incident.Updates))
	}
	if len(incident.Updates) > 0 {
		if incident.Updates[0].Type != "identified" {
			t.Errorf("expected first update type 'identified', got %q", incident.Updates[0].Type)
		}
	}
}

func TestContractIncident_MarshalCreateRequest(t *testing.T) {
	req := CreateIncidentRequest{
		Title:              LocalizedText{En: "Service Outage"},
		Text:               LocalizedText{En: "We are experiencing an outage."},
		Type:               "outage",
		AffectedComponents: []string{"comp_123"},
		StatusPages:        []string{"sp_main"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal incident create request: %v", err)
	}

	// Must use correct field names
	assertJSONContains(t, data, "title")
	assertJSONContains(t, data, "text")
	assertJSONContains(t, data, "type")
	assertJSONContains(t, data, "affectedComponents")
	assertJSONContains(t, data, "statuspages")

	// Must NOT use old field names
	assertJSONNotContains(t, data, "severity")
	assertJSONNotContains(t, data, "message")
	assertJSONNotContains(t, data, "monitorUuids")
}

func TestContractIncident_UnmarshalListResponse(t *testing.T) {
	data := loadTestData(t, "incidents/list_response.json")

	var incidents []Incident
	err := json.Unmarshal(data, &incidents)
	if err != nil {
		t.Fatalf("failed to unmarshal incident list: %v", err)
	}

	if len(incidents) != 2 {
		t.Errorf("expected 2 incidents, got %d", len(incidents))
	}

	// Verify first incident has type "outage"
	if incidents[0].Type != "outage" {
		t.Errorf("expected first incident type 'outage', got %q", incidents[0].Type)
	}
}

// =============================================================================
// Helper functions
// =============================================================================

func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
