// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// mockClient implements APIClient for testing.
type mockClient struct {
	monitors     []client.Monitor
	healthchecks []client.Healthcheck
	statusPages  []client.StatusPage
	incidents    []client.Incident
	maintenance  []client.Maintenance
	outages      []client.Outage

	monitorsErr     error
	healthchecksErr error
	statusPagesErr  error
	incidentsErr    error
	maintenanceErr  error
	outagesErr      error
}

func (m *mockClient) ListMonitors(_ context.Context) ([]client.Monitor, error) {
	return m.monitors, m.monitorsErr
}

func (m *mockClient) ListHealthchecks(_ context.Context) ([]client.Healthcheck, error) {
	return m.healthchecks, m.healthchecksErr
}

func (m *mockClient) ListStatusPages(_ context.Context, _ *int, _ *string) (*client.StatusPagePaginatedResponse, error) {
	if m.statusPagesErr != nil {
		return nil, m.statusPagesErr
	}
	return &client.StatusPagePaginatedResponse{StatusPages: m.statusPages}, nil
}

func (m *mockClient) ListIncidents(_ context.Context) ([]client.Incident, error) {
	return m.incidents, m.incidentsErr
}

func (m *mockClient) ListMaintenance(_ context.Context) ([]client.Maintenance, error) {
	return m.maintenance, m.maintenanceErr
}

func (m *mockClient) ListOutages(_ context.Context) ([]client.Outage, error) {
	return m.outages, m.outagesErr
}

// =============================================================================
// terraformName Tests
// =============================================================================

func TestTerraformName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		prefix   string
		expected string
	}{
		{
			name:     "simple name",
			input:    "api",
			expected: "api",
		},
		{
			name:     "name with spaces",
			input:    "My API Monitor",
			expected: "my_api_monitor",
		},
		{
			name:     "name with special chars",
			input:    "[PROD]-API-Health",
			expected: "prod_api_health",
		},
		{
			name:     "name starting with number",
			input:    "123-test",
			expected: "r_123_test",
		},
		{
			name:     "with prefix",
			input:    "api",
			prefix:   "prod_",
			expected: "prod_api",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "resource",
		},
		{
			name:     "unicode characters",
			input:    "API-健康检查",
			expected: "api",
		},
		{
			name:     "only special chars",
			input:    "!!!@@@###",
			expected: "resource",
		},
		{
			name:     "multiple consecutive special chars",
			input:    "api---health___check",
			expected: "api_health_check",
		},
		{
			name:     "leading and trailing special chars",
			input:    "---api---",
			expected: "api",
		},
		{
			name:     "prefix with number starting name",
			input:    "123test",
			prefix:   "prod_",
			expected: "prod_r_123test",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			g := &Generator{prefix: tc.prefix}
			result := g.terraformName(tc.input)
			if result != tc.expected {
				t.Errorf("terraformName(%q) = %q, want %q", tc.input, result, tc.expected)
			}
		})
	}
}

// =============================================================================
// escapeHCL Tests
// =============================================================================

func TestEscapeHCL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with \"quotes\"", "with \\\"quotes\\\""},
		{"with\nnewline", "with\\nnewline"},
		{"with\\backslash", "with\\\\backslash"},
		{"mixed \"quote\" and\nnewline", "mixed \\\"quote\\\" and\\nnewline"},
		{"", ""},
	}

	for _, tc := range tests {
		result := escapeHCL(tc.input)
		if result != tc.expected {
			t.Errorf("escapeHCL(%q) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// =============================================================================
// formatStringList Tests
// =============================================================================

func TestFormatStringList(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{nil, "[]"},
		{[]string{}, "[]"},
		{[]string{"a"}, "[\"a\"]"},
		{[]string{"a", "b"}, "[\"a\", \"b\"]"},
		{[]string{"with \"quote\""}, "[\"with \\\"quote\\\"\"]"},
		{[]string{"a", "b", "c"}, "[\"a\", \"b\", \"c\"]"},
	}

	for _, tc := range tests {
		result := formatStringList(tc.input)
		if result != tc.expected {
			t.Errorf("formatStringList(%v) = %q, want %q", tc.input, result, tc.expected)
		}
	}
}

// =============================================================================
// Generate Tests
// =============================================================================

func TestGenerate_ImportFormat(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Test Monitor"},
		},
	}

	g := &Generator{
		client:    mock,
		resources: []string{"monitors"},
	}

	result, err := g.Generate(context.Background(), "import")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(result, "terraform import hyperping_monitor.test_monitor") {
		t.Error("Expected import command in output")
	}
	if strings.Contains(result, "resource \"hyperping_monitor\"") {
		t.Error("HCL should not appear in import format")
	}
}

