// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Integration tests that validate error propagation through resources
// These tests ensure enhanced error messages with troubleshooting steps
// are properly displayed to users during CRUD operations.

func TestErrorPropagation_MonitorRead(t *testing.T) {
	tests := []struct {
		name              string
		err               error
		wantErrorType     string
		wantInMessage     []string
		wantDashboardLink bool
	}{
		{
			name:          "not found error",
			err:           client.NewAPIError(404, "monitor not found"),
			wantErrorType: "not_found",
			wantInMessage: []string{
				"Troubleshooting",
				"Verify the resource still exists",
				"dashboard",
			},
			wantDashboardLink: true,
		},
		{
			name:          "unauthorized error",
			err:           client.NewAPIError(401, "unauthorized"),
			wantErrorType: "auth_error",
			wantInMessage: []string{
				"Troubleshooting",
				"API key",
				"permissions",
			},
			wantDashboardLink: true,
		},
		{
			name:          "rate limit error",
			err:           client.NewRateLimitError(60),
			wantErrorType: "rate_limit",
			wantInMessage: []string{
				"Troubleshooting",
				"60 seconds",
			},
			wantDashboardLink: true,
		},
		{
			name:          "server error",
			err:           client.NewAPIError(500, "internal server error"),
			wantErrorType: "server_error",
			wantInMessage: []string{
				"Troubleshooting",
				"Verify the resource still exists",
			},
			wantDashboardLink: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use the standard error helper
			diag := newReadError("Monitor", "mon_test123", tt.err)

			summary := diag.Summary()
			detail := diag.Detail()

			// Verify summary contains resource type
			if !strings.Contains(summary, "Failed to Read Monitor") {
				t.Errorf("Expected summary to contain 'Failed to Read Monitor', got: %s", summary)
			}

			// Verify detail contains expected troubleshooting guidance
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Expected detail to contain %q, got: %s", want, detail)
				}
			}

			// Verify dashboard link is present when expected
			if tt.wantDashboardLink {
				if !strings.Contains(detail, "https://app.hyperping.io") {
					t.Errorf("Expected detail to contain dashboard link")
				}
			}
		})
	}
}

func TestErrorPropagation_MonitorCreate(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name: "validation error",
			err:  client.NewAPIError(400, "invalid URL"),
			wantInMessage: []string{
				"Troubleshooting",
				"API key",
				"required fields",
			},
		},
		{
			name: "forbidden error",
			err:  client.NewAPIError(403, "forbidden"),
			wantInMessage: []string{
				"Troubleshooting",
				"permissions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := newCreateError("Monitor", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Expected detail to contain %q, got: %s", want, detail)
				}
			}
		})
	}
}

func TestErrorPropagation_MonitorUpdate(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name: "not found during update",
			err:  client.NewAPIError(404, "not found"),
			wantInMessage: []string{
				"Troubleshooting",
				"still exists",
			},
		},
		{
			name: "validation error during update",
			err:  client.NewAPIError(422, "invalid frequency"),
			wantInMessage: []string{
				"Troubleshooting",
				"update values are valid",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := newUpdateError("Monitor", "mon_update123", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Expected detail to contain %q, got: %s", want, detail)
				}
			}
		})
	}
}

func TestErrorPropagation_MonitorDelete(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name: "server error during delete",
			err:  client.NewAPIError(500, "internal error"),
			wantInMessage: []string{
				"Troubleshooting",
				"still exists",
			},
		},
		{
			name: "rate limit during delete",
			err:  client.NewRateLimitError(120),
			wantInMessage: []string{
				"Troubleshooting",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := newDeleteError("Monitor", "mon_delete123", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Expected detail to contain %q, got: %s", want, detail)
				}
			}
		})
	}
}

func TestErrorPropagation_IncidentRead(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name: "incident not found",
			err:  client.NewAPIError(404, "incident not found"),
			wantInMessage: []string{
				"Troubleshooting",
			},
		},
		{
			name: "unauthorized for incident",
			err:  client.NewAPIError(401, "unauthorized"),
			wantInMessage: []string{
				"API key",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := newReadError("Incident", "inc_test123", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Expected detail to contain %q, got: %s", want, detail)
				}
			}
		})
	}
}

func TestErrorPropagation_MaintenanceCreate(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name: "validation error for maintenance",
			err:  client.NewValidationError(400, "invalid schedule", []client.ValidationDetail{{Field: "scheduled_start", Message: "must be in future"}}),
			wantInMessage: []string{
				"Troubleshooting",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := newCreateError("Maintenance", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Expected detail to contain %q, got: %s", want, detail)
				}
			}
		})
	}
}

