// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Test BuildTroubleshootingSteps for each error type

func TestBuildTroubleshootingSteps_NotFound(t *testing.T) {
	ctx := ErrorContext{
		Type:         "not_found",
		HTTPStatus:   404,
		ResourceType: "Monitor",
		ResourceID:   "mon_test123",
		Operation:    "read",
	}

	steps := BuildTroubleshootingSteps(ctx)

	// Verify steps contain expected content
	if !strings.Contains(steps, "Troubleshooting:") {
		t.Error("Expected troubleshooting header")
	}
	if !strings.Contains(steps, "Verify the Monitor still exists") {
		t.Error("Expected resource existence check")
	}
	if !strings.Contains(steps, "https://app.hyperping.io/monitors") {
		t.Error("Expected dashboard link")
	}
	if !strings.Contains(steps, "mon_test123") {
		t.Error("Expected resource ID in steps")
	}
	if !strings.Contains(steps, "terraform state show") {
		t.Error("Expected terraform state show command")
	}
}

func TestBuildTroubleshootingSteps_AuthError(t *testing.T) {
	ctx := ErrorContext{
		Type:         "auth_error",
		HTTPStatus:   401,
		ResourceType: "Monitor",
		Operation:    "create",
	}

	steps := BuildTroubleshootingSteps(ctx)

	if !strings.Contains(steps, "HYPERPING_API_KEY") {
		t.Error("Expected API key environment variable mention")
	}
	if !strings.Contains(steps, "starts with 'sk_'") {
		t.Error("Expected API key format guidance")
	}
	if !strings.Contains(steps, "curl") {
		t.Error("Expected curl test command")
	}
	if !strings.Contains(steps, "https://app.hyperping.io/settings/api") {
		t.Error("Expected API key generation link")
	}
}

func TestBuildTroubleshootingSteps_AuthError403(t *testing.T) {
	ctx := ErrorContext{
		Type:         "auth_error",
		HTTPStatus:   403,
		ResourceType: "Incident",
		Operation:    "delete",
	}

	steps := BuildTroubleshootingSteps(ctx)

	if !strings.Contains(steps, "required permissions") {
		t.Error("Expected permissions check for 403")
	}
	if !strings.Contains(steps, "Incident resources") {
		t.Error("Expected resource-specific permission mention")
	}
}

func TestBuildTroubleshootingSteps_RateLimit(t *testing.T) {
	ctx := ErrorContext{
		Type:       "rate_limit",
		HTTPStatus: 429,
		RetryAfter: 120,
		Operation:  "update",
	}

	steps := BuildTroubleshootingSteps(ctx)

	if !strings.Contains(steps, "Wait 120 seconds") {
		t.Error("Expected retry time from context")
	}
	if !strings.Contains(steps, "terraform apply -parallelism=1") {
		t.Error("Expected parallelism reduction guidance")
	}
	if !strings.Contains(steps, "rate-limits") {
		t.Error("Expected rate limits documentation link")
	}
}

func TestBuildTroubleshootingSteps_RateLimitDefaultWait(t *testing.T) {
	ctx := ErrorContext{
		Type:       "rate_limit",
		HTTPStatus: 429,
		RetryAfter: 0, // No retry-after specified
		Operation:  "update",
	}

	steps := BuildTroubleshootingSteps(ctx)

	if !strings.Contains(steps, "Wait 60 seconds") {
		t.Error("Expected default 60 second wait time")
	}
}

func TestBuildTroubleshootingSteps_ServerError(t *testing.T) {
	ctx := ErrorContext{
		Type:       "server_error",
		HTTPStatus: 500,
		Operation:  "create",
	}

	steps := BuildTroubleshootingSteps(ctx)

	if !strings.Contains(steps, "https://status.hyperping.app") {
		t.Error("Expected status page link")
	}
	if !strings.Contains(steps, "retry") {
		t.Error("Expected retry guidance")
	}
	if !strings.Contains(steps, "support") {
		t.Error("Expected support contact information")
	}
}

func TestBuildTroubleshootingSteps_Validation(t *testing.T) {
	ctx := ErrorContext{
		Type:         "validation",
		HTTPStatus:   400,
		ResourceType: "Monitor",
		Operation:    "create",
	}

	steps := BuildTroubleshootingSteps(ctx)

	if !strings.Contains(steps, "required fields") {
		t.Error("Expected required fields mention")
	}
	if !strings.Contains(steps, "validation") {
		t.Error("Expected validation mention")
	}
	if !strings.Contains(steps, "Monitor documentation") {
		t.Error("Expected resource-specific documentation reference")
	}
}

