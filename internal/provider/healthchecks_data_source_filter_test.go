// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestHealthchecksDataSource_shouldIncludeHealthcheck(t *testing.T) {
	ds := &HealthchecksDataSource{}

	tests := []struct {
		name        string
		healthcheck client.Healthcheck
		filter      *HealthcheckFilterModel
		expected    bool
		hasError    bool
	}{
		{
			name: "empty filter - includes all",
			healthcheck: client.Healthcheck{
				Name:   "Daily Backup Check",
				IsDown: false,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex match",
			healthcheck: client.Healthcheck{
				Name:   "Backup-DB-Daily",
				IsDown: false,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringValue("Backup-.*"),
				Status:    types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex no match",
			healthcheck: client.Healthcheck{
				Name:   "Cron-Job-Monitor",
				IsDown: false,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringValue("Backup-.*"),
				Status:    types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "status filter down - match",
			healthcheck: client.Healthcheck{
				Name:   "Test Healthcheck",
				IsDown: true,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringValue("down"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "status filter down - no match",
			healthcheck: client.Healthcheck{
				Name:   "Test Healthcheck",
				IsDown: false,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringValue("down"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "status filter up - match",
			healthcheck: client.Healthcheck{
				Name:   "Test Healthcheck",
				IsDown: false,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringValue("up"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "status filter up - no match",
			healthcheck: client.Healthcheck{
				Name:   "Test Healthcheck",
				IsDown: true,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringValue("up"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "combined filters - all match",
			healthcheck: client.Healthcheck{
				Name:   "Backup-Daily-Cron",
				IsDown: false,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringValue("Backup-.*"),
				Status:    types.StringValue("up"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "combined filters - name matches but status doesn't",
			healthcheck: client.Healthcheck{
				Name:   "Backup-Daily-Cron",
				IsDown: true,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringValue("Backup-.*"),
				Status:    types.StringValue("up"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "invalid regex",
			healthcheck: client.Healthcheck{
				Name:   "Test Healthcheck",
				IsDown: false,
			},
			filter: &HealthcheckFilterModel{
				NameRegex: types.StringValue("[invalid"),
				Status:    types.StringNull(),
			},
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := ds.shouldIncludeHealthcheck(&tt.healthcheck, tt.filter, &diags)

			if tt.hasError {
				if !diags.HasError() {
					t.Errorf("expected error but got none")
				}
			} else {
				if diags.HasError() {
					t.Errorf("unexpected error: %v", diags)
				}
			}

			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}
