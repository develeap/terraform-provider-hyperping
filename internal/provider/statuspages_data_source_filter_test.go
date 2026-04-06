// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hyperping "github.com/develeap/hyperping-go"
)

func TestStatusPagesDataSource_shouldIncludeStatusPage(t *testing.T) {
	ds := &StatusPagesDataSource{}

	customHostname := "status.example.com"
	noHostname := (*string)(nil)

	tests := []struct {
		name       string
		statusPage hyperping.StatusPage
		filter     *StatusPageFilterModel
		expected   bool
		hasError   bool
	}{
		{
			name: "empty filter - includes all",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Production Status",
				},
				Hostname: &customHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringNull(),
				Hostname:  types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex match",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Production API Status",
				},
				Hostname: noHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringValue("Production.*"),
				Hostname:  types.StringNull(),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "name regex no match",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Development Status",
				},
				Hostname: noHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringValue("Production.*"),
				Hostname:  types.StringNull(),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "hostname filter match",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Test Status",
				},
				Hostname: &customHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringNull(),
				Hostname:  types.StringValue("status.example.com"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "hostname filter no match",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Test Status",
				},
				Hostname: &customHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringNull(),
				Hostname:  types.StringValue("other.example.com"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "hostname filter with null hostname",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Test Status",
				},
				Hostname: noHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringNull(),
				Hostname:  types.StringValue("status.example.com"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "combined filters - all match",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Production Status Page",
				},
				Hostname: &customHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringValue("Production.*"),
				Hostname:  types.StringValue("status.example.com"),
			},
			expected: true,
			hasError: false,
		},
		{
			name: "combined filters - name matches but hostname doesn't",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Production Status Page",
				},
				Hostname: &customHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringValue("Production.*"),
				Hostname:  types.StringValue("other.example.com"),
			},
			expected: false,
			hasError: false,
		},
		{
			name: "invalid regex",
			statusPage: hyperping.StatusPage{
				Settings: hyperping.StatusPageSettings{
					Name: "Test Status",
				},
				Hostname: noHostname,
			},
			filter: &StatusPageFilterModel{
				NameRegex: types.StringValue("[invalid"),
				Hostname:  types.StringNull(),
			},
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := ds.shouldIncludeStatusPage(&tt.statusPage, tt.filter, &diags)

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
