// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/sony/gobreaker"

	hyperping "github.com/develeap/hyperping-go"
)

// ---------------------------------------------------------------------------
// DetectErrorContext
// ---------------------------------------------------------------------------

func TestDetectErrorContext(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType string
		resourceID   string
		operation    string
		err          error
		wantType     string
		wantStatus   int
		wantRetry    int
		wantMessage  string
	}{
		{
			name:         "not found 404",
			resourceType: "Monitor",
			resourceID:   "mon_123",
			operation:    "read",
			err:          hyperping.NewAPIError(404, "resource not found"),
			wantType:     "not_found",
			wantStatus:   404,
			wantMessage:  "404",
		},
		{
			name:         "unauthorized 401",
			resourceType: "Monitor",
			resourceID:   "mon_123",
			operation:    "read",
			err:          hyperping.NewAPIError(401, "invalid API key"),
			wantType:     "auth_error",
			wantStatus:   401,
		},
		{
			name:         "forbidden 403",
			resourceType: "Monitor",
			resourceID:   "mon_123",
			operation:    "update",
			err:          hyperping.NewAPIError(403, "access forbidden"),
			wantType:     "auth_error",
			wantStatus:   403,
		},
		{
			name:         "rate limit 429",
			resourceType: "Monitor",
			resourceID:   "",
			operation:    "create",
			err:          hyperping.NewRateLimitError(120),
			wantType:     "rate_limit",
			wantStatus:   429,
			wantRetry:    120,
		},
		{
			name:         "server error 500",
			resourceType: "Incident",
			resourceID:   "inc_123",
			operation:    "delete",
			err:          hyperping.NewAPIError(500, "internal server error"),
			wantType:     "server_error",
			wantStatus:   500,
		},
		{
			name:         "validation error 400",
			resourceType: "Monitor",
			resourceID:   "",
			operation:    "create",
			err: hyperping.NewValidationError(400, "validation failed", []hyperping.ValidationDetail{
				{Field: "url", Message: "Invalid URL format"},
			}),
			wantType:   "validation",
			wantStatus: 400,
		},
		{
			name:         "unknown error",
			resourceType: "Monitor",
			resourceID:   "mon_123",
			operation:    "read",
			err:          errors.New("generic error"),
			wantType:     "unknown",
			wantStatus:   0,
			wantMessage:  "generic error",
		},
		{
			name:         "nil error",
			resourceType: "Monitor",
			resourceID:   "mon_123",
			operation:    "read",
			err:          nil,
			wantType:     "unknown",
			wantStatus:   0,
		},
		{
			name:         "circuit breaker open",
			resourceType: "Monitor",
			resourceID:   "mon_123",
			operation:    "read",
			err:          gobreaker.ErrOpenState,
			wantType:     "circuit_breaker",
			wantStatus:   0,
			wantMessage:  "circuit breaker is open",
		},
		{
			name:         "empty resource ID for validation",
			resourceType: "Monitor",
			resourceID:   "",
			operation:    "create",
			err:          hyperping.NewAPIError(400, "validation error"),
			wantType:     "validation",
			wantStatus:   400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := DetectErrorContext(tt.resourceType, tt.resourceID, tt.operation, tt.err)

			if ctx.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", ctx.Type, tt.wantType)
			}
			if ctx.HTTPStatus != tt.wantStatus {
				t.Errorf("HTTPStatus = %d, want %d", ctx.HTTPStatus, tt.wantStatus)
			}
			if ctx.ResourceType != tt.resourceType {
				t.Errorf("ResourceType = %q, want %q", ctx.ResourceType, tt.resourceType)
			}
			if ctx.ResourceID != tt.resourceID {
				t.Errorf("ResourceID = %q, want %q", ctx.ResourceID, tt.resourceID)
			}
			if ctx.Operation != tt.operation {
				t.Errorf("Operation = %q, want %q", ctx.Operation, tt.operation)
			}
			if tt.wantRetry != 0 && ctx.RetryAfter != tt.wantRetry {
				t.Errorf("RetryAfter = %d, want %d", ctx.RetryAfter, tt.wantRetry)
			}
			if tt.wantMessage != "" && !strings.Contains(ctx.Message, tt.wantMessage) {
				t.Errorf("Message = %q, want it to contain %q", ctx.Message, tt.wantMessage)
			}
			if tt.err == nil && ctx.Message != "" {
				t.Errorf("Message = %q for nil error, want empty", ctx.Message)
			}
		})
	}
}

