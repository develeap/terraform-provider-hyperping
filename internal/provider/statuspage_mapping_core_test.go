// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// TestMapStringMapToTF tests conversion of map[string]string to types.Map
func TestMapStringMapToTF(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]string
		wantNull bool
	}{
		{
			name:     "nil map returns null",
			input:    nil,
			wantNull: true,
		},
		{
			name:     "empty map returns null",
			input:    map[string]string{},
			wantNull: true,
		},
		{
			name: "single value",
			input: map[string]string{
				"en": "English Name",
			},
			wantNull: false,
		},
		{
			name: "multiple languages",
			input: map[string]string{
				"en": "English Name",
				"fr": "Nom Français",
				"de": "Deutscher Name",
			},
			wantNull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapStringMapToTF(tt.input)

			if tt.wantNull {
				if !result.IsNull() {
					t.Errorf("expected null map, got %v", result)
				}
				return
			}

			if result.IsNull() {
				t.Errorf("expected non-null map")
				return
			}

			elements := result.Elements()
			if len(elements) != len(tt.input) {
				t.Errorf("expected %d elements, got %d", len(tt.input), len(elements))
			}

			for k, v := range tt.input {
				attrVal, ok := elements[k]
				if !ok {
					t.Errorf("expected key %q not found", k)
					continue
				}

				strVal, ok := attrVal.(types.String)
				if !ok {
					t.Errorf("expected types.String for key %q, got %T", k, attrVal)
					continue
				}

				if strVal.ValueString() != v {
					t.Errorf("key %q: expected %q, got %q", k, v, strVal.ValueString())
				}
			}
		})
	}
}

// TestMapTFToStringMap tests conversion of types.Map to map[string]string
func TestMapTFToStringMap(t *testing.T) {
	tests := []struct {
		name      string
		input     types.Map
		want      map[string]string
		wantError bool
	}{
		{
			name:      "null map returns empty",
			input:     types.MapNull(types.StringType),
			want:      map[string]string{},
			wantError: false,
		},
		{
			name:      "unknown map returns empty",
			input:     types.MapUnknown(types.StringType),
			want:      map[string]string{},
			wantError: false,
		},
		{
			name: "single value",
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"en": types.StringValue("English Name"),
			}),
			want: map[string]string{
				"en": "English Name",
			},
			wantError: false,
		},
		{
			name: "multiple values",
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"en": types.StringValue("English"),
				"fr": types.StringValue("Français"),
				"de": types.StringValue("Deutsch"),
			}),
			want: map[string]string{
				"en": "English",
				"fr": "Français",
				"de": "Deutsch",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapTFToStringMap(tt.input, &diags)

			if tt.wantError {
				if !diags.HasError() {
					t.Errorf("expected error, got none")
				}
				return
			}

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if len(result) != len(tt.want) {
				t.Errorf("expected %d elements, got %d", len(tt.want), len(result))
			}

			for k, v := range tt.want {
				got, ok := result[k]
				if !ok {
					t.Errorf("expected key %q not found", k)
					continue
				}
				if got != v {
					t.Errorf("key %q: expected %q, got %q", k, v, got)
				}
			}
		})
	}
}

// TestMapStatusPageCommonFields tests the main mapping function
func TestMapStatusPageCommonFields(t *testing.T) {
	tests := []struct {
		name     string
		input    *client.StatusPage
		wantNull bool
	}{
		{
			name:     "nil status page returns null fields",
			input:    nil,
			wantNull: true,
		},
		{
			name: "minimal status page",
			input: &client.StatusPage{
				UUID:            "sp_abc123",
				Name:            "Production Status",
				HostedSubdomain: "status",
				URL:             "https://status.hyperping.app",
				Settings: client.StatusPageSettings{
					Theme:       "dark",
					Font:        "Inter",
					AccentColor: "#36b27e",
					Languages:   []string{"en"},
				},
				Sections: []client.StatusPageSection{},
			},
			wantNull: false,
		},
		{
			name: "full status page with hostname",
			input: &client.StatusPage{
				UUID:            "sp_full123",
				Name:            "Full Status",
				Hostname:        stringPtr("status.example.com"),
				HostedSubdomain: "status",
				URL:             "https://status.example.com",
				Settings: client.StatusPageSettings{
					Theme:       "light",
					Font:        "Roboto",
					AccentColor: "#0066cc",
					Languages:   []string{"en", "fr"},
				},
				Sections: []client.StatusPageSection{
					{
						Name: map[string]string{
							"en": "API Services",
							"fr": "Services API",
						},
						IsSplit: true,
						Services: []client.StatusPageService{
							{
								ID:                "svc_1",
								UUID:              "mon_abc123",
								Name:              map[string]string{"en": "Main API"},
								ShowUptime:        true,
								ShowResponseTimes: true,
							},
						},
					},
				},
			},
			wantNull: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := MapStatusPageCommonFields(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if tt.wantNull {
				if !result.ID.IsNull() {
					t.Errorf("expected null ID")
				}
				if !result.Name.IsNull() {
					t.Errorf("expected null Name")
				}
				return
			}

			if result.ID.IsNull() {
				t.Errorf("expected non-null ID")
			}

			if result.ID.ValueString() != tt.input.UUID {
				t.Errorf("ID: expected %q, got %q", tt.input.UUID, result.ID.ValueString())
			}

			if result.Name.ValueString() != tt.input.Name {
				t.Errorf("Name: expected %q, got %q", tt.input.Name, result.Name.ValueString())
			}

			// Check hostname handling
			if tt.input.Hostname != nil {
				if result.Hostname.IsNull() {
					t.Errorf("expected non-null Hostname")
				}
				if result.Hostname.ValueString() != *tt.input.Hostname {
					t.Errorf("Hostname: expected %q, got %q", *tt.input.Hostname, result.Hostname.ValueString())
				}
			} else {
				if !result.Hostname.IsNull() {
					t.Errorf("expected null Hostname")
				}
			}
		})
	}
}

// TestMapListToStringSlice tests conversion of TF list to string slice
func TestMapListToStringSlice(t *testing.T) {
	tests := []struct {
		name      string
		input     types.List
		wantCount int
	}{
		{
			name:      "null list returns empty",
			input:     types.ListNull(types.StringType),
			wantCount: 0,
		},
		{
			name:      "unknown list returns empty",
			input:     types.ListUnknown(types.StringType),
			wantCount: 0,
		},
		{
			name: "single item",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("en"),
			}),
			wantCount: 1,
		},
		{
			name: "multiple items",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("en"),
				types.StringValue("fr"),
				types.StringValue("de"),
			}),
			wantCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapListToStringSlice(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if len(result) != tt.wantCount {
				t.Errorf("expected %d items, got %d", tt.wantCount, len(result))
			}
		})
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
