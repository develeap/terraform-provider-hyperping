// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Contract Test Framework Examples
// =============================================================================
//
// This file demonstrates how to use the contract testing framework.
// These tests serve as executable examples and can be copied as templates.
//
// To run these examples:
//   RECORD_MODE=true HYPERPING_API_KEY=your_key go test -v -run TestContractExample

// =============================================================================
// Example 1: Basic List Operation
// =============================================================================

func TestContractExample_ListMonitors(t *testing.T) {
	t.Skip("Example test - enable when you have cassettes recorded")

	RunContractTest(t, ContractTestConfig{
		CassetteName: "example_monitor_list",
		ResourceType: "Monitor",
	}, func(t *testing.T, client *Client, ctx context.Context) {
		// Call the API
		monitors, err := client.ListMonitors(ctx)
		require.NoError(t, err)

		// Validate the contract
		validator := NewContractValidator(t, "Monitor")
		validator.ValidateMonitorList(monitors)

		// Additional assertions
		t.Logf("Found %d monitors", len(monitors))
		if len(monitors) > 0 {
			assert.NotEmpty(t, monitors[0].UUID)
			assert.NotEmpty(t, monitors[0].Name)
		}
	})
}

// =============================================================================
// Example 2: CRUD Operations
// =============================================================================

func TestContractExample_MonitorCRUD(t *testing.T) {
	t.Skip("Example test - enable when you have cassettes recorded")

	RunContractTest(t, ContractTestConfig{
		CassetteName: "example_monitor_crud",
		ResourceType: "Monitor",
	}, func(t *testing.T, client *Client, ctx context.Context) {
		validator := NewContractValidator(t, "Monitor")

		// CREATE
		createReq := BuildTestMonitor("Example Monitor")
		monitor, err := client.CreateMonitor(ctx, createReq)
		require.NoError(t, err, "Create should succeed")
		validator.ValidateMonitor(monitor)
		t.Logf("Created monitor: %s (%s)", monitor.Name, monitor.UUID)

		// Ensure cleanup
		testCtx := MonitorTestContext{
			Monitor: monitor,
			Client:  client,
			Context: ctx,
		}
		defer CleanupMonitor(t, testCtx)

		// READ
		retrieved, err := client.GetMonitor(ctx, monitor.UUID)
		require.NoError(t, err, "Read should succeed")
		validator.ValidateMonitor(retrieved)
		assert.Equal(t, monitor.UUID, retrieved.UUID)
		t.Logf("Retrieved monitor: %s", retrieved.Name)

		// UPDATE
		newName := "Updated Example Monitor"
		updateReq := UpdateMonitorRequest{
			Name: &newName,
		}
		updated, err := client.UpdateMonitor(ctx, monitor.UUID, updateReq)
		require.NoError(t, err, "Update should succeed")
		validator.ValidateMonitor(updated)
		assert.Equal(t, newName, updated.Name)
		t.Logf("Updated monitor name to: %s", updated.Name)

		// LIST (verify it appears)
		monitors, err := client.ListMonitors(ctx)
		require.NoError(t, err, "List should succeed")
		validator.ValidateMonitorList(monitors)

		found := false
		for _, m := range monitors {
			if m.UUID == monitor.UUID {
				found = true
				break
			}
		}
		assert.True(t, found, "Created monitor should appear in list")

		// DELETE is handled by defer CleanupMonitor
	})
}

// =============================================================================
// Example 3: Error Handling
// =============================================================================

func TestContractExample_ErrorHandling(t *testing.T) {
	t.Skip("Example test - enable when you have cassettes recorded")

	RunContractTest(t, ContractTestConfig{
		CassetteName: "example_error_handling",
		ResourceType: "Monitor",
	}, func(t *testing.T, client *Client, ctx context.Context) {
		// Test: Not Found Error
		_, err := client.GetMonitor(ctx, "mon_nonexistent123456")
		ExpectNotFoundError(t, err)

		// Test: Validation Error (empty UUID)
		_, err = client.GetMonitor(ctx, "")
		ExpectValidationError(t, err, "UUID")

		// Test: Validation Error (invalid request)
		invalidReq := CreateMonitorRequest{
			Name:     "", // Empty name should fail validation
			URL:      "invalid-url",
			Protocol: "invalid",
		}
		_, err = client.CreateMonitor(ctx, invalidReq)
		require.Error(t, err, "Invalid request should fail")
		t.Logf("Got expected validation error: %v", err)
	})
}

