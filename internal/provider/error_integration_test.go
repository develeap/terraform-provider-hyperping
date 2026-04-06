// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"strings"
	"testing"

	hyperping "github.com/develeap/hyperping-go"
)

// Integration tests that validate error propagation through resources.
// These tests ensure enhanced error messages with troubleshooting steps
// are properly displayed to users during CRUD operations.

// ---------------------------------------------------------------------------
// Error propagation: Monitor CRUD (WithContext versions)
// ---------------------------------------------------------------------------

func TestErrorPropagation_MonitorRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		err               error
		wantErrorType     string
		wantInMessage     []string
		wantDashboardLink bool
	}{
		{
			name:          "not found error",
			err:           hyperping.NewAPIError(404, "monitor not found"),
			wantErrorType: "not_found",
			wantInMessage: []string{
				"Troubleshooting",
				"Verify the Monitor still exists",
				"dashboard",
			},
			wantDashboardLink: true,
		},
		{
			name:          "unauthorized error",
			err:           hyperping.NewAPIError(401, "unauthorized"),
			wantErrorType: "auth_error",
			wantInMessage: []string{
				"Troubleshooting",
				"API key",
			},
			wantDashboardLink: false,
		},
		{
			name:          "rate limit error",
			err:           hyperping.NewRateLimitError(60),
			wantErrorType: "rate_limit",
			wantInMessage: []string{
				"Troubleshooting",
				"60 seconds",
			},
			wantDashboardLink: false,
		},
		{
			name:          "server error",
			err:           hyperping.NewAPIError(500, "internal server error"),
			wantErrorType: "server_error",
			wantInMessage: []string{
				"Troubleshooting",
				"status.hyperping.app",
			},
			wantDashboardLink: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diag := NewReadErrorWithContext("Monitor", "mon_test123", tt.err)

			summary := diag.Summary()
			detail := diag.Detail()

			if !strings.Contains(summary, "Failed to Read Monitor") {
				t.Errorf("Summary = %q, want it to contain 'Failed to Read Monitor'", summary)
			}

			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Detail missing %q, got: %s", want, detail)
				}
			}

			if tt.wantDashboardLink && !strings.Contains(detail, "app.hyperping.io") {
				t.Error("Detail missing dashboard link")
			}
		})
	}
}

func TestErrorPropagation_MonitorCreate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name: "validation error",
			err:  hyperping.NewAPIError(400, "invalid URL"),
			wantInMessage: []string{
				"Troubleshooting",
				"required fields",
			},
		},
		{
			name: "forbidden error",
			err:  hyperping.NewAPIError(403, "forbidden"),
			wantInMessage: []string{
				"Troubleshooting",
				"permissions",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diag := NewCreateErrorWithContext("Monitor", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Detail missing %q, got: %s", want, detail)
				}
			}
		})
	}
}

func TestErrorPropagation_MonitorUpdate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name: "not found during update",
			err:  hyperping.NewAPIError(404, "not found"),
			wantInMessage: []string{
				"Troubleshooting",
				"still exists",
			},
		},
		{
			name: "validation error during update",
			err:  hyperping.NewAPIError(422, "invalid frequency"),
			wantInMessage: []string{
				"Troubleshooting",
				"required fields",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diag := NewUpdateErrorWithContext("Monitor", "mon_update123", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Detail missing %q, got: %s", want, detail)
				}
			}
		})
	}
}

