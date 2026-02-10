// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Contract Validation Framework
// =============================================================================
//
// This file provides reusable contract validation functions for testing
// API responses. Contract tests verify that the API structure matches our
// expectations and protects against breaking changes.
//
// Usage:
//   validator := NewContractValidator(t, "Monitor")
//   validator.ValidateMonitor(monitor)

// =============================================================================
// Basic Field Validators
// =============================================================================

// ValidateUUID checks if a string is a valid UUID format.
// Hyperping UUIDs follow the format: prefix_base62id (e.g., mon_abc123xyz)
func ValidateUUID(t *testing.T, fieldName, value string) {
	t.Helper()
	require.NotEmpty(t, value, "%s should not be empty", fieldName)

	// Check format: prefix_base62id (e.g., mon_abc123, sp_xyz789, tok_abc123def456)
	// Prefix is lowercase letters, ID is alphanumeric
	uuidPattern := regexp.MustCompile(`^[a-z]+_[a-zA-Z0-9]+$`)
	assert.Regexp(t, uuidPattern, value,
		"%s should match UUID format 'prefix_id', got: %s", fieldName, value)
}

// ValidateTimestamp checks if a string is a valid ISO 8601 timestamp.
// Hyperping API returns timestamps in RFC3339 format (ISO 8601 with timezone).
func ValidateTimestamp(t *testing.T, fieldName, value string) {
	t.Helper()
	require.NotEmpty(t, value, "%s should not be empty", fieldName)

	_, err := time.Parse(time.RFC3339, value)
	assert.NoError(t, err, "%s should be valid RFC3339 timestamp, got: %s", fieldName, value)
}

// ValidateOptionalTimestamp checks if a timestamp is valid when present.
func ValidateOptionalTimestamp(t *testing.T, fieldName string, value *string) {
	t.Helper()
	if value != nil && *value != "" {
		ValidateTimestamp(t, fieldName, *value)
	}
}

// ValidateEnum checks if a value is in the allowed set.
func ValidateEnum(t *testing.T, fieldName, value string, allowedValues []string) {
	t.Helper()
	require.NotEmpty(t, value, "%s should not be empty", fieldName)
	assert.Contains(t, allowedValues, value,
		"%s should be one of %v, got: %s", fieldName, allowedValues, value)
}

// ValidateOptionalEnum checks if an enum value is valid when present.
func ValidateOptionalEnum(t *testing.T, fieldName string, value *string, allowedValues []string) {
	t.Helper()
	if value != nil && *value != "" {
		ValidateEnum(t, fieldName, *value, allowedValues)
	}
}

// ValidateStringField checks basic string field requirements.
func ValidateStringField(t *testing.T, fieldName, value string, required bool) {
	t.Helper()
	if required {
		require.NotEmpty(t, value, "%s is required", fieldName)
	}
}

// ValidateStringLength checks that a string does not exceed max length.
func ValidateStringLength(t *testing.T, fieldName, value string, maxLength int) {
	t.Helper()
	length := len(value)
	assert.LessOrEqual(t, length, maxLength,
		"%s length should be <= %d, got %d", fieldName, maxLength, length)
}

// ValidateURL checks if a string is a valid URL format.
func ValidateURL(t *testing.T, fieldName, value string) {
	t.Helper()
	require.NotEmpty(t, value, "%s should not be empty", fieldName)

	// Basic URL validation - starts with http:// or https://
	urlPattern := regexp.MustCompile(`^https?://`)
	assert.Regexp(t, urlPattern, value,
		"%s should be a valid URL, got: %s", fieldName, value)
}

// ValidateOptionalURL checks if a URL is valid when present.
func ValidateOptionalURL(t *testing.T, fieldName string, value *string) {
	t.Helper()
	if value != nil && *value != "" {
		ValidateURL(t, fieldName, *value)
	}
}