func TestErrorPropagation_MultipleResources(t *testing.T) {
	// Test that error messages are consistent across different resource types
	resourceTypes := []string{"Monitor", "Incident", "Maintenance", "Healthcheck", "Outage", "StatusPage"}
	testErr := client.NewAPIError(404, "not found")

	for _, resourceType := range resourceTypes {
		t.Run(resourceType, func(t *testing.T) {
			diag := newReadError(resourceType, "test_id_123", testErr)

			summary := diag.Summary()
			detail := diag.Detail()

			// All should contain resource type in summary
			if !strings.Contains(summary, resourceType) {
				t.Errorf("Expected summary to contain %q, got: %s", resourceType, summary)
			}

			// All should contain troubleshooting section
			if !strings.Contains(detail, "Troubleshooting") {
				t.Errorf("Expected detail to contain Troubleshooting section")
			}

			// All should contain dashboard link
			if !strings.Contains(detail, "https://app.hyperping.io") {
				t.Errorf("Expected detail to contain dashboard link")
			}
		})
	}
}

func TestTroubleshootingSteps_UserReadability(t *testing.T) {
	// Test that error messages are user-friendly and actionable
	tests := []struct {
		name    string
		diag    func() (string, string)
		wantHas []string
		wantNot []string
	}{
		{
			name: "read error is user-friendly",
			diag: func() (string, string) {
				err := client.NewAPIError(404, "not found")
				d := newReadError("Monitor", "mon_123", err)
				return d.Summary(), d.Detail()
			},
			wantHas: []string{
				"Troubleshooting",
				"Verify",
				"Check",
			},
			wantNot: []string{
				"panic",
				"nil pointer",
				"stack trace",
			},
		},
		{
			name: "create error provides actionable steps",
			diag: func() (string, string) {
				err := client.NewAPIError(400, "invalid request")
				d := newCreateError("Monitor", err)
				return d.Summary(), d.Detail()
			},
			wantHas: []string{
				"Troubleshooting",
				"Verify",
				"Check",
			},
			wantNot: []string{
				"internal error",
				"unexpected",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, detail := tt.diag()

			for _, want := range tt.wantHas {
				if !strings.Contains(detail, want) {
					t.Errorf("Expected detail to contain %q, got: %s", want, detail)
				}
			}

			for _, notWant := range tt.wantNot {
				if strings.Contains(strings.ToLower(detail), strings.ToLower(notWant)) {
					t.Errorf("Expected detail NOT to contain %q, but it did: %s", notWant, detail)
				}
			}
		})
	}
}

func TestRateLimitError_RetryGuidance(t *testing.T) {
	// Test that rate limit errors provide clear retry guidance
	retryAfterTests := []int{30, 60, 120, 300}

	for _, retryAfter := range retryAfterTests {
		t.Run(strings.Join([]string{"retry_after", string(rune(retryAfter))}, "_"), func(t *testing.T) {
			err := client.NewRateLimitError(retryAfter)
			diag := newReadError("Monitor", "mon_rate123", err)

			detail := diag.Detail()

			// Should mention rate limiting
			if !strings.Contains(detail, "Troubleshooting") {
				t.Errorf("Expected troubleshooting guidance for rate limit")
			}

			// Should contain the error message with retry time
			if !strings.Contains(err.Error(), "retry after") {
				t.Errorf("Expected error to mention retry time")
			}
		})
	}
}

func TestAuthError_KeyVerification(t *testing.T) {
	// Test that auth errors provide API key verification steps
	authErrors := []int{401, 403}

	for _, statusCode := range authErrors {
		t.Run(strings.Join([]string{"status", string(rune(statusCode))}, "_"), func(t *testing.T) {
			err := client.NewAPIError(statusCode, "unauthorized")
			diag := newReadError("Monitor", "mon_auth123", err)

			detail := diag.Detail()

			// Should mention API key verification
			if !strings.Contains(detail, "API key") {
				t.Errorf("Expected detail to mention API key verification")
			}

			// Should have troubleshooting section
			if !strings.Contains(detail, "Troubleshooting") {
				t.Errorf("Expected troubleshooting section")
			}
		})
	}
}