func TestGenerate_HCLFormat(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Test Monitor", URL: "https://example.com", Protocol: "http"},
		},
	}

	g := &Generator{
		client:    mock,
		resources: []string{"monitors"},
	}

	result, err := g.Generate(context.Background(), "hcl")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if strings.Contains(result, "terraform import") {
		t.Error("Import commands should not appear in hcl format")
	}
	if !strings.Contains(result, "resource \"hyperping_monitor\"") {
		t.Error("Expected HCL resource in output")
	}
}

func TestGenerate_BothFormat(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Test Monitor", URL: "https://example.com", Protocol: "http"},
		},
	}

	g := &Generator{
		client:    mock,
		resources: []string{"monitors"},
	}

	result, err := g.Generate(context.Background(), "both")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if !strings.Contains(result, "terraform import") {
		t.Error("Expected import commands in output")
	}
	if !strings.Contains(result, "resource \"hyperping_monitor\"") {
		t.Error("Expected HCL resource in output")
	}
	if !strings.Contains(result, "Terraform Import Commands") {
		t.Error("Expected section header for imports")
	}
	if !strings.Contains(result, "Terraform HCL Configuration") {
		t.Error("Expected section header for HCL")
	}
}

func TestGenerate_UnknownFormat(t *testing.T) {
	g := &Generator{
		client:    &mockClient{},
		resources: []string{},
	}

	_, err := g.Generate(context.Background(), "invalid")
	if err == nil {
		t.Fatal("Expected error for unknown format")
	}
	if !strings.Contains(err.Error(), "unknown format") {
		t.Errorf("Expected 'unknown format' error, got: %v", err)
	}
}

func TestGenerate_EmptyResources(t *testing.T) {
	g := &Generator{
		client:    &mockClient{},
		resources: []string{},
	}

	result, err := g.Generate(context.Background(), "import")
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result != "" {
		t.Errorf("Expected empty output for no resources, got: %q", result)
	}
}

// =============================================================================
// fetchResources Tests
// =============================================================================

func TestFetchResources_AllTypes(t *testing.T) {
	mock := &mockClient{
		monitors:     []client.Monitor{{UUID: "mon_1", Name: "Monitor"}},
		healthchecks: []client.Healthcheck{{UUID: "hc_1", Name: "Healthcheck"}},
		statusPages:  []client.StatusPage{{UUID: "sp_1", Name: "Status Page"}},
		incidents:    []client.Incident{{UUID: "inc_1", Title: client.LocalizedText{En: "Incident"}}},
		maintenance:  []client.Maintenance{{UUID: "maint_1", Name: "Maintenance"}},
		outages:      []client.Outage{{UUID: "out_1", Monitor: client.MonitorReference{Name: "Monitor"}}},
	}

	g := &Generator{
		client:    mock,
		resources: []string{"monitors", "healthchecks", "statuspages", "incidents", "maintenance", "outages"},
	}

	data, err := g.fetchResources(context.Background())
	if err != nil {
		t.Fatalf("fetchResources() error = %v", err)
	}

	if len(data.Monitors) != 1 {
		t.Errorf("Expected 1 monitor, got %d", len(data.Monitors))
	}
	if len(data.Healthchecks) != 1 {
		t.Errorf("Expected 1 healthcheck, got %d", len(data.Healthchecks))
	}
	if len(data.StatusPages) != 1 {
		t.Errorf("Expected 1 status page, got %d", len(data.StatusPages))
	}
	if len(data.Incidents) != 1 {
		t.Errorf("Expected 1 incident, got %d", len(data.Incidents))
	}
	if len(data.Maintenance) != 1 {
		t.Errorf("Expected 1 maintenance, got %d", len(data.Maintenance))
	}
	if len(data.Outages) != 1 {
		t.Errorf("Expected 1 outage, got %d", len(data.Outages))
	}
}

func TestFetchResources_MonitorsError(t *testing.T) {
	mock := &mockClient{
		monitorsErr: errors.New("API error"),
	}

	g := &Generator{
		client:    mock,
		resources: []string{"monitors"},
	}

	_, err := g.fetchResources(context.Background())
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "fetching monitors") {
		t.Errorf("Expected 'fetching monitors' in error, got: %v", err)
	}
}