// ValidateArrayNotEmpty checks that an array is not nil and has elements.
func ValidateArrayNotEmpty(t *testing.T, fieldName string, value interface{}) {
	t.Helper()
	require.NotNil(t, value, "%s should not be nil", fieldName)

	// Use type assertion to check length
	switch v := value.(type) {
	case []string:
		assert.NotEmpty(t, v, "%s should not be empty", fieldName)
	case []RequestHeader:
		assert.NotEmpty(t, v, "%s should not be empty", fieldName)
	case []IncidentUpdate:
		assert.NotEmpty(t, v, "%s should not be empty", fieldName)
	case []MaintenanceUpdate:
		assert.NotEmpty(t, v, "%s should not be empty", fieldName)
	case []StatusPageSection:
		assert.NotEmpty(t, v, "%s should not be empty", fieldName)
	default:
		t.Logf("Warning: ValidateArrayNotEmpty - unhandled type for %s", fieldName)
	}
}

// ValidateArrayLength checks that an array has expected length.
func ValidateArrayLength(t *testing.T, fieldName string, value interface{}, minLength, maxLength int) {
	t.Helper()

	var length int
	switch v := value.(type) {
	case []string:
		length = len(v)
	case []RequestHeader:
		length = len(v)
	case []IncidentUpdate:
		length = len(v)
	case []MaintenanceUpdate:
		length = len(v)
	case []StatusPageSection:
		length = len(v)
	default:
		t.Fatalf("ValidateArrayLength: unhandled type for %s", fieldName)
		return
	}

	if minLength >= 0 {
		assert.GreaterOrEqual(t, length, minLength,
			"%s length should be >= %d, got %d", fieldName, minLength, length)
	}
	if maxLength >= 0 {
		assert.LessOrEqual(t, length, maxLength,
			"%s length should be <= %d, got %d", fieldName, maxLength, length)
	}
}

// ValidateIntegerRange checks that an integer is within a valid range.
func ValidateIntegerRange(t *testing.T, fieldName string, value, minVal, maxVal int) {
	t.Helper()
	assert.GreaterOrEqual(t, value, minVal,
		"%s should be >= %d, got %d", fieldName, minVal, value)
	assert.LessOrEqual(t, value, maxVal,
		"%s should be <= %d, got %d", fieldName, maxVal, value)
}

// ValidatePositiveInteger checks that an integer is positive.
func ValidatePositiveInteger(t *testing.T, fieldName string, value int) {
	t.Helper()
	assert.Greater(t, value, 0, "%s should be positive, got %d", fieldName, value)
}

// ValidateOptionalInteger checks that an optional integer is valid when present.
func ValidateOptionalInteger(t *testing.T, fieldName string, value *int, minVal, maxVal int) {
	t.Helper()
	if value != nil {
		ValidateIntegerRange(t, fieldName, *value, minVal, maxVal)
	}
}

// ValidateLocalizedText checks that localized text has valid content.
func ValidateLocalizedText(t *testing.T, fieldName string, text LocalizedText, maxLength int) {
	t.Helper()

	// At least one language should be present
	hasContent := text.En != "" || text.Fr != "" || text.De != "" || text.Es != ""
	assert.True(t, hasContent, "%s should have at least one language set", fieldName)

	// Validate length for each present language
	if text.En != "" {
		ValidateStringLength(t, fmt.Sprintf("%s.en", fieldName), text.En, maxLength)
	}
	if text.Fr != "" {
		ValidateStringLength(t, fmt.Sprintf("%s.fr", fieldName), text.Fr, maxLength)
	}
	if text.De != "" {
		ValidateStringLength(t, fmt.Sprintf("%s.de", fieldName), text.De, maxLength)
	}
	if text.Es != "" {
		ValidateStringLength(t, fmt.Sprintf("%s.es", fieldName), text.Es, maxLength)
	}
}

