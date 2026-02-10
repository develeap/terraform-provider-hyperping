// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"
	"strings"
	"testing"
)

// Comprehensive tests for error helper coverage
func TestNewReadAfterUpdateError_Comprehensive(t *testing.T) {
	tests := []struct {
		name          string
		resourceType  string
		resourceID    string
		err           error
		wantInSummary string
		wantInDetail  []string
	}{
		{
			name:          "incident update error",
			resourceType:  "Incident",
			resourceID:    "inc_789",
			err:           errors.New("connection timeout"),
			wantInSummary: "Updated But Read Failed",
			wantInDetail:  []string{"inc_789", "connection timeout", "terraform refresh"},
		},
		{
			name:          "monitor update error",
			resourceType:  "Monitor",
			resourceID:    "mon_abc",
			err:           errors.New("network error"),
			wantInSummary: "Updated But Read Failed",
			wantInDetail:  []string{"mon_abc", "network error", "terraform refresh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diag := newReadAfterUpdateError(tt.resourceType, tt.resourceID, tt.err)

			summary := diag.Summary()
			if !strings.Contains(summary, tt.wantInSummary) {
				t.Errorf("Summary missing %q, got: %s", tt.wantInSummary, summary)
			}

			detail := diag.Detail()
			for _, want := range tt.wantInDetail {
				if !strings.Contains(detail, want) {
					t.Errorf("Detail missing %q, got: %s", want, detail)
				}
			}
		})
	}
}

// Test all error helpers return non-empty diagnostics
func TestAllErrorHelpersReturnValidDiagnostics(t *testing.T) {
	testErr := errors.New("test error")

	tests := []struct {
		name   string
		create func() (string, string)
	}{
		{
			name: "newCreateError",
			create: func() (string, string) {
				d := newCreateError("Resource", testErr)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newReadError",
			create: func() (string, string) {
				d := newReadError("Resource", "id_123", testErr)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newUpdateError",
			create: func() (string, string) {
				d := newUpdateError("Resource", "id_456", testErr)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newDeleteError",
			create: func() (string, string) {
				d := newDeleteError("Resource", "id_789", testErr)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newListError",
			create: func() (string, string) {
				d := newListError("Resources", testErr)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newReadAfterCreateError",
			create: func() (string, string) {
				d := newReadAfterCreateError("Resource", "new_123", testErr)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newReadAfterUpdateError",
			create: func() (string, string) {
				d := newReadAfterUpdateError("Resource", "upd_456", testErr)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newConfigError",
			create: func() (string, string) {
				d := newConfigError("Invalid configuration")
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newValidationError",
			create: func() (string, string) {
				d := newValidationError("Field", "Invalid value")
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newImportError",
			create: func() (string, string) {
				d := newImportError("Resource", testErr)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newUnexpectedConfigTypeError",
			create: func() (string, string) {
				d := newUnexpectedConfigTypeError("*client.Client", 42)
				return d.Summary(), d.Detail()
			},
		},
		{
			name: "newDeleteWarning",
			create: func() (string, string) {
				d := newDeleteWarning("Resource", "Already deleted")
				return d.Summary(), d.Detail()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

// Test deprecated functions for backward compatibility
func TestDeprecatedFunctions(t *testing.T) {
	t.Run("formatResourceError with ID", func(t *testing.T) {
		summary, detail := formatResourceError("Monitor", "Creating", "mon_123", errors.New("test"))

		if !strings.Contains(summary, "Error Creating Monitor") {
			t.Errorf("unexpected summary: %s", summary)
		}
		if !strings.Contains(detail, "mon_123") {
			t.Errorf("missing resource ID in detail: %s", detail)
		}
		if !strings.Contains(detail, "test") {
			t.Errorf("missing error in detail: %s", detail)
		}
	})

	t.Run("formatResourceError without ID", func(t *testing.T) {
		summary, detail := formatResourceError("StatusPage", "Reading", "", errors.New("not found"))

		if !strings.Contains(summary, "Error Reading StatusPage") {
			t.Errorf("unexpected summary: %s", summary)
		}
		if !strings.Contains(detail, "not found") {
			t.Errorf("missing error in detail: %s", detail)
		}
	})

	t.Run("formatAPIError", func(t *testing.T) {
		summary, detail := formatAPIError("Monitor Creation", errors.New("connection refused"))

		if !strings.Contains(summary, "API Error During Monitor Creation") {
			t.Errorf("unexpected summary: %s", summary)
		}
		if !strings.Contains(detail, "connection refused") {
			t.Errorf("missing error in detail: %s", detail)
		}
		if !strings.Contains(detail, "Invalid API key") {
			t.Errorf("missing troubleshooting hint in detail: %s", detail)
		}
	})
}
