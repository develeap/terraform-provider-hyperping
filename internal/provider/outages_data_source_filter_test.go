// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestOutagesDataSource_shouldIncludeOutage(t *testing.T) {
	ds := &OutagesDataSource{}

	tests := []struct {
		name     string
		outage   client.Outage
		filter   *OutageFilterModel
		expected bool
		hasError bool
	}{
		{
			name: "empty filter - includes all",
			outage: client.Outage{
				Monitor: client.MonitorReference{
					UUID: "mon-001",
					Name: "API Monitor",
				},
			},
			filter: &OutageFilterModel{
				NameRegex:   types.StringNull(),
				MonitorUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "monitor name regex match",
			outage: client.Outage{
				Monitor: client.MonitorReference{
					UUID: "mon-001",
					Name: "[PROD]-API-Health",
				},
			},
			filter: &OutageFilterModel{
				NameRegex:   types.StringValue("\\[PROD\\]-.*"),
				MonitorUUID: types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "monitor name regex no match",
			outage: client.Outage{
				Monitor: client.MonitorReference{
					UUID: "mon-001",
					Name: "[DEV]-API-Health",
				},
			},
			filter: &OutageFilterModel{
				NameRegex:   types.StringValue("\\[PROD\\]-.*"),
				MonitorUUID: types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "monitor UUID filter match",
			outage: client.Outage{
				Monitor: client.MonitorReference{
					UUID: "mon-abc-123",
					Name: "Test Monitor",
				},
			},
			filter: &OutageFilterModel{
				NameRegex:   types.StringNull(),
				MonitorUUID: types.StringValue("mon-abc-123"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "monitor UUID filter no match",
			outage: client.Outage{
				Monitor: client.MonitorReference{
					UUID: "mon-xyz-456",
					Name: "Test Monitor",
				},
			},
			filter: &OutageFilterModel{
				NameRegex:   types.StringNull(),
				MonitorUUID: types.StringValue("mon-abc-123"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "combined filters - all match",
			outage: client.Outage{
				Monitor: client.MonitorReference{
					UUID: "mon-prod-001",
					Name: "[PROD]-Database-Monitor",
				},
			},
			filter: &OutageFilterModel{
				NameRegex:   types.StringValue("\\[PROD\\]-.*"),
				MonitorUUID: types.StringValue("mon-prod-001"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "combined filters - name matches but UUID doesn't",
			outage: client.Outage{
				Monitor: client.MonitorReference{
					UUID: "mon-prod-002",
					Name: "[PROD]-Database-Monitor",
				},
			},
			filter: &OutageFilterModel{
				NameRegex:   types.StringValue("\\[PROD\\]-.*"),
				MonitorUUID: types.StringValue("mon-prod-001"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "invalid regex",
			outage: client.Outage{
				Monitor: client.MonitorReference{
					UUID: "mon-001",
					Name: "Test Monitor",
				},
			},
			filter: &OutageFilterModel{
				NameRegex:   types.StringValue("[invalid"),
				MonitorUUID: types.StringNull(),
			},
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := ds.shouldIncludeOutage(&tt.outage, tt.filter, &diags)

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