// ValidateOptionalLocalizedText checks localized text when present.
func ValidateOptionalLocalizedText(t *testing.T, fieldName string, text *LocalizedText, maxLength int) {
	t.Helper()
	if text != nil {
		hasContent := text.En != "" || text.Fr != "" || text.De != "" || text.Es != ""
		if hasContent {
			ValidateLocalizedText(t, fieldName, *text, maxLength)
		}
	}
}

// ValidateHexColor checks if a string is a valid hex color code.
func ValidateHexColor(t *testing.T, fieldName, value string) {
	t.Helper()
	require.NotEmpty(t, value, "%s should not be empty", fieldName)

	hexColorPattern := regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)
	assert.Regexp(t, hexColorPattern, value,
		"%s should be valid hex color (e.g., #ff5733), got: %s", fieldName, value)
}

// =============================================================================
// ContractValidator - Fluent Interface for Resource Validation
// =============================================================================

// ContractValidator provides a fluent interface for validating API responses.
type ContractValidator struct {
	t            *testing.T
	resourceName string
}

// NewContractValidator creates a new contract validator.
func NewContractValidator(t *testing.T, resourceName string) *ContractValidator {
	t.Helper()
	return &ContractValidator{
		t:            t,
		resourceName: resourceName,
	}
}

// =============================================================================
// Monitor Validators
// =============================================================================

// ValidateMonitor validates a Monitor response structure.
func (cv *ContractValidator) ValidateMonitor(monitor *Monitor) {
	cv.t.Helper()
	require.NotNil(cv.t, monitor, "Monitor should not be nil")

	// Required fields
	ValidateUUID(cv.t, "UUID", monitor.UUID)
	ValidateStringField(cv.t, "Name", monitor.Name, true)
	ValidateStringLength(cv.t, "Name", monitor.Name, maxNameLength)
	ValidateURL(cv.t, "URL", monitor.URL)
	ValidateStringLength(cv.t, "URL", monitor.URL, maxURLLength)
	ValidateEnum(cv.t, "Protocol", monitor.Protocol, AllowedProtocols)
	ValidateEnum(cv.t, "HTTPMethod", monitor.HTTPMethod, AllowedMethods)

	// Numeric fields
	ValidateIntegerRange(cv.t, "CheckFrequency", monitor.CheckFrequency, 10, 86400)

	// Arrays
	if len(monitor.Regions) > 0 {
		for i, region := range monitor.Regions {
			ValidateEnum(cv.t, fmt.Sprintf("Regions[%d]", i), region, AllowedRegions)
		}
	}

	// Optional fields
	ValidateOptionalInteger(cv.t, "Port", monitor.Port, 1, 65535)
	ValidateOptionalInteger(cv.t, "SSLExpiration", monitor.SSLExpiration, 0, 365)

	// Read-only fields
	if monitor.Status != "" {
		ValidateEnum(cv.t, "Status", monitor.Status, []string{"up", "down"})
	}
}

// ValidateMonitorList validates a list of Monitor responses.
func (cv *ContractValidator) ValidateMonitorList(monitors []Monitor) {
	cv.t.Helper()
	require.NotNil(cv.t, monitors, "Monitor list should not be nil")

	for i, monitor := range monitors {
		cv.t.Run(fmt.Sprintf("Monitor[%d]", i), func(t *testing.T) {
			validator := NewContractValidator(t, cv.resourceName)
			validator.ValidateMonitor(&monitor)
		})
	}
}

// =============================================================================
// Incident Validators
// =============================================================================