func TestFetchResources_HealthchecksError(t *testing.T) {
	mock := &mockClient{
		healthchecksErr: errors.New("API error"),
	}

	g := &Generator{
		client:    mock,
		resources: []string{"healthchecks"},
	}

	_, err := g.fetchResources(context.Background())
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "fetching healthchecks") {
		t.Errorf("Expected 'fetching healthchecks' in error, got: %v", err)
	}
}

func TestFetchResources_StatusPagesError(t *testing.T) {
	mock := &mockClient{
		statusPagesErr: errors.New("API error"),
	}

	g := &Generator{
		client:    mock,
		resources: []string{"statuspages"},
	}

	_, err := g.fetchResources(context.Background())
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "fetching status pages") {
		t.Errorf("Expected 'fetching status pages' in error, got: %v", err)
	}
}

func TestFetchResources_IncidentsError(t *testing.T) {
	mock := &mockClient{
		incidentsErr: errors.New("API error"),
	}

	g := &Generator{
		client:    mock,
		resources: []string{"incidents"},
	}

	_, err := g.fetchResources(context.Background())
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "fetching incidents") {
		t.Errorf("Expected 'fetching incidents' in error, got: %v", err)
	}
}

func TestFetchResources_MaintenanceError(t *testing.T) {
	mock := &mockClient{
		maintenanceErr: errors.New("API error"),
	}

	g := &Generator{
		client:    mock,
		resources: []string{"maintenance"},
	}

	_, err := g.fetchResources(context.Background())
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "fetching maintenance") {
		t.Errorf("Expected 'fetching maintenance' in error, got: %v", err)
	}
}

func TestFetchResources_OutagesError(t *testing.T) {
	mock := &mockClient{
		outagesErr: errors.New("API error"),
	}

	g := &Generator{
		client:    mock,
		resources: []string{"outages"},
	}

	_, err := g.fetchResources(context.Background())
	if err == nil {
		t.Fatal("Expected error")
	}
	if !strings.Contains(err.Error(), "fetching outages") {
		t.Errorf("Expected 'fetching outages' in error, got: %v", err)
	}
}

// =============================================================================
// generateImports Tests
// =============================================================================

func TestGenerateImports_AllResourceTypes(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Test Monitor"},
		},
		Healthchecks: []client.Healthcheck{
			{UUID: "hc_456", Name: "Backup Job"},
		},
		StatusPages: []client.StatusPage{
			{UUID: "sp_789", Name: "Main Status"},
		},
		Incidents: []client.Incident{
			{UUID: "inc_abc", Title: client.LocalizedText{En: "API Outage"}},
		},
		Maintenance: []client.Maintenance{
			{UUID: "maint_def", Title: client.LocalizedText{En: "DB Maintenance"}},
		},
		Outages: []client.Outage{
			{UUID: "out_ghi", Monitor: client.MonitorReference{Name: "API Monitor"}},
		},
	}

	g.generateImports(&sb, data)
	result := sb.String()

	expected := []string{
		`terraform import hyperping_monitor.test_monitor "mon_123"`,
		`terraform import hyperping_healthcheck.backup_job "hc_456"`,
		`terraform import hyperping_statuspage.main_status "sp_789"`,
		`terraform import hyperping_incident.api_outage "inc_abc"`,
		`terraform import hyperping_maintenance.db_maintenance "maint_def"`,
		`terraform import hyperping_outage.api_monitor "out_ghi"`,
	}

	for _, exp := range expected {
		if !strings.Contains(result, exp) {
			t.Errorf("Missing import command: %s", exp)
		}
	}
}

func TestGenerateImports_MaintenanceFallbackToName(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	data := &ResourceData{
		Maintenance: []client.Maintenance{
			{UUID: "maint_1", Name: "Fallback Name", Title: client.LocalizedText{En: ""}},
		},
	}

	g.generateImports(&sb, data)
	result := sb.String()

	if !strings.Contains(result, "hyperping_maintenance.fallback_name") {
		t.Errorf("Expected fallback to Name field, got: %s", result)
	}
}

