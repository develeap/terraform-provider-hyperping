// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestIncidentsDataSource_shouldIncludeIncident(t *testing.T) {
	ds := &IncidentsDataSource{}

	tests := []struct {
		name     string
		incident client.Incident
		filter   *IncidentFilterModel
		expected bool
		hasError bool
	}{
		{
			name: "empty filter - includes all",
			incident: client.Incident{
				Title: client.LocalizedText{En: "Test Incident"},
				Type:  "outage",
			},
			filter: &IncidentFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringNull(),
				Severity:  types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex match",
			incident: client.Incident{
				Title: client.LocalizedText{En: "Payment Gateway Outage"},
				Type:  "outage",
			},
			filter: &IncidentFilterModel{
				NameRegex: types.StringValue("Payment.*"),
				Status:    types.StringNull(),
				Severity:  types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex no match",
			incident: client.Incident{
				Title: client.LocalizedText{En: "API Slow Response"},
				Type:  "incident",
			},
			filter: &IncidentFilterModel{
				NameRegex: types.StringValue("Payment.*"),
				Status:    types.StringNull(),
				Severity:  types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "type filter match",
			incident: client.Incident{
				Title: client.LocalizedText{En: "Test Incident"},
				Type:  "outage",
			},
			filter: &IncidentFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringValue("outage"),
				Severity:  types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "type filter no match",
			incident: client.Incident{
				Title: client.LocalizedText{En: "Test Incident"},
				Type:  "incident",
			},
			filter: &IncidentFilterModel{
				NameRegex: types.StringNull(),
				Status:    types.StringValue("outage"),
				Severity:  types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "combined filters - all match",
			incident: client.Incident{
				Title: client.LocalizedText{En: "Database Outage"},
				Type:  "outage",
			},
			filter: &IncidentFilterModel{
				NameRegex: types.StringValue("Database.*"),
				Status:    types.StringValue("outage"),
				Severity:  types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "combined filters - name matches but type doesn't",
			incident: client.Incident{
				Title: client.LocalizedText{En: "Database Issue"},
				Type:  "incident",
			},
			filter: &IncidentFilterModel{
				NameRegex: types.StringValue("Database.*"),
				Status:    types.StringValue("outage"),
				Severity:  types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "invalid regex",
			incident: client.Incident{
				Title: client.LocalizedText{En: "Test Incident"},
				Type:  "outage",
			},
			filter: &IncidentFilterModel{
				NameRegex: types.StringValue("[invalid"),
				Status:    types.StringNull(),
				Severity:  types.StringNull(),
			},
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := ds.shouldIncludeIncident(&tt.incident, tt.filter, &diags)

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