func TestDetectErrorContext_AllFields(t *testing.T) {
	t.Parallel()

	err := hyperping.NewRateLimitError(90)
	ctx := DetectErrorContext("Maintenance", "maint_456", "update", err)

	if ctx.Type == "" {
		t.Error("Type is empty")
	}
	if ctx.HTTPStatus == 0 {
		t.Error("HTTPStatus is zero")
	}
	if ctx.RetryAfter == 0 {
		t.Error("RetryAfter is zero")
	}
	if ctx.ResourceType == "" {
		t.Error("ResourceType is empty")
	}
	if ctx.ResourceID == "" {
		t.Error("ResourceID is empty")
	}
	if ctx.Operation == "" {
		t.Error("Operation is empty")
	}
	if ctx.Message == "" {
		t.Error("Message is empty")
	}

	if ctx.Type != "rate_limit" {
		t.Errorf("Type = %q, want %q", ctx.Type, "rate_limit")
	}
	if ctx.HTTPStatus != 429 {
		t.Errorf("HTTPStatus = %d, want 429", ctx.HTTPStatus)
	}
	if ctx.RetryAfter != 90 {
		t.Errorf("RetryAfter = %d, want 90", ctx.RetryAfter)
	}
}

func TestDetectErrorContext_DifferentOperations(t *testing.T) {
	t.Parallel()

	operations := []string{"create", "read", "update", "delete"}
	err := hyperping.NewAPIError(404, "not found")

	for _, op := range operations {
		t.Run(op, func(t *testing.T) {
			t.Parallel()
			ctx := DetectErrorContext("Monitor", "mon_123", op, err)
			if ctx.Operation != op {
				t.Errorf("Operation = %q, want %q", ctx.Operation, op)
			}
			if ctx.Type != "not_found" {
				t.Errorf("Type = %q, want %q", ctx.Type, "not_found")
			}
		})
	}
}

func TestDetectErrorContext_DifferentResources(t *testing.T) {
	t.Parallel()

	resources := []string{"Monitor", "Incident", "Maintenance", "Healthcheck", "Status Page"}
	err := hyperping.NewAPIError(401, "unauthorized")

	for _, resource := range resources {
		t.Run(resource, func(t *testing.T) {
			t.Parallel()
			ctx := DetectErrorContext(resource, "test_123", "read", err)
			if ctx.ResourceType != resource {
				t.Errorf("ResourceType = %q, want %q", ctx.ResourceType, resource)
			}
			if ctx.Type != "auth_error" {
				t.Errorf("Type = %q, want %q", ctx.Type, "auth_error")
			}
		})
	}
}

func TestDetectErrorContext_ServerErrorVariants(t *testing.T) {
	t.Parallel()

	codes := []int{500, 502, 503, 504}
	for _, code := range codes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			t.Parallel()
			err := hyperping.NewAPIError(code, fmt.Sprintf("server error %d", code))
			ctx := DetectErrorContext("Monitor", "mon_123", "read", err)

			if ctx.Type != "server_error" {
				t.Errorf("Type = %q, want %q", ctx.Type, "server_error")
			}
			if ctx.HTTPStatus != code {
				t.Errorf("HTTPStatus = %d, want %d", ctx.HTTPStatus, code)
			}
		})
	}
}

