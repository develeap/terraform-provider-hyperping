// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type maintenanceDateCase struct {
	name          string
	startDate     string
	endDate       string
	wantError     bool
	wantWarning   bool
	errorContains string
	warningType   string // "past_start" or "long_duration"
}

func buildMaintenanceDateCases() []maintenanceDateCase {
	return []maintenanceDateCase{
		{
			name:        "valid future dates 2h duration",
			startDate:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			endDate:     time.Now().Add(26 * time.Hour).Format(time.RFC3339),
			wantError:   false,
			wantWarning: false,
		},
		{
			name:          "end before start",
			startDate:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			endDate:       time.Now().Add(22 * time.Hour).Format(time.RFC3339),
			wantError:     true,
			errorContains: "end_date must be after start_date",
		},
		{
			name:          "end equal to start",
			startDate:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			endDate:       time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			wantError:     true,
			errorContains: "end_date must be after start_date",
		},
		{
			name:          "start date in past",
			startDate:     time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
			endDate:       time.Now().Add(2 * time.Hour).Format(time.RFC3339),
			wantError:     false,
			wantWarning:   true,
			warningType:   "past_start",
			errorContains: "start_date is in the past",
		},
		{
			name:          "duration > 7 days",
			startDate:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			endDate:       time.Now().Add(24*time.Hour + 8*24*time.Hour).Format(time.RFC3339),
			wantError:     false,
			wantWarning:   true,
			warningType:   "long_duration",
			errorContains: "Maintenance window duration",
		},
		{
			name:        "duration exactly 7 days",
			startDate:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			endDate:     time.Now().Add(24*time.Hour + 7*24*time.Hour).Format(time.RFC3339),
			wantError:   false,
			wantWarning: false,
		},
		{
			name:          "invalid start date format",
			startDate:     "2026-01-29 10:00:00",
			endDate:       time.Now().Add(26 * time.Hour).Format(time.RFC3339),
			wantError:     true,
			errorContains: "Could not parse start_date",
		},
		{
			name:          "invalid end date format",
			startDate:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			endDate:       "invalid-date",
			wantError:     true,
			errorContains: "Could not parse end_date",
		},
		{
			name:        "null start date",
			startDate:   "",
			endDate:     time.Now().Add(26 * time.Hour).Format(time.RFC3339),
			wantError:   false,
			wantWarning: false,
		},
		{
			name:        "null end date",
			startDate:   time.Now().Add(24 * time.Hour).Format(time.RFC3339),
			endDate:     "",
			wantError:   false,
			wantWarning: false,
		},
	}
}

func TestValidateMaintenanceDates(t *testing.T) {
	t.Parallel()

	for _, tt := range buildMaintenanceDateCases() {
		t.Run(tt.name, func(t *testing.T) {
			model := buildMaintenanceModel(tt.startDate, tt.endDate)
			diags := validateMaintenanceDates(model)

			assertMaintenanceErrors(t, diags, tt.wantError, tt.errorContains)
			assertMaintenanceWarnings(t, diags, tt.wantWarning, tt.errorContains)
		})
	}
}

// buildMaintenanceModel constructs a MaintenanceResourceModel from string dates.
func buildMaintenanceModel(startDate, endDate string) *MaintenanceResourceModel {
	model := &MaintenanceResourceModel{}
	if startDate != "" {
		model.StartDate = types.StringValue(startDate)
	} else {
		model.StartDate = types.StringNull()
	}
	if endDate != "" {
		model.EndDate = types.StringValue(endDate)
	} else {
		model.EndDate = types.StringNull()
	}
	return model
}

// assertMaintenanceErrors checks error presence and optional substring match.
func assertMaintenanceErrors(t *testing.T, diags diag.Diagnostics, wantError bool, errorContains string) {
	t.Helper()
	hasError := diags.HasError()
	if hasError != wantError {
		t.Errorf("got error=%v, want error=%v", hasError, wantError)
		for _, d := range diags.Errors() {
			t.Logf("  Error: %s - %s", d.Summary(), d.Detail())
		}
	}
	if !wantError || errorContains == "" {
		return
	}
	assertDiagContains(t, diags.Errors(), errorContains, "error")
}

// assertMaintenanceWarnings checks warning presence and optional substring match.
func assertMaintenanceWarnings(t *testing.T, diags diag.Diagnostics, wantWarning bool, errorContains string) {
	t.Helper()
	hasWarning := len(diags.Warnings()) > 0
	if hasWarning != wantWarning {
		t.Errorf("got warning=%v, want warning=%v", hasWarning, wantWarning)
		for _, d := range diags.Warnings() {
			t.Logf("  Warning: %s - %s", d.Summary(), d.Detail())
		}
	}
	if !wantWarning || errorContains == "" {
		return
	}
	assertDiagContains(t, diags.Warnings(), errorContains, "warning")
}

// assertDiagContains checks that at least one diagnostic contains the expected substring.
func assertDiagContains(t *testing.T, entries []diag.Diagnostic, substr, kind string) {
	t.Helper()
	for _, d := range entries {
		if containsString(d.Detail(), substr) || containsString(d.Summary(), substr) {
			return
		}
	}
	t.Errorf("expected %s containing %q, got: %v", kind, substr, entries)
}

func TestValidateMaintenanceDates_BothNull(t *testing.T) {
	model := &MaintenanceResourceModel{
		StartDate: types.StringNull(),
		EndDate:   types.StringNull(),
	}

	diags := validateMaintenanceDates(model)

	if diags.HasError() {
		t.Errorf("validateMaintenanceDates() with null dates: unexpected error: %v", diags.Errors())
	}

	if len(diags.Warnings()) > 0 {
		t.Errorf("validateMaintenanceDates() with null dates: unexpected warning: %v", diags.Warnings())
	}
}

func TestValidateMaintenanceDates_PastStartAndLongDuration(t *testing.T) {
	model := &MaintenanceResourceModel{
		StartDate: types.StringValue(time.Now().Add(-2 * time.Hour).Format(time.RFC3339)),
		EndDate:   types.StringValue(time.Now().Add(8*24*time.Hour - 2*time.Hour).Format(time.RFC3339)),
	}

	diags := validateMaintenanceDates(model)

	if diags.HasError() {
		t.Errorf("validateMaintenanceDates(): unexpected error: %v", diags.Errors())
	}

	if len(diags.Warnings()) != 2 {
		t.Errorf("validateMaintenanceDates(): expected 2 warnings, got %d", len(diags.Warnings()))
		for _, d := range diags.Warnings() {
			t.Logf("  Warning: %s - %s", d.Summary(), d.Detail())
		}
	}
}

// containsString checks if a string contains a substring (case-insensitive helper)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && len(substr) > 0 && contains(s, substr))
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
