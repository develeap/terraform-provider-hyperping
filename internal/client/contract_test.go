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

func TestContractMaintenance_UnmarshalListResponse(t *testing.T) {
	data := loadTestData(t, "maintenance/list_response.json")

	var maintenances []Maintenance
	err := json.Unmarshal(data, &maintenances)
	if err != nil {
		t.Fatalf("failed to unmarshal maintenance list: %v", err)
	}

	if len(maintenances) != 2 {
		t.Errorf("expected 2 maintenance windows, got %d", len(maintenances))
	}

	// Verify first maintenance has expected UUID
	if maintenances[0].UUID != "mw_TY2vFNUbdzdskD" {
		t.Errorf("expected first maintenance UUID 'mw_TY2vFNUbdzdskD', got %q", maintenances[0].UUID)
	}

	// Verify first maintenance has name
	if maintenances[0].Name != "Scheduled API Upgrade" {
		t.Errorf("expected first maintenance name 'Scheduled API Upgrade', got %q", maintenances[0].Name)
	}

	// Verify second maintenance has expected UUID
	if maintenances[1].UUID != "mw_ABC123xyz" {
		t.Errorf("expected second maintenance UUID 'mw_ABC123xyz', got %q", maintenances[1].UUID)
	}
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
// Healthcheck Contract Tests
// =============================================================================

func TestContractHealthcheck_UnmarshalResponse(t *testing.T) {
	data := loadTestData(t, "healthchecks/response.json")

	var healthcheck Healthcheck
	err := json.Unmarshal(data, &healthcheck)
	if err != nil {
		t.Fatalf("failed to unmarshal healthcheck response: %v", err)
	}

	if healthcheck.UUID != "tok_hy4GCxOqbwahr4ctj0cNOIIp" {
		t.Errorf("expected UUID 'tok_hy4GCxOqbwahr4ctj0cNOIIp', got %q", healthcheck.UUID)
	}
	if healthcheck.Name != "vcr-test-healthcheck" {
		t.Errorf("expected Name 'vcr-test-healthcheck', got %q", healthcheck.Name)
	}
	if healthcheck.PeriodValue == nil || *healthcheck.PeriodValue != 60 {
		if healthcheck.PeriodValue == nil {
			t.Error("expected PeriodValue to be set")
		} else {
			t.Errorf("expected PeriodValue 60, got %d", *healthcheck.PeriodValue)
		}
	}
	if healthcheck.PeriodType != "seconds" {
		t.Errorf("expected PeriodType 'seconds', got %q", healthcheck.PeriodType)
	}
	if healthcheck.GracePeriodValue != 300 {
		t.Errorf("expected GracePeriodValue 300, got %d", healthcheck.GracePeriodValue)
	}
	if healthcheck.PingURL != "https://hc.hyperping.io/tok_hy4GCxOqbwahr4ctj0cNOIIp" {
		t.Errorf("expected PingURL to be set, got %q", healthcheck.PingURL)
	}
}

func TestContractHealthcheck_MarshalCreateRequest(t *testing.T) {
	periodValue := 60
	periodType := "seconds"
	req := CreateHealthcheckRequest{
		Name:             "test-healthcheck",
		PeriodValue:      &periodValue,
		PeriodType:       &periodType,
		GracePeriodValue: 300,
		GracePeriodType:  "seconds",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal healthcheck create request: %v", err)
	}

	// Must use snake_case field names
	assertJSONContains(t, data, "name")
	assertJSONContains(t, data, "period_value")
	assertJSONContains(t, data, "period_type")
	assertJSONContains(t, data, "grace_period_value")
	assertJSONContains(t, data, "grace_period_type")
}

func TestContractHealthcheck_UnmarshalListResponse(t *testing.T) {
	data := loadTestData(t, "healthchecks/list_response.json")

	var response struct {
		Healthchecks []Healthcheck `json:"healthchecks"`
	}
	err := json.Unmarshal(data, &response)
	if err != nil {
		t.Fatalf("failed to unmarshal healthcheck list: %v", err)
	}

	if len(response.Healthchecks) != 1 {
		t.Errorf("expected 1 healthcheck, got %d", len(response.Healthchecks))
	}

	if len(response.Healthchecks) > 0 {
		if response.Healthchecks[0].UUID != "tok_hy4GCxOqbwahr4ctj0cNOIIp" {
			t.Errorf("expected first healthcheck UUID 'tok_hy4GCxOqbwahr4ctj0cNOIIp', got %q", response.Healthchecks[0].UUID)
		}
	}
}

// =============================================================================
// Outage Contract Tests
// =============================================================================

func TestContractOutage_UnmarshalResponse(t *testing.T) {
	data := loadTestData(t, "outages/response.json")

	var outage Outage
	err := json.Unmarshal(data, &outage)
	if err != nil {
		t.Fatalf("failed to unmarshal outage response: %v", err)
	}

	if outage.UUID != "outage_nqZWRijx1VQEQg" {
		t.Errorf("expected UUID 'outage_nqZWRijx1VQEQg', got %q", outage.UUID)
	}
	if outage.StatusCode != 500 {
		t.Errorf("expected StatusCode 500, got %d", outage.StatusCode)
	}
	if outage.Description != "Internal Server Error" {
		t.Errorf("expected Description 'Internal Server Error', got %q", outage.Description)
	}
	if outage.Monitor.UUID == "" {
		t.Error("expected Monitor.UUID to be set")
	} else if outage.Monitor.UUID != "mon_MdnOEmCL1uSAoV" {
		t.Errorf("expected Monitor.UUID 'mon_MdnOEmCL1uSAoV', got %q", outage.Monitor.UUID)
	}
}

func TestContractOutage_MarshalCreateRequest(t *testing.T) {
	endDate := "2026-02-09T05:00:00.000Z"
	req := CreateOutageRequest{
		MonitorUUID: "mon_abc123",
		StartDate:   "2026-02-09T04:00:00.000Z",
		EndDate:     &endDate,
		StatusCode:  500,
		Description: "Manual outage",
		OutageType:  "manual",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal outage create request: %v", err)
	}

	// Must use correct field names
	assertJSONContains(t, data, "monitorUuid")
	assertJSONContains(t, data, "startDate")
	assertJSONContains(t, data, "endDate")
	assertJSONContains(t, data, "statusCode")
	assertJSONContains(t, data, "description")
	assertJSONContains(t, data, "outageType")
}

func TestContractOutage_UnmarshalListResponse(t *testing.T) {
	data := loadTestData(t, "outages/list_response.json")

	var outages []Outage
	err := json.Unmarshal(data, &outages)
	if err != nil {
		t.Fatalf("failed to unmarshal outage list: %v", err)
	}

	if len(outages) != 1 {
		t.Errorf("expected 1 outage, got %d", len(outages))
	}

	if len(outages) > 0 {
		if outages[0].UUID != "outage_nqZWRijx1VQEQg" {
			t.Errorf("expected first outage UUID 'outage_nqZWRijx1VQEQg', got %q", outages[0].UUID)
		}
	}
}

// =============================================================================
// StatusPage Contract Tests
// =============================================================================

func TestContractStatusPage_UnmarshalResponse(t *testing.T) {
	data := loadTestData(t, "statuspages/response.json")

	var statusPage StatusPage
	err := json.Unmarshal(data, &statusPage)
	if err != nil {
		t.Fatalf("failed to unmarshal status page response: %v", err)
	}

	if statusPage.UUID != "sp_ide5MNPlpQBQgV" {
		t.Errorf("expected UUID 'sp_ide5MNPlpQBQgV', got %q", statusPage.UUID)
	}
	if statusPage.Name != "VCR Test Status Page" {
		t.Errorf("expected Name 'VCR Test Status Page', got %q", statusPage.Name)
	}
	if statusPage.HostedSubdomain != "vcr-test-20260205095502.hyperping.app" {
		t.Errorf("expected HostedSubdomain to be set, got %q", statusPage.HostedSubdomain)
	}
	if statusPage.Settings.Theme != "system" {
		t.Errorf("expected Settings.Theme 'system', got %q", statusPage.Settings.Theme)
	}
	if len(statusPage.Settings.Languages) != 1 {
		t.Errorf("expected 1 language, got %d", len(statusPage.Settings.Languages))
	}
}

func TestContractStatusPage_MarshalCreateRequest(t *testing.T) {
	subdomain := "test-statuspage"
	req := CreateStatusPageRequest{
		Name:      "Test Status Page",
		Subdomain: &subdomain,
		Languages: []string{"en"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("failed to marshal status page create request: %v", err)
	}

	// Must use correct field names
	assertJSONContains(t, data, "name")
	assertJSONContains(t, data, "subdomain")
	assertJSONContains(t, data, "languages")
}

func TestContractStatusPage_UnmarshalListResponse(t *testing.T) {
	data := loadTestData(t, "statuspages/list_response.json")

	var response StatusPagePaginatedResponse
	err := json.Unmarshal(data, &response)
	if err != nil {
		t.Fatalf("failed to unmarshal status page list: %v", err)
	}

	if len(response.StatusPages) != 1 {
		t.Errorf("expected 1 status page, got %d", len(response.StatusPages))
	}

	if len(response.StatusPages) > 0 {
		if response.StatusPages[0].UUID != "sp_ide5MNPlpQBQgV" {
			t.Errorf("expected first status page UUID 'sp_ide5MNPlpQBQgV', got %q", response.StatusPages[0].UUID)
		}
	}

	if response.Total != 1 {
		t.Errorf("expected Total 1, got %d", response.Total)
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