// ValidateIncident validates an Incident response structure.
func (cv *ContractValidator) ValidateIncident(incident *Incident) {
	cv.t.Helper()
	require.NotNil(cv.t, incident, "Incident should not be nil")

	// Required fields
	ValidateUUID(cv.t, "UUID", incident.UUID)
	ValidateLocalizedText(cv.t, "Title", incident.Title, maxNameLength)
	ValidateLocalizedText(cv.t, "Text", incident.Text, maxMessageLength)
	ValidateEnum(cv.t, "Type", incident.Type, AllowedIncidentTypes)

	// Arrays
	require.NotNil(cv.t, incident.StatusPages, "StatusPages should not be nil")

	// Validate updates if present
	if len(incident.Updates) > 0 {
		for i, update := range incident.Updates {
			cv.t.Run(fmt.Sprintf("Update[%d]", i), func(t *testing.T) {
				ValidateUUID(t, "Update.UUID", update.UUID)
				ValidateTimestamp(t, "Update.Date", update.Date)
				ValidateLocalizedText(t, "Update.Text", update.Text, maxMessageLength)
				ValidateEnum(t, "Update.Type", update.Type, AllowedIncidentUpdateTypes)
			})
		}
	}

	// Optional timestamp
	if incident.Date != "" {
		ValidateTimestamp(cv.t, "Date", incident.Date)
	}
}

// ValidateIncidentList validates a list of Incident responses.
func (cv *ContractValidator) ValidateIncidentList(incidents []Incident) {
	cv.t.Helper()
	require.NotNil(cv.t, incidents, "Incident list should not be nil")

	for i, incident := range incidents {
		cv.t.Run(fmt.Sprintf("Incident[%d]", i), func(t *testing.T) {
			validator := NewContractValidator(t, cv.resourceName)
			validator.ValidateIncident(&incident)
		})
	}
}

// =============================================================================
// Maintenance Validators
// =============================================================================

// ValidateMaintenance validates a Maintenance response structure.
func (cv *ContractValidator) ValidateMaintenance(maintenance *Maintenance) {
	cv.t.Helper()
	require.NotNil(cv.t, maintenance, "Maintenance should not be nil")

	// Required fields
	ValidateUUID(cv.t, "UUID", maintenance.UUID)
	ValidateStringField(cv.t, "Name", maintenance.Name, true)
	ValidateStringLength(cv.t, "Name", maintenance.Name, maxNameLength)

	// Optional localized text
	ValidateOptionalLocalizedText(cv.t, "Title", &maintenance.Title, maxNameLength)
	ValidateOptionalLocalizedText(cv.t, "Text", &maintenance.Text, maxMessageLength)

	// Timestamps
	ValidateOptionalTimestamp(cv.t, "StartDate", maintenance.StartDate)
	ValidateOptionalTimestamp(cv.t, "EndDate", maintenance.EndDate)
	if maintenance.CreatedAt != "" {
		ValidateTimestamp(cv.t, "CreatedAt", maintenance.CreatedAt)
	}

	// Arrays
	require.NotNil(cv.t, maintenance.Monitors, "Monitors should not be nil")

	// Optional fields
	if maintenance.NotificationOption != "" {
		ValidateEnum(cv.t, "NotificationOption", maintenance.NotificationOption,
			AllowedNotificationOptions)
	}
	ValidateOptionalInteger(cv.t, "NotificationMinutes", maintenance.NotificationMinutes, 0, 10080)

	// Read-only fields
	if maintenance.Status != "" {
		ValidateEnum(cv.t, "Status", maintenance.Status,
			[]string{"upcoming", "ongoing", "completed"})
	}

	// Validate updates if present
	if len(maintenance.Updates) > 0 {
		for i, update := range maintenance.Updates {
			cv.t.Run(fmt.Sprintf("Update[%d]", i), func(t *testing.T) {
				ValidateTimestamp(t, "Update.Date", update.Date)
				ValidateLocalizedText(t, "Update.Text", update.Text, maxMessageLength)
			})
		}
	}
}

// ValidateMaintenanceList validates a list of Maintenance responses.
func (cv *ContractValidator) ValidateMaintenanceList(maintenances []Maintenance) {
	cv.t.Helper()
	require.NotNil(cv.t, maintenances, "Maintenance list should not be nil")

	for i, maintenance := range maintenances {
		cv.t.Run(fmt.Sprintf("Maintenance[%d]", i), func(t *testing.T) {
			validator := NewContractValidator(t, cv.resourceName)
			validator.ValidateMaintenance(&maintenance)
		})
	}
}

