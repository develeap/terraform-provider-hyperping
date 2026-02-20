// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"os"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// =============================================================================
// Contract Test Helpers
// =============================================================================
//
// This file provides helper functions for running contract tests with VCR
// cassettes. Contract tests verify that API responses match our expected
// structure and protect against breaking changes.
//
// Usage:
//   RunContractTest(t, ContractTestConfig{
//       CassetteName: "monitor_list",
//       ResourceType: "Monitor",
//   }, func(t *testing.T, client *Client, ctx context.Context) {
//       monitors, err := client.ListMonitors(ctx)
//       require.NoError(t, err)
//
//       validator := NewContractValidator(t, "Monitor")
//       validator.ValidateMonitorList(monitors)
//   })

// =============================================================================
// Test Configuration
// =============================================================================

// ContractTestConfig holds configuration for contract tests.
type ContractTestConfig struct {
	// CassetteName is the name of the VCR cassette file (without extension).
	CassetteName string

	// ResourceType is the name of the resource being tested (for logging).
	ResourceType string

	// CassetteDir is the directory where cassettes are stored.
	// Defaults to "testdata/cassettes" if not specified.
	CassetteDir string

	// SkipRecording skips the test if no cassette exists and not in record mode.
	// Set to false to fail the test instead of skipping.
	SkipRecording bool
}

// =============================================================================
// Contract Test Runner
// =============================================================================

// RunContractTest is a helper to run contract tests with VCR cassettes.
// It handles:
// - VCR recorder setup and teardown
// - API client creation with custom HTTP client
// - Context creation
// - Error handling for recorder stop
//
// The test function receives a fully configured client and context.
func RunContractTest(t *testing.T, config ContractTestConfig, testFunc func(*testing.T, *Client, context.Context)) {
	t.Helper()

	// Ensure environment is set for recording
	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	// Set default cassette directory
	if config.CassetteDir == "" {
		config.CassetteDir = "testdata/cassettes"
	}

	// Create VCR recorder
	mode := testutil.GetRecordMode()
	if config.SkipRecording && mode == testutil.ModeAuto {
		mode = testutil.ModeAuto
	}

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: config.CassetteName,
		Mode:         mode,
		CassetteDir:  config.CassetteDir,
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop VCR recorder: %v", err)
		}
	}()

	// Create API client with VCR HTTP client
	apiKey := getAPIKeyForTest(t)
	client := NewClient(apiKey, WithHTTPClient(httpClient))
	ctx := context.Background()

	// Run the test function
	testFunc(t, client, ctx)
}

// RunContractTestWithCleanup is like RunContractTest but also runs a cleanup function.
// The cleanup function is called even if the test fails.
func RunContractTestWithCleanup(
	t *testing.T,
	config ContractTestConfig,
	testFunc func(*testing.T, *Client, context.Context),
	cleanupFunc func(*testing.T, *Client, context.Context),
) {
	t.Helper()

	testutil.RequireEnvForRecording(t, "HYPERPING_API_KEY")

	if config.CassetteDir == "" {
		config.CassetteDir = "testdata/cassettes"
	}

	mode := testutil.GetRecordMode()
	if config.SkipRecording && mode == testutil.ModeAuto {
		mode = testutil.ModeAuto
	}

	r, httpClient := testutil.NewVCRRecorder(t, testutil.VCRConfig{
		CassetteName: config.CassetteName,
		Mode:         mode,
		CassetteDir:  config.CassetteDir,
	})
	defer func() {
		if err := r.Stop(); err != nil {
			t.Errorf("failed to stop VCR recorder: %v", err)
		}
	}()

	apiKey := getAPIKeyForTest(t)
	client := NewClient(apiKey, WithHTTPClient(httpClient))
	ctx := context.Background()

	// Ensure cleanup runs even if test fails
	defer cleanupFunc(t, client, ctx)

	testFunc(t, client, ctx)
}

// =============================================================================
// Multi-Resource Test Helpers
// =============================================================================

// MonitorTestContext holds resources created during monitor contract tests.
type MonitorTestContext struct {
	Monitor *Monitor
	Client  *Client
	Context context.Context
}

// CleanupMonitor is a helper to clean up monitor resources.
func CleanupMonitor(t *testing.T, ctx MonitorTestContext) {
	t.Helper()
	if ctx.Monitor != nil && ctx.Client != nil {
		if err := ctx.Client.DeleteMonitor(ctx.Context, ctx.Monitor.UUID); err != nil {
			t.Logf("Warning: failed to cleanup monitor %s: %v", ctx.Monitor.UUID, err)
		}
	}
}