func TestBuildTroubleshootingSteps_ValidationMonitor(t *testing.T) {
	ctx := ErrorContext{
		Type:         "validation",
		HTTPStatus:   400,
		ResourceType: "Monitor",
		Operation:    "create",
	}

	steps := BuildTroubleshootingSteps(ctx)

	// Check for Monitor-specific validation guidance
	if !strings.Contains(steps, "Frequency must be one of") {
		t.Error("Expected frequency validation guidance")
	}
	if !strings.Contains(steps, "Timeout must be one of") {
		t.Error("Expected timeout validation guidance")
	}
	if !strings.Contains(steps, "valid Hyperping region codes") {
		t.Error("Expected region validation guidance")
	}
}

func TestBuildTroubleshootingSteps_ValidationIncident(t *testing.T) {
	ctx := ErrorContext{
		Type:         "validation",
		HTTPStatus:   422,
		ResourceType: "Incident",
		Operation:    "create",
	}

	steps := BuildTroubleshootingSteps(ctx)

	// Check for Incident-specific validation guidance
	if !strings.Contains(steps, "Status must be one of") {
		t.Error("Expected status validation guidance")
	}
	if !strings.Contains(steps, "Severity must be one of") {
		t.Error("Expected severity validation guidance")
	}
}

func TestBuildTroubleshootingSteps_ValidationMaintenance(t *testing.T) {
	ctx := ErrorContext{
		Type:         "validation",
		HTTPStatus:   400,
		ResourceType: "Maintenance",
		Operation:    "create",
	}

	steps := BuildTroubleshootingSteps(ctx)

	// Check for Maintenance-specific validation guidance
	if !strings.Contains(steps, "ISO 8601") {
		t.Error("Expected timestamp format guidance")
	}
	if !strings.Contains(steps, "scheduledEnd must be after scheduledStart") {
		t.Error("Expected date ordering guidance")
	}
}

func TestBuildTroubleshootingSteps_Unknown(t *testing.T) {
	ctx := ErrorContext{
		Type:       "unknown",
		HTTPStatus: 0,
		Operation:  "read",
	}

	steps := BuildTroubleshootingSteps(ctx)

	if !strings.Contains(steps, "network connectivity") {
		t.Error("Expected network check guidance")
	}
	if !strings.Contains(steps, "service status") {
		t.Error("Expected service status check")
	}
	if !strings.Contains(steps, "provider maintainers") {
		t.Error("Expected maintainer contact information")
	}
}

// Test enhanced error creation functions

func TestNewReadErrorWithContext(t *testing.T) {
	err := client.NewAPIError(404, "monitor not found")
	diag := NewReadErrorWithContext("Monitor", "mon_abc123", err)

	summary := diag.Summary()
	detail := diag.Detail()

	if summary != "Failed to Read Monitor" {
		t.Errorf("Expected summary 'Failed to Read Monitor', got: %s", summary)
	}
	if !strings.Contains(detail, "mon_abc123") {
		t.Error("Expected resource ID in detail")
	}
	if !strings.Contains(detail, "Troubleshooting:") {
		t.Error("Expected troubleshooting section")
	}
	if !strings.Contains(detail, "dashboard") {
		t.Error("Expected dashboard reference")
	}
}

func TestNewCreateErrorWithContext(t *testing.T) {
	err := client.NewAPIError(401, "unauthorized")
	diag := NewCreateErrorWithContext("Incident", err)

	summary := diag.Summary()
	detail := diag.Detail()

	if summary != "Failed to Create Incident" {
		t.Errorf("Expected summary 'Failed to Create Incident', got: %s", summary)
	}
	if !strings.Contains(detail, "Troubleshooting:") {
		t.Error("Expected troubleshooting section")
	}
	if !strings.Contains(detail, "HYPERPING_API_KEY") {
		t.Error("Expected API key guidance for auth error")
	}
}

func TestNewUpdateErrorWithContext(t *testing.T) {
	err := client.NewRateLimitError(90)
	diag := NewUpdateErrorWithContext("Monitor", "mon_xyz789", err)

	summary := diag.Summary()
	detail := diag.Detail()

	if summary != "Failed to Update Monitor" {
		t.Errorf("Expected summary 'Failed to Update Monitor', got: %s", summary)
	}
	if !strings.Contains(detail, "mon_xyz789") {
		t.Error("Expected resource ID in detail")
	}
	if !strings.Contains(detail, "Wait 90 seconds") {
		t.Error("Expected retry-after guidance")
	}
	if !strings.Contains(detail, "parallelism") {
		t.Error("Expected parallelism reduction guidance")
	}
}

func TestNewDeleteErrorWithContext(t *testing.T) {
	err := client.NewAPIError(500, "internal server error")
	diag := NewDeleteErrorWithContext("Maintenance", "maint_test", err)

	summary := diag.Summary()
	detail := diag.Detail()

	if summary != "Failed to Delete Maintenance" {
		t.Errorf("Expected summary 'Failed to Delete Maintenance', got: %s", summary)
	}
	if !strings.Contains(detail, "maint_test") {
		t.Error("Expected resource ID in detail")
	}
	if !strings.Contains(detail, "status.hyperping.app") {
		t.Error("Expected status page link for server error")
	}
}