func TestGenerateImports_WithPrefix(t *testing.T) {
	g := &Generator{prefix: "prod_"}
	var sb strings.Builder

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_123", Name: "API"},
		},
	}

	g.generateImports(&sb, data)
	result := sb.String()

	if !strings.Contains(result, "hyperping_monitor.prod_api") {
		t.Errorf("Expected prefixed resource name, got: %s", result)
	}
}

// =============================================================================
// generateMonitorHCL Tests
// =============================================================================

func TestGenerateMonitorHCL_Basic(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	monitor := client.Monitor{
		UUID:           "mon_123",
		Name:           "Test Monitor",
		URL:            "https://example.com",
		Protocol:       "http",
		CheckFrequency: 60,
		Regions:        []string{"virginia"},
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	assertions := []string{
		`resource "hyperping_monitor" "test_monitor"`,
		`name     = "Test Monitor"`,
		`url      = "https://example.com"`,
		`protocol = "http"`,
		`regions = ["virginia"]`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}

	// Should NOT contain check_frequency since it's default 60
	if strings.Contains(result, "check_frequency") {
		t.Error("Should not include default check_frequency")
	}
}

func TestGenerateMonitorHCL_AllOptionalFields(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	port := 8080
	keyword := "healthy"
	policy := "esc_123"
	monitor := client.Monitor{
		UUID:             "mon_123",
		Name:             "Full Monitor",
		URL:              "https://example.com/api",
		Protocol:         "http",
		HTTPMethod:       "POST",
		CheckFrequency:   30,
		Regions:          []string{"virginia", "london"},
		Port:             &port,
		FollowRedirects:  false,
		RequiredKeyword:  &keyword,
		Paused:           true,
		AlertsWait:       5,
		EscalationPolicy: &policy,
		RequestHeaders:   []client.RequestHeader{{Name: "Auth", Value: "Bearer token"}},
		RequestBody:      `{"test": true}`,
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	assertions := []string{
		`http_method = "POST"`,
		`check_frequency = 30`,
		`regions = ["virginia", "london"]`,
		`port = 8080`,
		`follow_redirects = false`,
		`required_keyword = "healthy"`,
		`paused = true`,
		`alerts_wait = 5`,
		`escalation_policy_uuid = "esc_123"`,
		`request_headers = {`,
		`"Auth" = "Bearer token"`,
		`request_body = "{\"test\": true}"`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}
}

func TestGenerateMonitorHCL_NilPointers(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	monitor := client.Monitor{
		UUID:             "mon_123",
		Name:             "Minimal",
		URL:              "https://example.com",
		Protocol:         "http",
		CheckFrequency:   60,
		Port:             nil,
		RequiredKeyword:  nil,
		EscalationPolicy: nil,
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	// Should not contain optional fields when nil
	if strings.Contains(result, "port =") {
		t.Error("Should not include nil port")
	}
	if strings.Contains(result, "required_keyword") {
		t.Error("Should not include nil required_keyword")
	}
	if strings.Contains(result, "escalation_policy") {
		t.Error("Should not include nil escalation_policy")
	}
}

func TestGenerateMonitorHCL_ZeroPort(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	port := 0
	monitor := client.Monitor{
		UUID:     "mon_123",
		Name:     "Test",
		URL:      "https://example.com",
		Protocol: "http",
		Port:     &port,
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	// Should not include port = 0
	if strings.Contains(result, "port =") {
		t.Error("Should not include zero port")
	}
}

// =============================================================================
// generateHealthcheckHCL Tests
// =============================================================================

func TestGenerateHealthcheckHCL_WithCron(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	healthcheck := client.Healthcheck{
		UUID:        "hc_123",
		Name:        "Backup Job",
		Cron:        "0 0 * * *",
		Timezone:    "America/New_York",
		GracePeriod: 300,
		IsPaused:    true,
	}

	g.generateHealthcheckHCL(&sb, healthcheck)
	result := sb.String()

	assertions := []string{
		`resource "hyperping_healthcheck" "backup_job"`,
		`name = "Backup Job"`,
		`cron = "0 0 * * *"`,
		`timezone = "America/New_York"`,
		`grace_period = 300`,
		`paused = true`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}
}

func TestGenerateHealthcheckHCL_WithPeriod(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	periodValue := 5
	healthcheck := client.Healthcheck{
		UUID:        "hc_123",
		Name:        "Heartbeat",
		PeriodValue: &periodValue,
		PeriodType:  "minutes",
	}

	g.generateHealthcheckHCL(&sb, healthcheck)
	result := sb.String()

	assertions := []string{
		`period_value = 5`,
		`period_type = "minutes"`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}

	// Should not contain cron
	if strings.Contains(result, "cron =") {
		t.Error("Should not include cron when using period")
	}
}

func TestGenerateHealthcheckHCL_Minimal(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	healthcheck := client.Healthcheck{
		UUID: "hc_123",
		Name: "Simple",
	}

	g.generateHealthcheckHCL(&sb, healthcheck)
	result := sb.String()

	if !strings.Contains(result, `resource "hyperping_healthcheck" "simple"`) {
		t.Error("Missing resource declaration")
	}
	if !strings.Contains(result, `name = "Simple"`) {
		t.Error("Missing name")
	}

	// Should not contain optional fields
	if strings.Contains(result, "paused =") {
		t.Error("Should not include paused when false")
	}
	if strings.Contains(result, "grace_period") {
		t.Error("Should not include zero grace_period")
	}
}

// =============================================================================
// generateStatusPageHCL Tests
// =============================================================================

func TestGenerateStatusPageHCL_Basic(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	statusPage := client.StatusPage{
		UUID:            "sp_123",
		Name:            "Main Status",
		HostedSubdomain: "status",
		Settings: client.StatusPageSettings{
			Languages: []string{"en"},
		},
	}

	g.generateStatusPageHCL(&sb, statusPage)
	result := sb.String()

	assertions := []string{
		`resource "hyperping_statuspage" "main_status"`,
		`name             = "Main Status"`,
		`hosted_subdomain = "status"`,
		`settings = {`,
		`name      = "Main Status"`,
		`languages = ["en"]`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}
}

func TestGenerateStatusPageHCL_WithHostname(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	hostname := "status.example.com"
	statusPage := client.StatusPage{
		UUID:            "sp_123",
		Name:            "Custom Status",
		HostedSubdomain: "custom",
		Hostname:        &hostname,
		Settings: client.StatusPageSettings{
			Languages:   []string{"en", "fr"},
			Theme:       "dark",
			Font:        "Roboto",
			AccentColor: "#ff0000",
		},
	}

	g.generateStatusPageHCL(&sb, statusPage)
	result := sb.String()

	assertions := []string{
		`hostname = "status.example.com"`,
		`languages = ["en", "fr"]`,
		`theme = "dark"`,
		`font = "Roboto"`,
		`accent_color = "#ff0000"`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}
}

func TestGenerateStatusPageHCL_WithSections(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	statusPage := client.StatusPage{
		UUID:            "sp_123",
		Name:            "With Sections",
		HostedSubdomain: "status",
		Settings: client.StatusPageSettings{
			Languages: []string{"en"},
		},
		Sections: []client.StatusPageSection{
			{Name: map[string]string{"en": "API"}},
		},
	}

	g.generateStatusPageHCL(&sb, statusPage)
	result := sb.String()

	if !strings.Contains(result, "# Note: Sections imported") {
		t.Error("Missing sections note")
	}
}

func TestGenerateStatusPageHCL_DefaultSettings(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	statusPage := client.StatusPage{
		UUID:            "sp_123",
		Name:            "Default",
		HostedSubdomain: "default",
		Settings: client.StatusPageSettings{
			Theme:       "system",
			Font:        "Inter",
			AccentColor: "#36b27e",
		},
	}

	g.generateStatusPageHCL(&sb, statusPage)
	result := sb.String()

	// Should not include default values
	if strings.Contains(result, "theme =") {
		t.Error("Should not include default theme")
	}
	if strings.Contains(result, "font =") {
		t.Error("Should not include default font")
	}
	if strings.Contains(result, "accent_color =") {
		t.Error("Should not include default accent_color")
	}
}

// =============================================================================
// generateIncidentHCL Tests
// =============================================================================

func TestGenerateIncidentHCL_Basic(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	incident := client.Incident{
		UUID:  "inc_123",
		Title: client.LocalizedText{En: "API Outage"},
		Text:  client.LocalizedText{En: "Investigating the issue"},
	}

	g.generateIncidentHCL(&sb, incident)
	result := sb.String()

	assertions := []string{
		`resource "hyperping_incident" "api_outage"`,
		`title = "API Outage"`,
		`text = "Investigating the issue"`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}
}

func TestGenerateIncidentHCL_AllFields(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	incident := client.Incident{
		UUID:               "inc_123",
		Title:              client.LocalizedText{En: "Major Outage"},
		Text:               client.LocalizedText{En: "All systems down"},
		Type:               "maintenance",
		StatusPages:        []string{"sp_1", "sp_2"},
		AffectedComponents: []string{"api", "web"},
	}

	g.generateIncidentHCL(&sb, incident)
	result := sb.String()

	assertions := []string{
		`type = "maintenance"`,
		`status_pages = ["sp_1", "sp_2"]`,
		`affected_components = ["api", "web"]`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}
}

func TestGenerateIncidentHCL_DefaultType(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	incident := client.Incident{
		UUID:  "inc_123",
		Title: client.LocalizedText{En: "Test"},
		Type:  "incident",
	}

	g.generateIncidentHCL(&sb, incident)
	result := sb.String()

	// Should not include default type
	if strings.Contains(result, "type =") {
		t.Error("Should not include default type 'incident'")
	}
}

// =============================================================================
// generateMaintenanceHCL Tests
// =============================================================================

func TestGenerateMaintenanceHCL_Basic(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	startDate := "2026-01-20T02:00:00Z"
	endDate := "2026-01-20T04:00:00Z"
	maintenance := client.Maintenance{
		UUID:      "maint_123",
		Title:     client.LocalizedText{En: "DB Maintenance"},
		Text:      client.LocalizedText{En: "Routine maintenance"},
		StartDate: &startDate,
		EndDate:   &endDate,
	}

	g.generateMaintenanceHCL(&sb, maintenance)
	result := sb.String()

	assertions := []string{
		`resource "hyperping_maintenance" "db_maintenance"`,
		`title = "DB Maintenance"`,
		`text = "Routine maintenance"`,
		`start_date = "2026-01-20T02:00:00Z"`,
		`end_date   = "2026-01-20T04:00:00Z"`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}
}

func TestGenerateMaintenanceHCL_FallbackToName(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	maintenance := client.Maintenance{
		UUID:  "maint_123",
		Name:  "Fallback Name",
		Title: client.LocalizedText{En: ""},
	}

	g.generateMaintenanceHCL(&sb, maintenance)
	result := sb.String()

	if !strings.Contains(result, `title = "Fallback Name"`) {
		t.Errorf("Expected fallback to Name, got: %s", result)
	}
	if !strings.Contains(result, `"fallback_name"`) {
		t.Errorf("Expected resource name from fallback, got: %s", result)
	}
}

func TestGenerateMaintenanceHCL_WithStatusPages(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	maintenance := client.Maintenance{
		UUID:        "maint_123",
		Title:       client.LocalizedText{En: "Test"},
		StatusPages: []string{"sp_1", "sp_2"},
	}

	g.generateMaintenanceHCL(&sb, maintenance)
	result := sb.String()

	if !strings.Contains(result, `status_pages = ["sp_1", "sp_2"]`) {
		t.Errorf("Missing status_pages, got: %s", result)
	}
}

// =============================================================================
// generateOutageHCL Tests
// =============================================================================

func TestGenerateOutageHCL_Basic(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	outage := client.Outage{
		UUID: "out_123",
		Monitor: client.MonitorReference{
			UUID: "mon_456",
			Name: "API Monitor",
		},
	}

	g.generateOutageHCL(&sb, outage)
	result := sb.String()

	assertions := []string{
		`resource "hyperping_outage" "api_monitor"`,
		`monitor_uuid = "mon_456"`,
		`# Note: Outages are mostly read-only`,
	}

	for _, assertion := range assertions {
		if !strings.Contains(result, assertion) {
			t.Errorf("Missing: %s\nGot: %s", assertion, result)
		}
	}
}

func TestGenerateOutageHCL_WithDescription(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	outage := client.Outage{
		UUID:        "out_123",
		Description: "Connection timeout",
		Monitor: client.MonitorReference{
			UUID: "mon_456",
			Name: "Test",
		},
	}

	g.generateOutageHCL(&sb, outage)
	result := sb.String()

	if !strings.Contains(result, `# description = "Connection timeout"`) {
		t.Errorf("Missing commented description, got: %s", result)
	}
}

// =============================================================================
// generateHCL Tests
// =============================================================================

func TestGenerateHCL_MultipleResources(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_1", Name: "Monitor 1", URL: "https://example.com", Protocol: "http"},
			{UUID: "mon_2", Name: "Monitor 2", URL: "https://example.org", Protocol: "http"},
		},
		Healthchecks: []client.Healthcheck{
			{UUID: "hc_1", Name: "HC 1"},
		},
	}

	g.generateHCL(&sb, data)
	result := sb.String()

	if strings.Count(result, "resource \"hyperping_monitor\"") != 2 {
		t.Error("Expected 2 monitor resources")
	}
	if strings.Count(result, "resource \"hyperping_healthcheck\"") != 1 {
		t.Error("Expected 1 healthcheck resource")
	}
}

func TestGenerateHCL_EmptyData(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	data := &ResourceData{}

	g.generateHCL(&sb, data)
	result := sb.String()

	if result != "" {
		t.Errorf("Expected empty output for empty data, got: %s", result)
	}
}

func TestGenerateHCL_AllResourceTypes(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_1", Name: "Monitor", URL: "https://example.com", Protocol: "http"},
		},
		Healthchecks: []client.Healthcheck{
			{UUID: "hc_1", Name: "HC"},
		},
		StatusPages: []client.StatusPage{
			{UUID: "sp_1", Name: "Status", HostedSubdomain: "status", Settings: client.StatusPageSettings{Languages: []string{"en"}}},
		},
		Incidents: []client.Incident{
			{UUID: "inc_1", Title: client.LocalizedText{En: "Incident"}},
		},
		Maintenance: []client.Maintenance{
			{UUID: "maint_1", Title: client.LocalizedText{En: "Maintenance"}},
		},
		Outages: []client.Outage{
			{UUID: "out_1", Monitor: client.MonitorReference{UUID: "mon_1", Name: "Monitor"}},
		},
	}

	g.generateHCL(&sb, data)
	result := sb.String()

	// Verify all resource types are generated
	expectedResources := []string{
		`resource "hyperping_monitor"`,
		`resource "hyperping_healthcheck"`,
		`resource "hyperping_statuspage"`,
		`resource "hyperping_incident"`,
		`resource "hyperping_maintenance"`,
		`resource "hyperping_outage"`,
	}

	for _, exp := range expectedResources {
		if !strings.Contains(result, exp) {
			t.Errorf("Missing resource type: %s", exp)
		}
	}
}

