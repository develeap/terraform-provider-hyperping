// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestStatusPagesDataSource_shouldIncludeStatusPage(t *testing.T) {
	ds := &StatusPagesDataSource{}

	customHostname := "status.example.com"
	noHostname := (*string)(nil)

	tests := []struct {
		name       string
		statusPage client.StatusPage
		filter     *StatusPageFilterModel
		expected   bool
		hasError   bool
	}{
		{
			name: "empty filter - includes all",
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
			statusPage: client.StatusPage{
				Settings: client.StatusPageSettings{
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
