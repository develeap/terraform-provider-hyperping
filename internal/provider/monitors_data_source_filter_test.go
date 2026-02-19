// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestMonitorsDataSource_shouldIncludeMonitor(t *testing.T) {
	ds := &MonitorsDataSource{}

	tests := []struct {
		name     string
		monitor  client.Monitor
		filter   *MonitorFilterModel
		expected bool
		hasError bool
	}{
		{
			name: "empty filter - includes all",
			monitor: client.Monitor{
				Name:     "Test Monitor",
				Protocol: "http",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex match",
			monitor: client.Monitor{
				Name:     "[PROD]-API-Health",
				Protocol: "https",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringValue("\\[PROD\\]-.*"),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex no match",
			monitor: client.Monitor{
				Name:     "[DEV]-API-Health",
				Protocol: "https",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringValue("\\[PROD\\]-.*"),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "protocol exact match",
			monitor: client.Monitor{
				Name:     "Test Monitor",
				Protocol: "https",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringValue("https"),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "protocol no match",
			monitor: client.Monitor{
				Name:     "Test Monitor",
				Protocol: "http",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringValue("https"),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "paused filter match",
			monitor: client.Monitor{
				Name:     "Test Monitor",
				Protocol: "http",
				Paused:   true,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolValue(true),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "paused filter no match",
			monitor: client.Monitor{
				Name:     "Test Monitor",
				Protocol: "http",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolValue(true),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "combined filters - all match",
			monitor: client.Monitor{
				Name:     "[PROD]-API-Monitor",
				Protocol: "https",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringValue("\\[PROD\\]-.*"),
				Protocol:    types.StringValue("https"),
				Paused:      types.BoolValue(false),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "combined filters - name matches but protocol doesn't",
			monitor: client.Monitor{
				Name:     "[PROD]-API-Monitor",
				Protocol: "http",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringValue("\\[PROD\\]-.*"),
				Protocol:    types.StringValue("https"),
				Paused:      types.BoolValue(false),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "invalid regex",
			monitor: client.Monitor{
				Name:     "Test Monitor",
				Protocol: "http",
				Paused:   false,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringValue("[invalid"),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: false,
			hasError: true,
		},
		{
			name: "empty name regex matches all",
			monitor: client.Monitor{
				Name:     "Any Monitor Name",
				Protocol: "tcp",
				Paused:   true,
			},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name:    "status filter up matches up monitor",
			monitor: client.Monitor{Status: "up"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringValue("up"),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name:    "status filter up excludes down monitor",
			monitor: client.Monitor{Status: "down"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringValue("up"),
				ProjectUUID: types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name:    "status filter down matches down monitor",
			monitor: client.Monitor{Status: "down"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringValue("down"),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name:    "project_uuid filter exact match",
			monitor: client.Monitor{ProjectUUID: "proj_abc123"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringValue("proj_abc123"),
			},
			expected: true,
			hasError: false,
		},
		{
			name:    "project_uuid filter no match",
			monitor: client.Monitor{ProjectUUID: "proj_other"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringValue("proj_abc123"),
			},
			expected: false,
			hasError: false,
		},
		{
			name:    "combined status and project_uuid both match",
			monitor: client.Monitor{Status: "up", ProjectUUID: "proj_abc123"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringValue("up"),
				ProjectUUID: types.StringValue("proj_abc123"),
			},
			expected: true,
			hasError: false,
		},
		{
			name:    "combined status matches but project_uuid does not",
			monitor: client.Monitor{Status: "up", ProjectUUID: "proj_other"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringValue("up"),
				ProjectUUID: types.StringValue("proj_abc123"),
			},
			expected: false,
			hasError: false,
		},
		{
			name:    "nil status filter passes through",
			monitor: client.Monitor{Status: "down"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name:    "nil project_uuid filter passes through",
			monitor: client.Monitor{ProjectUUID: "proj_any"},
			filter: &MonitorFilterModel{
				NameRegex:   types.StringNull(),
				Protocol:    types.StringNull(),
				Paused:      types.BoolNull(),
				Status:      types.StringNull(),
				ProjectUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := ds.shouldIncludeMonitor(&tt.monitor, tt.filter, &diags)

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