func TestErrorPropagation_MonitorDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name: "server error during delete",
			err:  hyperping.NewAPIError(500, "internal error"),
			wantInMessage: []string{
				"Troubleshooting",
				"status.hyperping.app",
			},
		},
		{
			name: "rate limit during delete",
			err:  hyperping.NewRateLimitError(120),
			wantInMessage: []string{
				"Troubleshooting",
				"120 seconds",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diag := NewDeleteErrorWithContext("Monitor", "mon_delete123", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Detail missing %q, got: %s", want, detail)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Error propagation: other resource types
// ---------------------------------------------------------------------------

func TestErrorPropagation_IncidentRead(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		err           error
		wantInMessage []string
	}{
		{
			name:          "incident not found",
			err:           hyperping.NewAPIError(404, "incident not found"),
			wantInMessage: []string{"Troubleshooting"},
		},
		{
			name:          "unauthorized for incident",
			err:           hyperping.NewAPIError(401, "unauthorized"),
			wantInMessage: []string{"API key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			diag := NewReadErrorWithContext("Incident", "inc_test123", tt.err)

			detail := diag.Detail()
			for _, want := range tt.wantInMessage {
				if !strings.Contains(detail, want) {
					t.Errorf("Detail missing %q, got: %s", want, detail)
				}
			}
		})
	}
}

func TestErrorPropagation_MaintenanceCreate(t *testing.T) {
	t.Parallel()

	err := hyperping.NewValidationError(400, "invalid schedule", []hyperping.ValidationDetail{
		{Field: "scheduled_start", Message: "must be in future"},
	})
	diag := NewCreateErrorWithContext("Maintenance", err)

	detail := diag.Detail()
	if !strings.Contains(detail, "Troubleshooting") {
		t.Errorf("Detail missing 'Troubleshooting', got: %s", detail)
	}
}

// ---------------------------------------------------------------------------
// Cross-resource consistency (WithContext versions)
// ---------------------------------------------------------------------------

func TestErrorPropagation_MultipleResources(t *testing.T) {
	t.Parallel()

	resourceTypes := []string{"Monitor", "Incident", "Maintenance", "Healthcheck", "Outage", "StatusPage"}
	testErr := hyperping.NewAPIError(404, "not found")

	for _, resourceType := range resourceTypes {
		t.Run(resourceType, func(t *testing.T) {
			t.Parallel()
			diag := NewReadErrorWithContext(resourceType, "test_id_123", testErr)

			summary := diag.Summary()
			detail := diag.Detail()

			if !strings.Contains(summary, resourceType) {
				t.Errorf("Summary missing %q, got: %s", resourceType, summary)
			}
			if !strings.Contains(detail, "Troubleshooting") {
				t.Error("Detail missing Troubleshooting section")
			}
		})
	}
}

func TestErrorContext_AllResourceTypes_CRUD(t *testing.T) {
	t.Parallel()

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

	testErr := hyperping.NewAPIError(500, "server error")

	for _, resource := range allResources {
		t.Run(resource.resourceType, func(t *testing.T) {
			t.Parallel()

			t.Run("create", func(t *testing.T) {
				t.Parallel()
				diag := NewCreateErrorWithContext(resource.resourceType, testErr)
				if !strings.Contains(diag.Summary(), "Failed to Create") {
					t.Errorf("Create summary missing 'Failed to Create', got: %s", diag.Summary())
				}
			})

			t.Run("read", func(t *testing.T) {
				t.Parallel()
				diag := NewReadErrorWithContext(resource.resourceType, resource.resourceID, testErr)
				if !strings.Contains(diag.Summary(), "Failed to Read") {
					t.Errorf("Read summary missing 'Failed to Read', got: %s", diag.Summary())
				}
			})

			t.Run("update", func(t *testing.T) {
				t.Parallel()
				diag := NewUpdateErrorWithContext(resource.resourceType, resource.resourceID, testErr)
				if !strings.Contains(diag.Summary(), "Failed to Update") {
					t.Errorf("Update summary missing 'Failed to Update', got: %s", diag.Summary())
				}
			})

			t.Run("delete", func(t *testing.T) {
				t.Parallel()
				diag := NewDeleteErrorWithContext(resource.resourceType, resource.resourceID, testErr)
				if !strings.Contains(diag.Summary(), "Failed to Delete") {
					t.Errorf("Delete summary missing 'Failed to Delete', got: %s", diag.Summary())
				}
			})
		})
	}
}

// ---------------------------------------------------------------------------
// Error message formatting consistency (WithContext versions)
// ---------------------------------------------------------------------------

func TestErrorMessage_Format(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")

	tests := []struct {
		name        string
		createDiag  func() (string, string)
		wantSummary string
	}{
		{
			name: "create error format",
			createDiag: func() (string, string) {
				d := NewCreateErrorWithContext("Monitor", testErr)
				return d.Summary(), d.Detail()
			},
			wantSummary: "Failed to Create Monitor",
		},
		{
			name: "read error format",
			createDiag: func() (string, string) {
				d := NewReadErrorWithContext("Monitor", "mon_123", testErr)
				return d.Summary(), d.Detail()
			},
			wantSummary: "Failed to Read Monitor",
		},
		{
			name: "update error format",
			createDiag: func() (string, string) {
				d := NewUpdateErrorWithContext("Monitor", "mon_123", testErr)
				return d.Summary(), d.Detail()
			},
			wantSummary: "Failed to Update Monitor",
		},
		{
			name: "delete error format",
			createDiag: func() (string, string) {
				d := NewDeleteErrorWithContext("Monitor", "mon_123", testErr)
				return d.Summary(), d.Detail()
			},
			wantSummary: "Failed to Delete Monitor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			summary, detail := tt.createDiag()

			if summary != tt.wantSummary {
				t.Errorf("Summary = %q, want %q", summary, tt.wantSummary)
			}
			if len(detail) == 0 {
				t.Error("Detail is empty")
			}
			if !strings.Contains(detail, "Troubleshooting") {
				t.Error("Detail missing Troubleshooting section")
			}
		})
	}
}

func TestErrorMessage_AllOperations(t *testing.T) {
	t.Parallel()

	testErr := hyperping.NewAPIError(400, "test error")

	operations := map[string]func() string{
		"Create": func() string { return NewCreateErrorWithContext("Resource", testErr).Detail() },
		"Read":   func() string { return NewReadErrorWithContext("Resource", "id_123", testErr).Detail() },
		"Update": func() string { return NewUpdateErrorWithContext("Resource", "id_123", testErr).Detail() },
		"Delete": func() string { return NewDeleteErrorWithContext("Resource", "id_123", testErr).Detail() },
	}

	for opName, opFunc := range operations {
		t.Run(opName, func(t *testing.T) {
			t.Parallel()
			detail := opFunc()

			if !strings.Contains(detail, "Troubleshooting") {
				t.Errorf("%s operation missing troubleshooting guidance", opName)
			}
			if !strings.Contains(detail, "test error") {
				t.Errorf("%s operation missing original error message", opName)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// User readability / actionability
// ---------------------------------------------------------------------------

func TestTroubleshootingSteps_UserReadability(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		diag    func() (string, string)
		wantHas []string
		wantNot []string
	}{
		{
			name: "read error is user-friendly",
			diag: func() (string, string) {
				err := hyperping.NewAPIError(404, "not found")
				d := NewReadErrorWithContext("Monitor", "mon_123", err)
				return d.Summary(), d.Detail()
			},
			wantHas: []string{"Troubleshooting", "Verify"},
			wantNot: []string{"panic", "nil pointer", "stack trace"},
		},
		{
			name: "create error provides actionable steps",
			diag: func() (string, string) {
				err := hyperping.NewAPIError(400, "invalid request")
				d := NewCreateErrorWithContext("Monitor", err)
				return d.Summary(), d.Detail()
			},
			wantHas: []string{"Troubleshooting"},
			wantNot: []string{"panic", "nil pointer"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, detail := tt.diag()

			for _, want := range tt.wantHas {
				if !strings.Contains(detail, want) {
					t.Errorf("Detail missing %q, got: %s", want, detail)
				}
			}
			for _, notWant := range tt.wantNot {
				if strings.Contains(strings.ToLower(detail), strings.ToLower(notWant)) {
					t.Errorf("Detail should NOT contain %q, got: %s", notWant, detail)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Rate limit retry guidance across the error flow
// ---------------------------------------------------------------------------

func TestRateLimitError_RetryGuidance(t *testing.T) {
	t.Parallel()

	retryAfterTests := []int{30, 60, 120, 300}

	for _, retryAfter := range retryAfterTests {
		t.Run(strings.Join([]string{"retry_after", string(rune(retryAfter))}, "_"), func(t *testing.T) {
			t.Parallel()
			err := hyperping.NewRateLimitError(retryAfter)
			diag := NewReadErrorWithContext("Monitor", "mon_rate123", err)

			detail := diag.Detail()
			if !strings.Contains(detail, "Troubleshooting") {
				t.Error("Missing troubleshooting guidance for rate limit")
			}
			if !strings.Contains(err.Error(), "retry after") {
				t.Error("Error message missing retry time")
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Auth error key verification across the error flow
// ---------------------------------------------------------------------------

func TestAuthError_KeyVerification(t *testing.T) {
	t.Parallel()

	authErrors := []int{401, 403}

	for _, statusCode := range authErrors {
		t.Run(strings.Join([]string{"status", string(rune(statusCode))}, "_"), func(t *testing.T) {
			t.Parallel()
			err := hyperping.NewAPIError(statusCode, "unauthorized")
			diag := NewReadErrorWithContext("Monitor", "mon_auth123", err)

			detail := diag.Detail()
			if !strings.Contains(detail, "HYPERPING_API_KEY") {
				t.Error("Detail missing API key verification")
			}
			if !strings.Contains(detail, "Troubleshooting") {
				t.Error("Detail missing Troubleshooting section")
			}
		})
	}
}
