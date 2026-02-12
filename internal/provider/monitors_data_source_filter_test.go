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
				NameRegex: types.StringNull(),
				Protocol:  types.StringNull(),
				Paused:    types.BoolNull(),
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
				NameRegex: types.StringValue("\\[PROD\\]-.*"),
				Protocol:  types.StringNull(),
				Paused:    types.BoolNull(),
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
				NameRegex: types.StringValue("\\[PROD\\]-.*"),
				Protocol:  types.StringNull(),
				Paused:    types.BoolNull(),
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
				NameRegex: types.StringNull(),
				Protocol:  types.StringValue("https"),
				Paused:    types.BoolNull(),
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
				NameRegex: types.StringNull(),
				Protocol:  types.StringValue("https"),
				Paused:    types.BoolNull(),
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
				NameRegex: types.StringNull(),
				Protocol:  types.StringNull(),
				Paused:    types.BoolValue(true),
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
				NameRegex: types.StringNull(),
				Protocol:  types.StringNull(),
				Paused:    types.BoolValue(true),
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
				NameRegex: types.StringValue("\\[PROD\\]-.*"),
				Protocol:  types.StringValue("https"),
				Paused:    types.BoolValue(false),
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
				NameRegex: types.StringValue("\\[PROD\\]-.*"),
				Protocol:  types.StringValue("https"),
				Paused:    types.BoolValue(false),
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
				NameRegex: types.StringValue("[invalid"),
				Protocol:  types.StringNull(),
				Paused:    types.BoolNull(),
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
				NameRegex: types.StringNull(),
				Protocol:  types.StringNull(),
				Paused:    types.BoolNull(),
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