// =============================================================================
// Status Page Validators
// =============================================================================

// ValidateStatusPage validates a StatusPage response structure.
func (cv *ContractValidator) ValidateStatusPage(statusPage *StatusPage) {
	cv.t.Helper()
	require.NotNil(cv.t, statusPage, "StatusPage should not be nil")

	// Required fields
	ValidateUUID(cv.t, "UUID", statusPage.UUID)
	ValidateStringField(cv.t, "Name", statusPage.Name, true)
	ValidateStringField(cv.t, "HostedSubdomain", statusPage.HostedSubdomain, true)
	ValidateURL(cv.t, "URL", statusPage.URL)

	// Optional hostname
	ValidateOptionalURL(cv.t, "Hostname", statusPage.Hostname)

	// Validate settings
	cv.ValidateStatusPageSettings(&statusPage.Settings)

	// Validate sections
	if len(statusPage.Sections) > 0 {
		for i, section := range statusPage.Sections {
			cv.t.Run(fmt.Sprintf("Section[%d]", i), func(t *testing.T) {
				require.NotNil(t, section.Name, "Section.Name should not be nil")
				assert.NotEmpty(t, section.Name, "Section.Name should have at least one language")
			})
		}
	}
}

// ValidateStatusPageSettings validates StatusPageSettings structure.
func (cv *ContractValidator) ValidateStatusPageSettings(settings *StatusPageSettings) {
	cv.t.Helper()
	require.NotNil(cv.t, settings, "Settings should not be nil")

	// Required fields
	ValidateStringField(cv.t, "Settings.Name", settings.Name, true)

	// Arrays
	ValidateArrayNotEmpty(cv.t, "Settings.Languages", settings.Languages)
	for i, lang := range settings.Languages {
		ValidateEnum(cv.t, fmt.Sprintf("Settings.Languages[%d]", i), lang, AllowedLanguages)
	}
	ValidateEnum(cv.t, "Settings.DefaultLanguage", settings.DefaultLanguage, AllowedLanguages)

	// Enums
	ValidateEnum(cv.t, "Settings.Theme", settings.Theme, AllowedStatusPageThemes)
	ValidateEnum(cv.t, "Settings.Font", settings.Font, AllowedStatusPageFonts)

	// Hex color
	ValidateHexColor(cv.t, "Settings.AccentColor", settings.AccentColor)

	// Optional URL
	if settings.Website != "" {
		ValidateURL(cv.t, "Settings.Website", settings.Website)
	}
}

// ValidateStatusPageList validates a paginated StatusPage response.
func (cv *ContractValidator) ValidateStatusPageList(response *StatusPagePaginatedResponse) {
	cv.t.Helper()
	require.NotNil(cv.t, response, "StatusPagePaginatedResponse should not be nil")

	// Validate pagination fields
	ValidateIntegerRange(cv.t, "Page", response.Page, 0, 10000)
	ValidatePositiveInteger(cv.t, "Total", response.Total)
	ValidatePositiveInteger(cv.t, "ResultsPerPage", response.ResultsPerPage)

	// Validate each status page
	for i, statusPage := range response.StatusPages {
		cv.t.Run(fmt.Sprintf("StatusPage[%d]", i), func(t *testing.T) {
			validator := NewContractValidator(t, cv.resourceName)
			validator.ValidateStatusPage(&statusPage)
		})
	}
}

// =============================================================================
// Subscriber Validators
// =============================================================================