// IncidentTestContext holds resources created during incident contract tests.
type IncidentTestContext struct {
	Incident *Incident
	Client   *Client
	Context  context.Context
}

// CleanupIncident is a helper to clean up incident resources.
func CleanupIncident(t *testing.T, ctx IncidentTestContext) {
	t.Helper()
	if ctx.Incident != nil && ctx.Client != nil {
		if err := ctx.Client.DeleteIncident(ctx.Context, ctx.Incident.UUID); err != nil {
			t.Logf("Warning: failed to cleanup incident %s: %v", ctx.Incident.UUID, err)
		}
	}
}

// MaintenanceTestContext holds resources created during maintenance contract tests.
type MaintenanceTestContext struct {
	Maintenance *Maintenance
	Client      *Client
	Context     context.Context
}

// CleanupMaintenance is a helper to clean up maintenance resources.
func CleanupMaintenance(t *testing.T, ctx MaintenanceTestContext) {
	t.Helper()
	if ctx.Maintenance != nil && ctx.Client != nil {
		if err := ctx.Client.DeleteMaintenance(ctx.Context, ctx.Maintenance.UUID); err != nil {
			t.Logf("Warning: failed to cleanup maintenance %s: %v", ctx.Maintenance.UUID, err)
		}
	}
}

// StatusPageTestContext holds resources created during status page contract tests.
type StatusPageTestContext struct {
	StatusPage *StatusPage
	Client     *Client
	Context    context.Context
}

// CleanupStatusPage is a helper to clean up status page resources.
func CleanupStatusPage(t *testing.T, ctx StatusPageTestContext) {
	t.Helper()
	if ctx.StatusPage != nil && ctx.Client != nil {
		if err := ctx.Client.DeleteStatusPage(ctx.Context, ctx.StatusPage.UUID); err != nil {
			t.Logf("Warning: failed to cleanup status page %s: %v", ctx.StatusPage.UUID, err)
		}
	}
}

// HealthcheckTestContext holds resources created during healthcheck contract tests.
type HealthcheckTestContext struct {
	Healthcheck *Healthcheck
	Client      *Client
	Context     context.Context
}

// CleanupHealthcheck is a helper to clean up healthcheck resources.
func CleanupHealthcheck(t *testing.T, ctx HealthcheckTestContext) {
	t.Helper()
	if ctx.Healthcheck != nil && ctx.Client != nil {
		if err := ctx.Client.DeleteHealthcheck(ctx.Context, ctx.Healthcheck.UUID); err != nil {
			t.Logf("Warning: failed to cleanup healthcheck %s: %v", ctx.Healthcheck.UUID, err)
		}
	}
}

// =============================================================================
// Validation Helpers
// =============================================================================

// ValidateResourceCRUD validates a complete CRUD cycle for a resource.
// This is a high-level helper that validates:
// - Create response structure
// - Read response structure
// - Update response structure
// - List response includes the resource
type ValidateResourceCRUD struct {
	t            *testing.T
	resourceType string
}

// NewValidateResourceCRUD creates a new CRUD validator.
func NewValidateResourceCRUD(t *testing.T, resourceType string) *ValidateResourceCRUD {
	t.Helper()
	return &ValidateResourceCRUD{
		t:            t,
		resourceType: resourceType,
	}
}

// ValidateCreate validates that a create operation returns a valid resource.
func (v *ValidateResourceCRUD) ValidateCreate(resource interface{}) {
	v.t.Helper()
	validator := NewContractValidator(v.t, v.resourceType)

	switch r := resource.(type) {
	case *Monitor:
		validator.ValidateMonitor(r)
	case *Incident:
		validator.ValidateIncident(r)
	case *Maintenance:
		validator.ValidateMaintenance(r)
	case *StatusPage:
		validator.ValidateStatusPage(r)
	case *Healthcheck:
		validator.ValidateHealthcheck(r)
	default:
		v.t.Fatalf("ValidateCreate: unhandled resource type %T", resource)
	}
}

// ValidateRead validates that a read operation returns a valid resource.
func (v *ValidateResourceCRUD) ValidateRead(resource interface{}) {
	v.t.Helper()
	v.ValidateCreate(resource) // Same validation as create
}

// ValidateUpdate validates that an update operation returns a valid resource.
func (v *ValidateResourceCRUD) ValidateUpdate(resource interface{}) {
	v.t.Helper()
	v.ValidateCreate(resource) // Same validation as create
}