// =============================================================================
// parseResources Tests
// =============================================================================

func TestParseResources_All(t *testing.T) {
	result := parseResources("all")

	expected := []string{"monitors", "healthchecks", "statuspages", "incidents", "maintenance", "outages"}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d resources, got %d", len(expected), len(result))
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Expected %q at index %d, got %q", exp, i, result[i])
		}
	}
}

func TestParseResources_Single(t *testing.T) {
	result := parseResources("monitors")

	if len(result) != 1 || result[0] != "monitors" {
		t.Errorf("Expected [monitors], got %v", result)
	}
}

func TestParseResources_Multiple(t *testing.T) {
	result := parseResources("monitors,healthchecks,incidents")

	expected := []string{"monitors", "healthchecks", "incidents"}
	if len(result) != len(expected) {
		t.Fatalf("Expected %d resources, got %d", len(expected), len(result))
	}

	for i, exp := range expected {
		if result[i] != exp {
			t.Errorf("Expected %q at index %d, got %q", exp, i, result[i])
		}
	}
}

// =============================================================================
// Monitor HCL Edge Cases
// =============================================================================

func TestGenerateMonitorHCL_ExpectedStatusCode(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	monitor := client.Monitor{
		UUID:               "mon_123",
		Name:               "Test",
		URL:                "https://example.com",
		Protocol:           "http",
		ExpectedStatusCode: client.FlexibleString("201"),
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	if !strings.Contains(result, `expected_status_code = "201"`) {
		t.Errorf("Expected expected_status_code = \"201\", got: %s", result)
	}
}

func TestGenerateMonitorHCL_EmptyKeyword(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	emptyKeyword := ""
	monitor := client.Monitor{
		UUID:            "mon_123",
		Name:            "Test",
		URL:             "https://example.com",
		Protocol:        "http",
		RequiredKeyword: &emptyKeyword,
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	// Should not include empty required_keyword
	if strings.Contains(result, "required_keyword") {
		t.Error("Should not include empty required_keyword")
	}
}

func TestGenerateMonitorHCL_EmptyEscalationPolicy(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	emptyPolicy := ""
	monitor := client.Monitor{
		UUID:             "mon_123",
		Name:             "Test",
		URL:              "https://example.com",
		Protocol:         "http",
		EscalationPolicy: &emptyPolicy,
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	// Should not include empty escalation_policy
	if strings.Contains(result, "escalation_policy") {
		t.Error("Should not include empty escalation_policy")
	}
}

func TestGenerateMonitorHCL_DefaultHTTPMethod(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	monitor := client.Monitor{
		UUID:       "mon_123",
		Name:       "Test",
		URL:        "https://example.com",
		Protocol:   "http",
		HTTPMethod: "GET",
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	// Should not include default GET method
	if strings.Contains(result, "http_method") {
		t.Error("Should not include default http_method GET")
	}
}

func TestGenerateMonitorHCL_EmptyHTTPMethod(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	monitor := client.Monitor{
		UUID:       "mon_123",
		Name:       "Test",
		URL:        "https://example.com",
		Protocol:   "http",
		HTTPMethod: "",
	}

	g.generateMonitorHCL(&sb, monitor)
	result := sb.String()

	// Should not include empty http_method
	if strings.Contains(result, "http_method") {
		t.Error("Should not include empty http_method")
	}
}

// =============================================================================
// StatusPage HCL Edge Cases
// =============================================================================

func TestGenerateStatusPageHCL_EmptyHostname(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	emptyHostname := ""
	statusPage := client.StatusPage{
		UUID:            "sp_123",
		Name:            "Test",
		HostedSubdomain: "test",
		Hostname:        &emptyHostname,
		Settings: client.StatusPageSettings{
			Languages: []string{"en"},
		},
	}

	g.generateStatusPageHCL(&sb, statusPage)
	result := sb.String()

	// Should not include empty hostname
	if strings.Contains(result, "hostname =") {
		t.Error("Should not include empty hostname")
	}
}

func TestGenerateStatusPageHCL_EmptyLanguages(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	statusPage := client.StatusPage{
		UUID:            "sp_123",
		Name:            "Test",
		HostedSubdomain: "test",
		Settings: client.StatusPageSettings{
			Languages: []string{},
		},
	}

	g.generateStatusPageHCL(&sb, statusPage)
	result := sb.String()

	// Should default to ["en"]
	if !strings.Contains(result, `languages = ["en"]`) {
		t.Errorf("Expected default languages = [\"en\"], got: %s", result)
	}
}

// =============================================================================
// Incident HCL Edge Cases
// =============================================================================

func TestGenerateIncidentHCL_EmptyText(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	incident := client.Incident{
		UUID:  "inc_123",
		Title: client.LocalizedText{En: "Test"},
		Text:  client.LocalizedText{En: ""},
	}

	g.generateIncidentHCL(&sb, incident)
	result := sb.String()

	// Should not include empty text
	if strings.Contains(result, "text =") {
		t.Error("Should not include empty text")
	}
}

func TestGenerateIncidentHCL_EmptyType(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	incident := client.Incident{
		UUID:  "inc_123",
		Title: client.LocalizedText{En: "Test"},
		Type:  "",
	}

	g.generateIncidentHCL(&sb, incident)
	result := sb.String()

	// Should not include empty type
	if strings.Contains(result, "type =") {
		t.Error("Should not include empty type")
	}
}

// =============================================================================
// Maintenance HCL Edge Cases
// =============================================================================

func TestGenerateMaintenanceHCL_EmptyText(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	maintenance := client.Maintenance{
		UUID:  "maint_123",
		Title: client.LocalizedText{En: "Test"},
		Text:  client.LocalizedText{En: ""},
	}

	g.generateMaintenanceHCL(&sb, maintenance)
	result := sb.String()

	// Should not include empty text
	if strings.Contains(result, "text =") {
		t.Error("Should not include empty text")
	}
}

func TestGenerateMaintenanceHCL_NilDates(t *testing.T) {
	g := &Generator{}
	var sb strings.Builder

	maintenance := client.Maintenance{
		UUID:      "maint_123",
		Title:     client.LocalizedText{En: "Test"},
		StartDate: nil,
		EndDate:   nil,
	}

	g.generateMaintenanceHCL(&sb, maintenance)
	result := sb.String()

	// Should not include nil dates
	if strings.Contains(result, "start_date") {
		t.Error("Should not include nil start_date")
	}
	if strings.Contains(result, "end_date") {
		t.Error("Should not include nil end_date")
	}
}