// =============================================================================
// Example 4: Pagination
// =============================================================================

func TestContractExample_Pagination(t *testing.T) {
	t.Skip("Example test - enable when you have cassettes recorded")

	RunContractTest(t, ContractTestConfig{
		CassetteName: "example_pagination",
		ResourceType: "StatusPage",
	}, func(t *testing.T, client *Client, ctx context.Context) {
		validator := NewContractValidator(t, "StatusPage")

		// Get first page
		page := 0
		response, err := client.ListStatusPages(ctx, &page, nil)
		require.NoError(t, err)
		validator.ValidateStatusPageList(response)

		t.Logf("Page %d: %d results (total: %d)", response.Page, len(response.StatusPages), response.Total)
		assert.Equal(t, 0, response.Page)
		assert.NotNil(t, response.StatusPages)

		// If there are more pages, fetch next page
		if response.HasNextPage {
			nextPage := 1
			nextResponse, err := client.ListStatusPages(ctx, &nextPage, nil)
			require.NoError(t, err)
			validator.ValidateStatusPageList(nextResponse)

			t.Logf("Page %d: %d results", nextResponse.Page, len(nextResponse.StatusPages))
			assert.Equal(t, 1, nextResponse.Page)
		}
	})
}

// =============================================================================
// Example 5: Nested Resources
// =============================================================================

func TestContractExample_NestedResources(t *testing.T) {
	t.Skip("Example test - enable when you have cassettes recorded")

	RunContractTest(t, ContractTestConfig{
		CassetteName: "example_nested_resources",
		ResourceType: "Incident",
	}, func(t *testing.T, client *Client, ctx context.Context) {
		validator := NewContractValidator(t, "Incident")

		// Get a status page first
		pages, err := client.ListStatusPages(ctx, nil, nil)
		require.NoError(t, err)
		require.NotEmpty(t, pages.StatusPages, "Need at least one status page")
		statusPageUUID := pages.StatusPages[0].UUID

		// Create incident
		createReq := BuildTestIncident("Example Incident", []string{statusPageUUID})
		incident, err := client.CreateIncident(ctx, createReq)
		require.NoError(t, err)
		validator.ValidateIncident(incident)
		t.Logf("Created incident: %s", incident.UUID)

		testCtx := IncidentTestContext{
			Incident: incident,
			Client:   client,
			Context:  ctx,
		}
		defer CleanupIncident(t, testCtx)

		// Add update (nested resource)
		updateReq := AddIncidentUpdateRequest{
			Text: LocalizedText{En: "Update message"},
			Type: "identified",
			Date: "2026-02-10T12:00:00Z",
		}

		updated, err := client.AddIncidentUpdate(ctx, incident.UUID, updateReq)
		require.NoError(t, err)
		validator.ValidateIncident(updated)

		// Verify update was added
		require.NotEmpty(t, updated.Updates, "Should have at least one update")
		assert.Equal(t, "identified", updated.Updates[0].Type)
		t.Logf("Added incident update, total updates: %d", len(updated.Updates))
	})
}

// =============================================================================
// Example 6: Field-Level Validation
// =============================================================================

func TestContractExample_FieldValidation(t *testing.T) {
	// This example shows how to use individual validators directly

	// UUID validation
	ValidateUUID(t, "MonitorUUID", "mon_abc123xyz")

	// Timestamp validation
	ValidateTimestamp(t, "CreatedAt", "2026-02-10T12:00:00Z")

	// Enum validation
	ValidateEnum(t, "Protocol", "http", []string{"http", "https", "tcp"})

	// URL validation
	ValidateURL(t, "MonitorURL", "https://example.com/health")

	// String length validation
	ValidateStringLength(t, "Name", "Test Monitor", 255)

	// Integer range validation
	ValidateIntegerRange(t, "CheckFrequency", 60, 10, 86400)

	// Hex color validation
	ValidateHexColor(t, "AccentColor", "#ff5733")

	// Localized text validation
	text := LocalizedText{
		En: "English text",
		Fr: "Texte français",
	}
	ValidateLocalizedText(t, "Title", text, 255)

	t.Log("All field validations passed")
}