func TestDetectErrorContext_ValidationErrorVariants(t *testing.T) {
	t.Parallel()

	codes := []int{400, 422}
	for _, code := range codes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			t.Parallel()
			err := hyperping.NewAPIError(code, fmt.Sprintf("validation %d", code))
			ctx := DetectErrorContext("Monitor", "", "create", err)

			if ctx.Type != "validation" {
				t.Errorf("Type = %q, want %q", ctx.Type, "validation")
			}
			if ctx.HTTPStatus != code {
				t.Errorf("HTTPStatus = %d, want %d", ctx.HTTPStatus, code)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ErrorContext.String
// ---------------------------------------------------------------------------

func TestErrorContext_String(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Type:         "not_found",
		HTTPStatus:   404,
		ResourceType: "Monitor",
		ResourceID:   "mon_123",
		Operation:    "read",
	}

	str := ctx.String()
	for _, want := range []string{"not_found", "404", "Monitor", "mon_123", "read"} {
		if !strings.Contains(str, want) {
			t.Errorf("String() = %q, want it to contain %q", str, want)
		}
	}
}

// ---------------------------------------------------------------------------
// extractRetryAfter
// ---------------------------------------------------------------------------

func TestExtractRetryAfter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      error
		expected int
	}{
		// Pattern matches
		{"60 seconds", errors.New("API error (status 429): rate limit exceeded - retry after 60 seconds"), 60},
		{"120 seconds", errors.New("rate limit exceeded - retry after 120 seconds"), 120},
		{"single second", errors.New("retry after 1 second"), 1},
		{"300 seconds", errors.New("Please retry after 300 seconds"), 300},
		{"uppercase", errors.New("RETRY AFTER 90 SECONDS"), 90},
		{"mixed case", errors.New("Retry After 45 Seconds"), 45},
		{"lowercase", errors.New("retry after 30 seconds"), 30},
		{"large number", errors.New("retry after 3600 seconds"), 3600},
		// Default cases
		{"no retry pattern", errors.New("rate limit exceeded"), 60},
		{"different format", errors.New("too many requests"), 60},
		{"nil error", nil, 60},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractRetryAfter(tt.err)
			if result != tt.expected {
				t.Errorf("extractRetryAfter() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractStatusCode
// ---------------------------------------------------------------------------

func TestExtractStatusCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		err         error
		defaultCode int
		expected    int
	}{
		// Successful extractions
		{"500 status", errors.New("API error (status 500): internal server error"), 999, 500},
		{"502 status", errors.New("status 502 bad gateway"), 999, 502},
		{"503 status", errors.New("service unavailable status 503"), 999, 503},
		{"400 status", errors.New("API error (status 400): bad request"), 999, 400},
		// Default fallback
		{"no status code", errors.New("generic error"), 500, 500},
		{"nil error", nil, 400, 400},
		{"invalid status pattern", errors.New("status code not found"), 503, 503},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := extractStatusCode(tt.err, tt.defaultCode)
			if result != tt.expected {
				t.Errorf("extractStatusCode() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// newReadAfterCreateError
// ---------------------------------------------------------------------------

func TestNewReadAfterCreateError(t *testing.T) {
	t.Parallel()

	err := errors.New("resource not ready")
	diag := newReadAfterCreateError("Monitor", "mon_123", err)

	if !strings.Contains(diag.Summary(), "Created But Read Failed") {
		t.Errorf("Summary = %q, want it to contain 'Created But Read Failed'", diag.Summary())
	}
	if !strings.Contains(diag.Detail(), "mon_123") {
		t.Errorf("Detail missing resource ID, got: %s", diag.Detail())
	}
	if !strings.Contains(diag.Detail(), "terraform import") {
		t.Errorf("Detail missing import instructions, got: %s", diag.Detail())
	}
}

// ---------------------------------------------------------------------------
// newImportError
// ---------------------------------------------------------------------------

func TestNewImportError(t *testing.T) {
	t.Parallel()

	err := errors.New("invalid format")
	diag := newImportError("Monitor", err)

	if !strings.Contains(diag.Summary(), "Import Failed") {
		t.Errorf("Summary = %q, want it to contain 'Import Failed'", diag.Summary())
	}
	if !strings.Contains(diag.Detail(), "invalid format") {
		t.Errorf("Detail missing error message, got: %s", diag.Detail())
	}
}

// ---------------------------------------------------------------------------
// newUnexpectedConfigTypeError
// ---------------------------------------------------------------------------

func TestNewUnexpectedConfigTypeError(t *testing.T) {
	t.Parallel()

	diag := newUnexpectedConfigTypeError("*provider.hyperpingClient", "string")

	if !strings.Contains(diag.Summary(), "Unexpected Resource Configure Type") {
		t.Errorf("Summary = %q, want it to contain 'Unexpected Resource Configure Type'", diag.Summary())
	}
	if !strings.Contains(diag.Detail(), "*provider.hyperpingClient") {
		t.Errorf("Detail missing expected type, got: %s", diag.Detail())
	}
	if !strings.Contains(diag.Detail(), "provider bug") {
		t.Errorf("Detail missing 'provider bug', got: %s", diag.Detail())
	}
}

// ---------------------------------------------------------------------------
// TestAllErrorHelpersReturnValidDiagnostics -- ensures every helper produces
// non-empty Summary and Detail.
// ---------------------------------------------------------------------------

func TestAllErrorHelpersReturnValidDiagnostics(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")

	tests := []struct {
		name   string
		create func() (string, string)
	}{
		{"newReadAfterCreateError", func() (string, string) {
			d := newReadAfterCreateError("Resource", "new_123", testErr)
			return d.Summary(), d.Detail()
		}},
		{"newImportError", func() (string, string) { d := newImportError("Resource", testErr); return d.Summary(), d.Detail() }},
		{"newUnexpectedConfigTypeError", func() (string, string) {
			d := newUnexpectedConfigTypeError("*hyperping.Client", 42)
			return d.Summary(), d.Detail()
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			summary, detail := tt.create()

			if summary == "" {
				t.Error("Summary is empty")
			}
			if detail == "" {
				t.Error("Detail is empty")
			}
			if len(summary) < 5 {
				t.Errorf("Summary too short: %s", summary)
			}
			if len(detail) < 10 {
				t.Errorf("Detail too short: %s", detail)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidValueReference
// ---------------------------------------------------------------------------

func TestValidValueReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		resourceType string
		wantEmpty    bool
		wantContains []string
	}{
		{
			name:         "Monitor",
			resourceType: "Monitor",
			wantContains: []string{
				"protocol", "http_method", "check_frequency",
				"expected_status_code", "regions", "alerts_wait",
				// Verify actual values from client constants
				"http", "port", "icmp", "GET", "POST", "london", "frankfurt",
			},
		},
		{
			name:         "Maintenance Window",
			resourceType: "Maintenance Window",
			wantContains: []string{"notification_option", "scheduled", "immediate"},
		},
		{
			name:         "Incident",
			resourceType: "Incident",
			wantContains: []string{"type", "outage", "incident"},
		},
		{
			name:         "unknown type returns empty",
			resourceType: "UnknownType",
			wantEmpty:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ref := ValidValueReference(tt.resourceType)

			if tt.wantEmpty {
				if ref != "" {
					t.Errorf("Expected empty string, got: %s", ref)
				}
				return
			}

			if ref == "" {
				t.Fatalf("Expected non-empty reference for %s", tt.resourceType)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(ref, want) {
					t.Errorf("Reference missing %q, got: %s", want, ref)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// BuildTroubleshootingSteps
// ---------------------------------------------------------------------------

func TestBuildTroubleshootingSteps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        ErrorContext
		wantInText []string
	}{
		{
			name: "not_found",
			ctx: ErrorContext{
				Type: "not_found", HTTPStatus: 404,
				ResourceType: "Monitor", ResourceID: "mon_test123", Operation: "read",
			},
			wantInText: []string{
				"Troubleshooting:", "Verify the Monitor still exists",
				"https://app.hyperping.io/monitors", "mon_test123",
				"terraform state show",
			},
		},
		{
			name: "auth_error 401",
			ctx: ErrorContext{
				Type: "auth_error", HTTPStatus: 401,
				ResourceType: "Monitor", Operation: "create",
			},
			wantInText: []string{
				"HYPERPING_API_KEY", "starts with 'sk_'", "curl",
				"https://app.hyperping.io/settings/api",
			},
		},
		{
			name: "auth_error 403",
			ctx: ErrorContext{
				Type: "auth_error", HTTPStatus: 403,
				ResourceType: "Incident", Operation: "delete",
			},
			wantInText: []string{"required permissions", "Incident resources"},
		},
		{
			name: "rate_limit with retry",
			ctx: ErrorContext{
				Type: "rate_limit", HTTPStatus: 429,
				RetryAfter: 120, Operation: "update",
			},
			wantInText: []string{
				"Wait 120 seconds", "terraform apply -parallelism=1",
				"rate-limits",
			},
		},
		{
			name: "rate_limit default wait",
			ctx: ErrorContext{
				Type: "rate_limit", HTTPStatus: 429,
				RetryAfter: 0, Operation: "update",
			},
			wantInText: []string{"Wait 60 seconds"},
		},
		{
			name: "server_error",
			ctx: ErrorContext{
				Type: "server_error", HTTPStatus: 500, Operation: "create",
			},
			wantInText: []string{"https://status.hyperping.app", "retry", "support"},
		},
		{
			name: "validation generic",
			ctx: ErrorContext{
				Type: "validation", HTTPStatus: 400,
				ResourceType: "Monitor", Operation: "create",
			},
			wantInText: []string{"required fields", "validation", "Monitor documentation"},
		},
		{
			name: "validation Monitor specifics",
			ctx: ErrorContext{
				Type: "validation", HTTPStatus: 400,
				ResourceType: "Monitor", Operation: "create",
			},
			wantInText: []string{
				"Frequency must be one of",
				"Timeout must be one of",
				"valid Hyperping region codes",
			},
		},
		{
			name: "validation Incident specifics",
			ctx: ErrorContext{
				Type: "validation", HTTPStatus: 422,
				ResourceType: "Incident", Operation: "create",
			},
			wantInText: []string{"Status must be one of", "Severity must be one of"},
		},
		{
			name: "validation Maintenance specifics",
			ctx: ErrorContext{
				Type: "validation", HTTPStatus: 400,
				ResourceType: "Maintenance", Operation: "create",
			},
			wantInText: []string{"ISO 8601", "scheduledEnd must be after scheduledStart"},
		},
		{
			name: "unknown error type",
			ctx: ErrorContext{
				Type: "unknown", HTTPStatus: 0, Operation: "read",
			},
			wantInText: []string{"network connectivity", "service status", "provider maintainers"},
		},
		{
			name: "not_found empty resource ID",
			ctx: ErrorContext{
				Type: "not_found", HTTPStatus: 404,
				ResourceType: "Monitor", ResourceID: "", Operation: "read",
			},
			wantInText: []string{"Troubleshooting:"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			steps := BuildTroubleshootingSteps(tt.ctx)

			for _, want := range tt.wantInText {
				if !strings.Contains(steps, want) {
					t.Errorf("Steps missing %q, got:\n%s", want, steps)
				}
			}
		})
	}
}

func TestBuildTroubleshootingSteps_ContainsDashboardLink(t *testing.T) {
	t.Parallel()

	resourceTypes := []string{"Monitor", "Incident", "Maintenance", "Statuspage"}
	for _, resourceType := range resourceTypes {
		t.Run(resourceType, func(t *testing.T) {
			t.Parallel()
			ctx := ErrorContext{
				Type: "not_found", HTTPStatus: 404,
				ResourceType: resourceType, ResourceID: "test_id", Operation: "read",
			}

			steps := BuildTroubleshootingSteps(ctx)
			if !strings.Contains(steps, "app.hyperping.io") {
				t.Errorf("Expected dashboard link for %s, got:\n%s", resourceType, steps)
			}
		})
	}
}

func TestBuildTroubleshootingSteps_ContainsRetryTime(t *testing.T) {
	t.Parallel()

	retryTimes := []int{30, 60, 90, 120, 180}
	for _, retryTime := range retryTimes {
		t.Run(fmt.Sprintf("RetryAfter%d", retryTime), func(t *testing.T) {
			t.Parallel()
			ctx := ErrorContext{
				Type: "rate_limit", HTTPStatus: 429,
				RetryAfter: retryTime, Operation: "create",
			}

			steps := BuildTroubleshootingSteps(ctx)
			expectedMsg := fmt.Sprintf("Wait %d seconds", retryTime)
			if !strings.Contains(steps, expectedMsg) {
				t.Errorf("Steps missing %q, got:\n%s", expectedMsg, steps)
			}
		})
	}
}

func TestBuildTroubleshootingSteps_ResourceSpecific(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			ctx := ErrorContext{
				Type: tt.errorType, HTTPStatus: 400,
				ResourceType: tt.resourceType, Operation: "create",
			}

			steps := BuildTroubleshootingSteps(ctx)
			if !strings.Contains(strings.ToLower(steps), strings.ToLower(tt.expectedText)) {
				t.Errorf("Steps missing resource-specific text %q for %s %s", tt.expectedText, tt.resourceType, tt.errorType)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// build*Steps internal helpers
// ---------------------------------------------------------------------------

func TestBuildNotFoundSteps_Content(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Type: "not_found", HTTPStatus: 404,
		ResourceType: "Monitor", ResourceID: "mon_test", Operation: "read",
	}

	steps := buildNotFoundSteps(ctx)
	if len(steps) < 4 {
		t.Errorf("Expected at least 4 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	for _, want := range []string{"dashboard", "deleted outside"} {
		if !strings.Contains(stepsText, want) {
			t.Errorf("Steps missing %q", want)
		}
	}
}

func TestBuildAuthErrorSteps_Content(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Type: "auth_error", HTTPStatus: 401, Operation: "read",
	}

	steps := buildAuthErrorSteps(ctx)
	if len(steps) < 3 {
		t.Errorf("Expected at least 3 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	for _, want := range []string{"API key", "curl"} {
		if !strings.Contains(stepsText, want) {
			t.Errorf("Steps missing %q", want)
		}
	}
}

func TestBuildRateLimitSteps_Content(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Type: "rate_limit", HTTPStatus: 429, RetryAfter: 45, Operation: "create",
	}

	steps := buildRateLimitSteps(ctx)
	if len(steps) < 3 {
		t.Errorf("Expected at least 3 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	for _, want := range []string{"45 seconds", "parallelism"} {
		if !strings.Contains(stepsText, want) {
			t.Errorf("Steps missing %q", want)
		}
	}
}

func TestBuildServerErrorSteps_Content(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Type: "server_error", HTTPStatus: 500, Operation: "update",
	}

	steps := buildServerErrorSteps(ctx)
	if len(steps) < 3 {
		t.Errorf("Expected at least 3 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	for _, want := range []string{"status", "retry"} {
		if !strings.Contains(stepsText, want) {
			t.Errorf("Steps missing %q", want)
		}
	}
}

func TestBuildValidationErrorSteps_Content(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Type: "validation", HTTPStatus: 400,
		ResourceType: "Monitor", Operation: "create",
	}

	steps := buildValidationErrorSteps(ctx)
	if len(steps) < 4 {
		t.Errorf("Expected at least 4 steps, got: %d", len(steps))
	}

	stepsText := strings.Join(steps, "\n")
	for _, want := range []string{"required fields", "documentation"} {
		if !strings.Contains(stepsText, want) {
			t.Errorf("Steps missing %q", want)
		}
	}
}

func TestBuildValidationErrorSteps_UnknownResourceType(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Type: "validation", HTTPStatus: 400,
		ResourceType: "UnknownResource", Operation: "create",
	}

	steps := buildValidationErrorSteps(ctx)
	stepsText := strings.Join(steps, "\n")

	if !strings.Contains(stepsText, "Check field types and value constraints") {
		t.Error("Expected default validation guidance for unknown resource type")
	}
}

// ---------------------------------------------------------------------------
// getDashboardURL
// ---------------------------------------------------------------------------

func TestGetDashboardURL(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			url := getDashboardURL(tt.resourceType)
			if url != tt.expectedURL {
				t.Errorf("getDashboardURL(%q) = %q, want %q", tt.resourceType, url, tt.expectedURL)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// joinSteps
// ---------------------------------------------------------------------------

func TestJoinSteps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		steps    []string
		expected string
	}{
		{"multiple steps", []string{"Step 1", "Step 2", "Step 3"}, "Step 1\nStep 2\nStep 3\n"},
		{"empty steps", []string{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := joinSteps(tt.steps)
			if result != tt.expected {
				t.Errorf("joinSteps() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Enhanced error creation functions (New*ErrorWithContext)
// ---------------------------------------------------------------------------

func TestNewReadErrorWithContext(t *testing.T) {
	t.Parallel()

	err := hyperping.NewAPIError(404, "monitor not found")
	diag := NewReadErrorWithContext("Monitor", "mon_abc123", err)

	if diag.Summary() != "Failed to Read Monitor" {
		t.Errorf("Summary = %q, want 'Failed to Read Monitor'", diag.Summary())
	}

	detail := diag.Detail()
	for _, want := range []string{"mon_abc123", "Troubleshooting:", "dashboard"} {
		if !strings.Contains(detail, want) {
			t.Errorf("Detail missing %q, got: %s", want, detail)
		}
	}
}

func TestNewCreateErrorWithContext(t *testing.T) {
	t.Parallel()

	err := hyperping.NewAPIError(401, "unauthorized")
	diag := NewCreateErrorWithContext("Incident", err)

	if diag.Summary() != "Failed to Create Incident" {
		t.Errorf("Summary = %q, want 'Failed to Create Incident'", diag.Summary())
	}

	detail := diag.Detail()
	for _, want := range []string{"Troubleshooting:", "HYPERPING_API_KEY"} {
		if !strings.Contains(detail, want) {
			t.Errorf("Detail missing %q, got: %s", want, detail)
		}
	}
}

func TestNewUpdateErrorWithContext(t *testing.T) {
	t.Parallel()

	err := hyperping.NewRateLimitError(90)
	diag := NewUpdateErrorWithContext("Monitor", "mon_xyz789", err)

	if diag.Summary() != "Failed to Update Monitor" {
		t.Errorf("Summary = %q, want 'Failed to Update Monitor'", diag.Summary())
	}

	detail := diag.Detail()
	for _, want := range []string{"mon_xyz789", "Wait 90 seconds", "parallelism"} {
		if !strings.Contains(detail, want) {
			t.Errorf("Detail missing %q, got: %s", want, detail)
		}
	}
}

func TestNewDeleteErrorWithContext(t *testing.T) {
	t.Parallel()

	err := hyperping.NewAPIError(500, "internal server error")
	diag := NewDeleteErrorWithContext("Maintenance", "maint_test", err)

	if diag.Summary() != "Failed to Delete Maintenance" {
		t.Errorf("Summary = %q, want 'Failed to Delete Maintenance'", diag.Summary())
	}

	detail := diag.Detail()
	for _, want := range []string{"maint_test", "status.hyperping.app"} {
		if !strings.Contains(detail, want) {
			t.Errorf("Detail missing %q, got: %s", want, detail)
		}
	}
}

func TestNewReadErrorWithContext_CircuitBreaker(t *testing.T) {
	t.Parallel()

	err := gobreaker.ErrOpenState
	d := NewReadErrorWithContext("Monitor", "mon_abc123", err)

	if !strings.Contains(d.Summary(), "Failed to Read Monitor") {
		t.Errorf("Summary = %q, want it to contain 'Failed to Read Monitor'", d.Summary())
	}

	detail := d.Detail()
	for _, want := range []string{
		"circuit breaker is open",
		"Wait 30 seconds",
		"-parallelism=1",
		"status.hyperping.app",
	} {
		if !strings.Contains(detail, want) {
			t.Errorf("Detail missing %q, got: %s", want, detail)
		}
	}
}

// ---------------------------------------------------------------------------
// Enhanced error message formatting
// ---------------------------------------------------------------------------

func TestEnhancedErrorMessage_Format(t *testing.T) {
	t.Parallel()

	err := errors.New("test error")
	diag := NewReadErrorWithContext("Monitor", "mon_test", err)
	detail := diag.Detail()

	for _, want := range []string{"Unable to read", "got error:", "Troubleshooting:\n"} {
		if !strings.Contains(detail, want) {
			t.Errorf("Detail missing %q, got: %s", want, detail)
		}
	}
}

func TestEnhancedErrorMessage_AllOperations(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")

	tests := []struct {
		op      string
		summary string
	}{
		{"Read", NewReadErrorWithContext("Monitor", "mon_test", testErr).Summary()},
		{"Create", NewCreateErrorWithContext("Monitor", testErr).Summary()},
		{"Update", NewUpdateErrorWithContext("Monitor", "mon_test", testErr).Summary()},
		{"Delete", NewDeleteErrorWithContext("Monitor", "mon_test", testErr).Summary()},
	}

	for _, tt := range tests {
		t.Run(tt.op, func(t *testing.T) {
			t.Parallel()
			expected := fmt.Sprintf("Failed to %s Monitor", tt.op)
			if tt.summary != expected {
				t.Errorf("Summary = %q, want %q", tt.summary, expected)
			}
		})
	}
}