// ValidateList validates that a list operation returns valid resources.
func (v *ValidateResourceCRUD) ValidateList(resources interface{}) {
	v.t.Helper()
	validator := NewContractValidator(v.t, v.resourceType)

	switch r := resources.(type) {
	case []Monitor:
		validator.ValidateMonitorList(r)
	case []Incident:
		validator.ValidateIncidentList(r)
	case []Maintenance:
		validator.ValidateMaintenanceList(r)
	case *StatusPagePaginatedResponse:
		validator.ValidateStatusPageList(r)
	case []Healthcheck:
		validator.ValidateHealthcheckList(r)
	case []Outage:
		validator.ValidateOutageList(r)
	default:
		v.t.Fatalf("ValidateList: unhandled resource type %T", resources)
	}
}

// =============================================================================
// Error Validation Helpers
// =============================================================================

// ExpectValidationError asserts that an error is a validation error.
func ExpectValidationError(t *testing.T, err error, fieldName string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected validation error for %s, got nil", fieldName)
	}
	t.Logf("Got expected validation error for %s: %v", fieldName, err)
}

// ExpectAPIError asserts that an error is an API error with expected status.
func ExpectAPIError(t *testing.T, err error, expectedStatus int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected API error with status %d, got nil", expectedStatus)
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("expected *APIError with status %d, got %T: %v", expectedStatus, err, err)
		return
	}

	if apiErr.StatusCode != expectedStatus {
		t.Errorf("expected status code %d, got %d", expectedStatus, apiErr.StatusCode)
	}
	t.Logf("Got expected API error (status %d): %v", expectedStatus, err)
}

// ExpectNotFoundError asserts that an error is a 404 Not Found error.
func ExpectNotFoundError(t *testing.T, err error) {
	t.Helper()
	ExpectAPIError(t, err, 404)
}

// ExpectUnauthorizedError asserts that an error is a 401 Unauthorized error.
func ExpectUnauthorizedError(t *testing.T, err error) {
	t.Helper()
	ExpectAPIError(t, err, 401)
}

// =============================================================================
// Test Data Builders
// =============================================================================

// BuildTestMonitor creates a monitor for testing with sensible defaults.
func BuildTestMonitor(name string) CreateMonitorRequest {
	return CreateMonitorRequest{
		Name:           name,
		URL:            "https://httpstat.us/200",
		Protocol:       "http",
		HTTPMethod:     "GET",
		CheckFrequency: 60,
		Regions:        []string{"london", "frankfurt"},
	}
}

// BuildTestIncident creates an incident for testing with sensible defaults.
func BuildTestIncident(title string, statusPages []string) CreateIncidentRequest {
	return CreateIncidentRequest{
		Title: LocalizedText{
			En: title,
		},
		Text: LocalizedText{
			En: "Test incident description",
		},
		Type:        "incident",
		StatusPages: statusPages,
	}
}

// BuildTestMaintenance creates a maintenance for testing with sensible defaults.
func BuildTestMaintenance(name string, monitors []string) CreateMaintenanceRequest {
	return CreateMaintenanceRequest{
		Name:      name,
		StartDate: "2026-12-01T00:00:00Z",
		EndDate:   "2026-12-01T02:00:00Z",
		Monitors:  monitors,
	}
}

// BuildTestStatusPage creates a status page for testing with sensible defaults.
func BuildTestStatusPage(name, subdomain string) CreateStatusPageRequest {
	return CreateStatusPageRequest{
		Name:      name,
		Subdomain: &subdomain,
		Languages: []string{"en"},
	}
}

// BuildTestHealthcheck creates a healthcheck for testing with sensible defaults.
func BuildTestHealthcheck(name string) CreateHealthcheckRequest {
	periodValue := 60
	periodType := "minutes"
	return CreateHealthcheckRequest{
		Name:             name,
		PeriodValue:      &periodValue,
		PeriodType:       &periodType,
		GracePeriodValue: 5,
		GracePeriodType:  "minutes",
	}
}

// =============================================================================
// Utilities
// =============================================================================

// getAPIKeyForTest returns the API key for testing.
// Uses HYPERPING_API_KEY environment variable if set, otherwise returns a test key.
func getAPIKeyForTest(t *testing.T) string {
	t.Helper()

	// In record mode, require real API key
	if testutil.GetRecordMode() == testutil.ModeRecord {
		key := os.Getenv("HYPERPING_API_KEY")
		if key == "" {
			t.Fatal("HYPERPING_API_KEY required for recording mode")
		}
		return key
	}

	// In replay mode, use test key (will be masked in cassettes)
	key := os.Getenv("HYPERPING_API_KEY")
	if key == "" {
		key = "test_api_key"
	}
	return key
}
