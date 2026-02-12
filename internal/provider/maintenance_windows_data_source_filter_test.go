// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestMaintenanceWindowsDataSource_shouldIncludeMaintenance(t *testing.T) {
	ds := &MaintenanceWindowsDataSource{}

	tests := []struct {
		name        string
		maintenance client.Maintenance
		filter      *MaintenanceFilterModel
		expected    bool
		hasError    bool
	}{
		{
			name: "empty filter - includes all",
			maintenance: client.Maintenance{
				Name:   "maintenance-1",
				Title:  client.LocalizedText{En: "Database Maintenance"},
				Status: "upcoming",
			},
			filter: &MaintenanceFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex match on title",
			maintenance: client.Maintenance{
				Name:   "maint-001",
				Title:  client.LocalizedText{En: "Database Maintenance Window"},
				Status: "upcoming",
			},
			filter: &MaintenanceFilterModel{
				NameRegex: types.StringValue("Database.*"),
				Status:    types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex match on name",
			maintenance: client.Maintenance{
				Name:   "db-maintenance-001",
				Title:  client.LocalizedText{En: "Routine Maintenance"},
				Status: "upcoming",
			},
			filter: &MaintenanceFilterModel{
				NameRegex: types.StringValue("db-.*"),
				Status:    types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex no match",
			maintenance: client.Maintenance{
				Name:   "api-update",
				Title:  client.LocalizedText{En: "API Update"},
				Status: "upcoming",
			},
			filter: &MaintenanceFilterModel{
				NameRegex: types.StringValue("Database.*"),
				Status:    types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "status filter match",
			maintenance: client.Maintenance{
				Name:   "maint-001",
				Title:  client.LocalizedText{En: "Test Maintenance"},
				Status: "ongoing",
			},
			filter: &MaintenanceFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringValue("ongoing"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "status filter no match",
			maintenance: client.Maintenance{
				Name:   "maint-001",
				Title:  client.LocalizedText{En: "Test Maintenance"},
				Status: "upcoming",
			},
			filter: &MaintenanceFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringValue("ongoing"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "combined filters - all match",
			maintenance: client.Maintenance{
				Name:   "db-maint-001",
				Title:  client.LocalizedText{En: "Database Maintenance"},
				Status: "ongoing",
			},
			filter: &MaintenanceFilterModel{
				NameRegex: types.StringValue("Database.*"),
				Status:    types.StringValue("ongoing"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "combined filters - name matches but status doesn't",
			maintenance: client.Maintenance{
				Name:   "db-maint-001",
				Title:  client.LocalizedText{En: "Database Maintenance"},
				Status: "completed",
			},
			filter: &MaintenanceFilterModel{
				NameRegex: types.StringValue("Database.*"),
				Status:    types.StringValue("ongoing"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "invalid regex",
			maintenance: client.Maintenance{
				Name:   "test",
				Title:  client.LocalizedText{En: "Test"},
				Status: "upcoming",
			},
			filter: &MaintenanceFilterModel{
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
			result := ds.shouldIncludeMaintenance(&tt.maintenance, tt.filter, &diags)

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