// Test dashboard URL generation

func TestGetDashboardURL_AllResourceTypes(t *testing.T) {
	tests := []struct {
		resourceType string
		expectedURL  string
	}{
		{"Monitor", "https://app.hyperping.io/monitors"},
		{"Incident", "https://app.hyperping.io/incidents"},
		{"Maintenance", "https://app.hyperping.io/maintenance"},
		{"Statuspage", "https://app.hyperping.io/statuspages"},
		{"Healthcheck", "https://app.hyperping.io/healthchecks"},
		{"Outage", "https://app.hyperping.io/outages"},
		{"Unknown", "https://app.hyperping.io"},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			url := getDashboardURL(tt.resourceType)
			if url != tt.expectedURL {
				t.Errorf("Expected URL %s, got: %s", tt.expectedURL, url)
			}
		})
	}
}

// Test specific error content requirements

func TestTroubleshootingSteps_ContainsDashboardLink(t *testing.T) {
	resourceTypes := []string{"Monitor", "Incident", "Maintenance", "Statuspage"}

	for _, resourceType := range resourceTypes {
		t.Run(resourceType, func(t *testing.T) {
			ctx := ErrorContext{
				Type:         "not_found",
				HTTPStatus:   404,
				ResourceType: resourceType,
				ResourceID:   "test_id",
				Operation:    "read",
			}

			steps := BuildTroubleshootingSteps(ctx)

			if !strings.Contains(steps, "app.hyperping.io") {
				t.Errorf("Expected dashboard link for %s", resourceType)
			}
		})
	}
}

func TestTroubleshootingSteps_ContainsRetryTime(t *testing.T) {
	retryTimes := []int{30, 60, 90, 120, 180}

	for _, retryTime := range retryTimes {
		t.Run(fmt.Sprintf("RetryAfter%d", retryTime), func(t *testing.T) {
			ctx := ErrorContext{
				Type:       "rate_limit",
				HTTPStatus: 429,
				RetryAfter: retryTime,
				Operation:  "create",
			}

			steps := BuildTroubleshootingSteps(ctx)

			expectedMsg := fmt.Sprintf("Wait %d seconds", retryTime)
			if !strings.Contains(steps, expectedMsg) {
				t.Errorf("Expected retry time %d in steps", retryTime)
			}
		})
	}
}

func TestTroubleshootingSteps_ResourceSpecific(t *testing.T) {
	tests := []struct {
		resourceType string
		errorType    string
		expectedText string
	}{
		{"Monitor", "not_found", "monitors"},
		{"Incident", "not_found", "incidents"},
		{"Maintenance", "not_found", "maintenance"},
		{"Monitor", "validation", "Frequency must be"},
		{"Incident", "validation", "Status must be"},
		{"Maintenance", "validation", "ISO 8601"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.resourceType, tt.errorType), func(t *testing.T) {
			ctx := ErrorContext{
				Type:         tt.errorType,
				HTTPStatus:   400,
				ResourceType: tt.resourceType,
				Operation:    "create",
			}

			steps := BuildTroubleshootingSteps(ctx)

			if !strings.Contains(strings.ToLower(steps), strings.ToLower(tt.expectedText)) {
				t.Errorf("Expected resource-specific text '%s' for %s %s error",
					tt.expectedText, tt.resourceType, tt.errorType)
			}
		})
	}
}

// Test error message formatting

func TestEnhancedErrorMessage_Format(t *testing.T) {
	err := errors.New("test error")
	diag := NewReadErrorWithContext("Monitor", "mon_test", err)

	detail := diag.Detail()

	// Check formatting structure
	if !strings.Contains(detail, "Unable to read") {
		t.Error("Expected 'Unable to read' in message")
	}
	if !strings.Contains(detail, "got error:") {
		t.Error("Expected 'got error:' separator")
	}
	if !strings.Contains(detail, "Troubleshooting:\n") {
		t.Error("Expected 'Troubleshooting:' section with newline")
	}
}

func TestEnhancedErrorMessage_AllOperations(t *testing.T) {
	testErr := errors.New("test error")

	tests := []struct {
		op       string
		summary  string
		hasResID bool
	}{
		{"Read", NewReadErrorWithContext("Monitor", "mon_test", testErr).Summary(), true},
		{"Create", NewCreateErrorWithContext("Monitor", testErr).Summary(), false},
		{"Update", NewUpdateErrorWithContext("Monitor", "mon_test", testErr).Summary(), true},
		{"Delete", NewDeleteErrorWithContext("Monitor", "mon_test", testErr).Summary(), true},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			expected := fmt.Sprintf("Failed to %s Monitor", tt.op)

			if tt.summary != expected {
				t.Errorf("Expected summary '%s', got: %s", expected, tt.summary)
			}
		})
	}
}