// ValidateSubscriber validates a StatusPageSubscriber response structure.
func (cv *ContractValidator) ValidateSubscriber(subscriber *StatusPageSubscriber) {
	cv.t.Helper()
	require.NotNil(cv.t, subscriber, "Subscriber should not be nil")

	// Required fields
	ValidatePositiveInteger(cv.t, "ID", subscriber.ID)
	ValidateEnum(cv.t, "Type", subscriber.Type, AllowedSubscriberTypes)
	ValidateStringField(cv.t, "Value", subscriber.Value, true)
	ValidateEnum(cv.t, "Language", subscriber.Language, AllowedLanguages)
	ValidateTimestamp(cv.t, "CreatedAt", subscriber.CreatedAt)

	// Validate type-specific fields
	switch subscriber.Type {
	case "email":
		require.NotNil(cv.t, subscriber.Email, "Email should not be nil for email subscriber")
		ValidateStringField(cv.t, "Email", *subscriber.Email, true)
	case "sms":
		require.NotNil(cv.t, subscriber.Phone, "Phone should not be nil for SMS subscriber")
		ValidateStringField(cv.t, "Phone", *subscriber.Phone, true)
	}
}

// ValidateSubscriberList validates a paginated Subscriber response.
func (cv *ContractValidator) ValidateSubscriberList(response *SubscriberPaginatedResponse) {
	cv.t.Helper()
	require.NotNil(cv.t, response, "SubscriberPaginatedResponse should not be nil")

	// Validate pagination fields
	ValidateIntegerRange(cv.t, "Page", response.Page, 0, 10000)
	ValidateIntegerRange(cv.t, "Total", response.Total, 0, 1000000)
	ValidatePositiveInteger(cv.t, "ResultsPerPage", response.ResultsPerPage)

	// Validate each subscriber
	for i, subscriber := range response.Subscribers {
		cv.t.Run(fmt.Sprintf("Subscriber[%d]", i), func(t *testing.T) {
			validator := NewContractValidator(t, cv.resourceName)
			validator.ValidateSubscriber(&subscriber)
		})
	}
}

// =============================================================================
// Healthcheck Validators
// =============================================================================

// ValidateHealthcheck validates a Healthcheck response structure.
func (cv *ContractValidator) ValidateHealthcheck(healthcheck *Healthcheck) {
	cv.t.Helper()
	require.NotNil(cv.t, healthcheck, "Healthcheck should not be nil")

	// Required fields
	ValidateUUID(cv.t, "UUID", healthcheck.UUID)
	ValidateStringField(cv.t, "Name", healthcheck.Name, true)
	ValidateStringLength(cv.t, "Name", healthcheck.Name, maxNameLength)
	ValidateURL(cv.t, "PingURL", healthcheck.PingURL)

	// Period validation
	ValidatePositiveInteger(cv.t, "Period", healthcheck.Period)
	ValidatePositiveInteger(cv.t, "GracePeriod", healthcheck.GracePeriod)

	if healthcheck.PeriodType != "" {
		ValidateEnum(cv.t, "PeriodType", healthcheck.PeriodType, AllowedPeriodTypes)
	}
	ValidateEnum(cv.t, "GracePeriodType", healthcheck.GracePeriodType, AllowedPeriodTypes)

	// Optional fields
	if healthcheck.CreatedAt != "" {
		ValidateTimestamp(cv.t, "CreatedAt", healthcheck.CreatedAt)
	}
	if healthcheck.LastPing != "" {
		ValidateTimestamp(cv.t, "LastPing", healthcheck.LastPing)
	}
	if healthcheck.DueDate != "" {
		ValidateTimestamp(cv.t, "DueDate", healthcheck.DueDate)
	}
}

// ValidateHealthcheckList validates a list of Healthcheck responses.
func (cv *ContractValidator) ValidateHealthcheckList(healthchecks []Healthcheck) {
	cv.t.Helper()
	require.NotNil(cv.t, healthchecks, "Healthcheck list should not be nil")

	for i, healthcheck := range healthchecks {
		cv.t.Run(fmt.Sprintf("Healthcheck[%d]", i), func(t *testing.T) {
			validator := NewContractValidator(t, cv.resourceName)
			validator.ValidateHealthcheck(&healthcheck)
		})
	}
}

