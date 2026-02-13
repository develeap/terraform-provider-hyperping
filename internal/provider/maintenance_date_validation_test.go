// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateMaintenanceDates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		startDate     string
		endDate       string
		wantError     bool
		wantWarning   bool
		errorContains string
		warningType   string // "past_start" or "long_duration"
	}{
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
			endDate:       time.Now().Add(24*time.Hour + 8*24*time.Hour).Format(time.RFC3339), // 8 days
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
			wantWarning: false, // Exactly 7 days should NOT warn
		},
		{
			name:          "invalid start date format",
			startDate:     "2026-01-29 10:00:00", // Space instead of T
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &MaintenanceResourceModel{}

			if tt.startDate != "" {
				model.StartDate = types.StringValue(tt.startDate)
			} else {
				model.StartDate = types.StringNull()
			}

			if tt.endDate != "" {
				model.EndDate = types.StringValue(tt.endDate)
			} else {
				model.EndDate = types.StringNull()
			}

			diags := validateMaintenanceDates(model)

			hasError := diags.HasError()
			if hasError != tt.wantError {
				t.Errorf("validateMaintenanceDates(): got error=%v, want error=%v", hasError, tt.wantError)
				for _, d := range diags.Errors() {
					t.Logf("  Error: %s - %s", d.Summary(), d.Detail())
				}
			}

			if tt.wantError && tt.errorContains != "" {
				found := false
				for _, d := range diags.Errors() {
					if containsString(d.Detail(), tt.errorContains) || containsString(d.Summary(), tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateMaintenanceDates(): expected error containing %q, got errors: %v", tt.errorContains, diags.Errors())
				}
			}

			hasWarning := len(diags.Warnings()) > 0
			if hasWarning != tt.wantWarning {
				t.Errorf("validateMaintenanceDates(): got warning=%v, want warning=%v", hasWarning, tt.wantWarning)
				for _, d := range diags.Warnings() {
					t.Logf("  Warning: %s - %s", d.Summary(), d.Detail())
				}
			}

			if tt.wantWarning && tt.errorContains != "" {
				found := false
				for _, d := range diags.Warnings() {
					if containsString(d.Detail(), tt.errorContains) || containsString(d.Summary(), tt.errorContains) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateMaintenanceDates(): expected warning containing %q, got warnings: %v", tt.errorContains, diags.Warnings())
				}
			}
		})
	}
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
	// Test case where both warnings should trigger
	model := &MaintenanceResourceModel{
		StartDate: types.StringValue(time.Now().Add(-2 * time.Hour).Format(time.RFC3339)),
		EndDate:   types.StringValue(time.Now().Add(8*24*time.Hour - 2*time.Hour).Format(time.RFC3339)), // ~8 days from start
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