// Test helper functions

func TestBuildNotFoundSteps_Content(t *testing.T) {
	ctx := ErrorContext{
		Type:         "not_found",
		HTTPStatus:   404,
		ResourceType: "Monitor",
		ResourceID:   "mon_test",
		Operation:    "read",
	}

	steps := buildNotFoundSteps(ctx)

	if len(steps) < 4 {
		t.Errorf("Expected at least 4 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	if !strings.Contains(stepsText, "dashboard") {
		t.Error("Expected dashboard mention")
	}
	if !strings.Contains(stepsText, "deleted outside") {
		t.Error("Expected external deletion mention")
	}
}

func TestBuildAuthErrorSteps_Content(t *testing.T) {
	ctx := ErrorContext{
		Type:       "auth_error",
		HTTPStatus: 401,
		Operation:  "read",
	}

	steps := buildAuthErrorSteps(ctx)

	if len(steps) < 3 {
		t.Errorf("Expected at least 3 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	if !strings.Contains(stepsText, "API key") {
		t.Error("Expected API key mention")
	}
	if !strings.Contains(stepsText, "curl") {
		t.Error("Expected curl command")
	}
}

func TestBuildRateLimitSteps_Content(t *testing.T) {
	ctx := ErrorContext{
		Type:       "rate_limit",
		HTTPStatus: 429,
		RetryAfter: 45,
		Operation:  "create",
	}

	steps := buildRateLimitSteps(ctx)

	if len(steps) < 3 {
		t.Errorf("Expected at least 3 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	if !strings.Contains(stepsText, "45 seconds") {
		t.Error("Expected specific retry time")
	}
	if !strings.Contains(stepsText, "parallelism") {
		t.Error("Expected parallelism mention")
	}
}

func TestBuildServerErrorSteps_Content(t *testing.T) {
	ctx := ErrorContext{
		Type:       "server_error",
		HTTPStatus: 500,
		Operation:  "update",
	}

	steps := buildServerErrorSteps(ctx)

	if len(steps) < 3 {
		t.Errorf("Expected at least 3 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	if !strings.Contains(stepsText, "status") {
		t.Error("Expected status page mention")
	}
	if !strings.Contains(stepsText, "retry") {
		t.Error("Expected retry mention")
	}
}

func TestBuildValidationErrorSteps_Content(t *testing.T) {
	ctx := ErrorContext{
		Type:         "validation",
		HTTPStatus:   400,
		ResourceType: "Monitor",
		Operation:    "create",
	}

	steps := buildValidationErrorSteps(ctx)

	if len(steps) < 4 {
		t.Errorf("Expected at least 4 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	if !strings.Contains(stepsText, "required fields") {
		t.Error("Expected required fields mention")
	}
	if !strings.Contains(stepsText, "documentation") {
		t.Error("Expected documentation mention")
	}
}

func TestBuildValidationErrorSteps_UnknownResourceType(t *testing.T) {
	ctx := ErrorContext{
		Type:         "validation",
		HTTPStatus:   400,
		ResourceType: "UnknownResource",
		Operation:    "create",
	}

	steps := buildValidationErrorSteps(ctx)
	stepsText := strings.Join(steps, "\n")

	if !strings.Contains(stepsText, "Check field types and value constraints") {
		t.Error("Expected default validation guidance for unknown resource type")
	}
}

// Test edge cases

func TestBuildTroubleshootingSteps_EmptyResourceID(t *testing.T) {
	ctx := ErrorContext{
		Type:         "not_found",
		HTTPStatus:   404,
		ResourceType: "Monitor",
		ResourceID:   "", // Empty resource ID
		Operation:    "read",
	}

	steps := BuildTroubleshootingSteps(ctx)

	// Should still generate steps, but without resource ID specific guidance
	if !strings.Contains(steps, "Troubleshooting:") {
		t.Error("Expected troubleshooting header even with empty resource ID")
	}
}

func TestJoinSteps(t *testing.T) {
	steps := []string{"Step 1", "Step 2", "Step 3"}
	result := joinSteps(steps)

	expected := "Step 1\nStep 2\nStep 3\n"
	if result != expected {
		t.Errorf("Expected joined steps:\n%s\nGot:\n%s", expected, result)
	}
}

func TestJoinSteps_Empty(t *testing.T) {
	steps := []string{}
	result := joinSteps(steps)

	if result != "" {
		t.Errorf("Expected empty string for empty steps, got: %s", result)
	}
}