// =============================================================================
// Example 7: Using Test Data Builders
// =============================================================================

func TestContractExample_TestDataBuilders(t *testing.T) {
	// Monitors
	monitorReq := BuildTestMonitor("My Monitor")
	assert.Equal(t, "My Monitor", monitorReq.Name)
	assert.Equal(t, "https://httpstat.us/200", monitorReq.URL)
	assert.Equal(t, "http", monitorReq.Protocol)

	// Incidents
	incidentReq := BuildTestIncident("Outage", []string{"sp_abc123"})
	assert.Equal(t, "Outage", incidentReq.Title.En)
	assert.Equal(t, "incident", incidentReq.Type)

	// Maintenance
	maintenanceReq := BuildTestMaintenance("DB Maintenance", []string{"mon_abc123"})
	assert.Equal(t, "DB Maintenance", maintenanceReq.Name)
	assert.NotEmpty(t, maintenanceReq.StartDate)

	// Status Pages
	statusPageReq := BuildTestStatusPage("My Status Page", "mysubdomain")
	assert.Equal(t, "My Status Page", statusPageReq.Name)
	assert.Equal(t, "mysubdomain", *statusPageReq.Subdomain)

	// Healthchecks
	healthcheckReq := BuildTestHealthcheck("Cron Monitor")
	assert.Equal(t, "Cron Monitor", healthcheckReq.Name)
	assert.NotNil(t, healthcheckReq.PeriodValue)

	t.Log("All test data builders work correctly")
}

// =============================================================================
// Example 8: Complete Workflow
// =============================================================================

func TestContractExample_CompleteWorkflow(t *testing.T) {
	t.Skip("Example test - enable when you have cassettes recorded")

	RunContractTest(t, ContractTestConfig{
		CassetteName: "example_complete_workflow",
		ResourceType: "Monitor",
	}, func(t *testing.T, client *Client, ctx context.Context) {
		validator := NewContractValidator(t, "Monitor")

		// Step 1: Create a monitor
		t.Log("Step 1: Creating monitor...")
		createReq := BuildTestMonitor("Production API")
		monitor, err := client.CreateMonitor(ctx, createReq)
		require.NoError(t, err)
		validator.ValidateMonitor(monitor)
		defer CleanupMonitor(t, MonitorTestContext{monitor, client, ctx})

		// Step 2: Verify it appears in list
		t.Log("Step 2: Verifying monitor in list...")
		monitors, err := client.ListMonitors(ctx)
		require.NoError(t, err)
		validator.ValidateMonitorList(monitors)

		foundInList := false
		for _, m := range monitors {
			if m.UUID == monitor.UUID {
				foundInList = true
				break
			}
		}
		assert.True(t, foundInList, "Monitor should appear in list")

		// Step 3: Update the monitor
		t.Log("Step 3: Updating monitor...")
		newName := "Production API (Updated)"
		paused := true
		updateReq := UpdateMonitorRequest{
			Name:   &newName,
			Paused: &paused,
		}
		updated, err := client.UpdateMonitor(ctx, monitor.UUID, updateReq)
		require.NoError(t, err)
		validator.ValidateMonitor(updated)
		assert.Equal(t, newName, updated.Name)
		assert.True(t, updated.Paused)

		// Step 4: Read it back
		t.Log("Step 4: Reading updated monitor...")
		retrieved, err := client.GetMonitor(ctx, monitor.UUID)
		require.NoError(t, err)
		validator.ValidateMonitor(retrieved)
		assert.Equal(t, newName, retrieved.Name)
		assert.True(t, retrieved.Paused)

		t.Log("Complete workflow validated successfully!")
	})
}