// =============================================================================
// Outage Validators
// =============================================================================

// ValidateOutage validates an Outage response structure.
func (cv *ContractValidator) ValidateOutage(outage *Outage) {
	cv.t.Helper()
	require.NotNil(cv.t, outage, "Outage should not be nil")

	// Required fields
	ValidateUUID(cv.t, "UUID", outage.UUID)
	ValidateTimestamp(cv.t, "StartDate", outage.StartDate)
	ValidateStringField(cv.t, "Description", outage.Description, true)
	ValidateStringLength(cv.t, "Description", outage.Description, maxMessageLength)

	// Enum fields
	ValidateEnum(cv.t, "OutageType", outage.OutageType, []string{"manual", "automatic"})

	// Optional fields
	ValidateOptionalTimestamp(cv.t, "EndDate", outage.EndDate)
	ValidateOptionalTimestamp(cv.t, "AcknowledgedAt", outage.AcknowledgedAt)

	// Status code validation
	ValidateIntegerRange(cv.t, "StatusCode", outage.StatusCode, 100, 599)

	// Validate monitor reference
	require.NotNil(cv.t, outage.Monitor, "Monitor reference should not be nil")
	ValidateUUID(cv.t, "Monitor.UUID", outage.Monitor.UUID)
	ValidateStringField(cv.t, "Monitor.Name", outage.Monitor.Name, true)
}

// ValidateOutageList validates a list of Outage responses.
func (cv *ContractValidator) ValidateOutageList(outages []Outage) {
	cv.t.Helper()
	require.NotNil(cv.t, outages, "Outage list should not be nil")

	for i, outage := range outages {
		cv.t.Run(fmt.Sprintf("Outage[%d]", i), func(t *testing.T) {
			validator := NewContractValidator(t, cv.resourceName)
			validator.ValidateOutage(&outage)
		})
	}
}

// =============================================================================
// Report Validators
// =============================================================================

// ValidateMonitorReport validates a MonitorReport response structure.
func (cv *ContractValidator) ValidateMonitorReport(report *MonitorReport) {
	cv.t.Helper()
	require.NotNil(cv.t, report, "MonitorReport should not be nil")

	// Required fields
	ValidateUUID(cv.t, "UUID", report.UUID)
	ValidateStringField(cv.t, "Name", report.Name, true)
	ValidateEnum(cv.t, "Protocol", report.Protocol, AllowedProtocols)

	// SLA validation (0-100%)
	assert.GreaterOrEqual(cv.t, report.SLA, 0.0, "SLA should be >= 0")
	assert.LessOrEqual(cv.t, report.SLA, 100.0, "SLA should be <= 100")

	// Validate period
	require.NotNil(cv.t, &report.Period, "Period should not be nil")
	ValidateTimestamp(cv.t, "Period.From", report.Period.From)
	ValidateTimestamp(cv.t, "Period.To", report.Period.To)

	// Validate outage stats
	require.NotNil(cv.t, &report.Outages, "Outages should not be nil")
	ValidateIntegerRange(cv.t, "Outages.Count", report.Outages.Count, 0, 1000000)
}

// ValidateMonitorReportList validates a list report response.
func (cv *ContractValidator) ValidateMonitorReportList(response *ListMonitorReportsResponse) {
	cv.t.Helper()
	require.NotNil(cv.t, response, "ListMonitorReportsResponse should not be nil")

	// Validate period
	ValidateTimestamp(cv.t, "Period.From", response.Period.From)
	ValidateTimestamp(cv.t, "Period.To", response.Period.To)

	// Validate each monitor report
	for i, report := range response.Monitors {
		cv.t.Run(fmt.Sprintf("MonitorReport[%d]", i), func(t *testing.T) {
			validator := NewContractValidator(t, cv.resourceName)
			validator.ValidateMonitorReport(&report)
		})
	}
}