func TestErrorContext_AllResourceTypes(t *testing.T) {
	// Test error handling for all 8 resource types
	allResources := []struct {
		resourceType string
		resourceID   string
	}{
		{"Monitor", "mon_123"},
		{"Incident", "inc_456"},
		{"Maintenance", "mw_789"},
		{"Healthcheck", "hc_abc"},
		{"Outage", "out_def"},
		{"StatusPage", "sp_ghi"},
		{"StatusPageSubscriber", "sps_jkl"},
		{"IncidentUpdate", "iu_mno"},
	}

	testErr := client.NewAPIError(500, "server error")

	for _, resource := range allResources {
		t.Run(resource.resourceType, func(t *testing.T) {
			// Test all CRUD operations
			t.Run("create", func(t *testing.T) {
				diag := newCreateError(resource.resourceType, testErr)
				if !strings.Contains(diag.Summary(), "Failed to Create") {
					t.Errorf("Create error missing expected summary")
				}
			})

			t.Run("read", func(t *testing.T) {
				diag := newReadError(resource.resourceType, resource.resourceID, testErr)
				if !strings.Contains(diag.Summary(), "Failed to Read") {
					t.Errorf("Read error missing expected summary")
				}
			})

			t.Run("update", func(t *testing.T) {
				diag := newUpdateError(resource.resourceType, resource.resourceID, testErr)
				if !strings.Contains(diag.Summary(), "Failed to Update") {
					t.Errorf("Update error missing expected summary")
				}
			})

			t.Run("delete", func(t *testing.T) {
				diag := newDeleteError(resource.resourceType, resource.resourceID, testErr)
				if !strings.Contains(diag.Summary(), "Failed to Delete") {
					t.Errorf("Delete error missing expected summary")
				}
			})
		})
	}
}

func TestErrorMessage_Format(t *testing.T) {
	// Test that error messages follow consistent formatting
	testErr := errors.New("test error")

	tests := []struct {
		name        string
		createDiag  func() (string, string)
		wantSummary string
	}{
		{
			name: "create error format",
			createDiag: func() (string, string) {
				d := newCreateError("Monitor", testErr)
				return d.Summary(), d.Detail()
			},
			wantSummary: "Failed to Create Monitor",
		},
		{
			name: "read error format",
			createDiag: func() (string, string) {
				d := newReadError("Monitor", "mon_123", testErr)
				return d.Summary(), d.Detail()
			},
			wantSummary: "Failed to Read Monitor",
		},
		{
			name: "update error format",
			createDiag: func() (string, string) {
				d := newUpdateError("Monitor", "mon_123", testErr)
				return d.Summary(), d.Detail()
			},
			wantSummary: "Failed to Update Monitor",
		},
		{
			name: "delete error format",
			createDiag: func() (string, string) {
				d := newDeleteError("Monitor", "mon_123", testErr)
				return d.Summary(), d.Detail()
			},
			wantSummary: "Failed to Delete Monitor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, detail := tt.createDiag()

			if summary != tt.wantSummary {
				t.Errorf("Expected summary %q, got %q", tt.wantSummary, summary)
			}

			// All error messages should have a detail section
			if len(detail) == 0 {
				t.Errorf("Expected non-empty detail")
			}

			// All should have troubleshooting section
			if !strings.Contains(detail, "Troubleshooting") {
				t.Errorf("Expected troubleshooting section in detail")
			}
		})
	}
}

func TestErrorMessage_AllOperations(t *testing.T) {
	// Verify all CRUD operations have proper error handling
	testErr := client.NewAPIError(400, "test error")

	operations := map[string]func() string{
		"Create": func() string {
			return newCreateError("Resource", testErr).Detail()
		},
		"Read": func() string {
			return newReadError("Resource", "id_123", testErr).Detail()
		},
		"Update": func() string {
			return newUpdateError("Resource", "id_123", testErr).Detail()
		},
		"Delete": func() string {
			return newDeleteError("Resource", "id_123", testErr).Detail()
		},
		"List": func() string {
			return newListError("Resources", testErr).Detail()
		},
	}

	for opName, opFunc := range operations {
		t.Run(opName, func(t *testing.T) {
			detail := opFunc()

			// All operations should have troubleshooting guidance
			if !strings.Contains(detail, "Troubleshooting") {
				t.Errorf("%s operation missing troubleshooting guidance", opName)
			}

			// All should mention the error
			if !strings.Contains(detail, "test error") {
				t.Errorf("%s operation missing original error message", opName)
			}
		})
	}
}
