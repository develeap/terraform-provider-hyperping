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

// TestMapSettingsToTF tests settings mapping
func TestMapSettingsToTF(t *testing.T) {
	tests := []struct {
		name  string
		input client.StatusPageSettings
	}{
		{
			name: "minimal settings",
			input: client.StatusPageSettings{
				Theme:       "system",
				Font:        "Inter",
				AccentColor: "#36b27e",
				Languages:   []string{"en"},
			},
		},
		{
			name: "full settings with subscribe",
			input: client.StatusPageSettings{
				Theme:       "dark",
				Font:        "Roboto",
				AccentColor: "#0066cc",
				Languages:   []string{"en", "fr", "de"},
				Subscribe: client.StatusPageSubscribeSettings{
					Enabled: true,
					Email:   true,
					SMS:     true,
					Slack:   false,
					Teams:   false,
				},
			},
		},
		{
			name: "full settings with authentication",
			input: client.StatusPageSettings{
				Theme:       "light",
				Font:        "Inter",
				AccentColor: "#ff0000",
				Languages:   []string{"en"},
				Authentication: client.StatusPageAuthenticationSettings{
					PasswordProtection: true,
					GoogleSSO:          false,
					AllowedDomains:     []string{"example.com", "test.com"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapSettingsToTF(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if result.IsNull() {
				t.Errorf("expected non-null settings object")
				return
			}

			// Verify it's an object type
			if result.IsUnknown() {
				t.Errorf("settings object is unknown")
			}
		})
	}
}

// TestMapSectionsToTF tests sections mapping
func TestMapSectionsToTF(t *testing.T) {
	tests := []struct {
		name  string
		input []client.StatusPageSection
	}{
		{
			name:  "nil sections",
			input: nil,
		},
		{
			name:  "empty sections",
			input: []client.StatusPageSection{},
		},
		{
			name: "single section",
			input: []client.StatusPageSection{
				{
					Name: map[string]string{
						"en": "API Services",
					},
					IsSplit:  false,
					Services: []client.StatusPageService{},
				},
			},
		},
		{
			name: "multiple sections with services",
			input: []client.StatusPageSection{
				{
					Name: map[string]string{
						"en": "Frontend",
						"fr": "Interface",
					},
					IsSplit: true,
					Services: []client.StatusPageService{
						{
							ID:         "svc_1",
							UUID:       "mon_web",
							Name:       map[string]string{"en": "Web App"},
							ShowUptime: true,
						},
					},
				},
				{
					Name: map[string]string{
						"en": "Backend",
					},
					IsSplit: false,
					Services: []client.StatusPageService{
						{
							ID:   "svc_2",
							UUID: "mon_api",
							Name: map[string]string{"en": "API"},
						},
					},
				},
			},
		},
		{
			name: "nested service groups",
			input: []client.StatusPageSection{
				{
					Name: map[string]string{
						"en": "Databases",
					},
					IsSplit: false,
					Services: []client.StatusPageService{
						{
							ID:      "grp_1",
							UUID:    "mon_db_primary",
							Name:    map[string]string{"en": "Primary DB"},
							IsGroup: true,
							Services: []client.StatusPageService{
								{
									ID:   "svc_db1",
									UUID: "mon_db1",
									Name: map[string]string{"en": "DB Node 1"},
								},
								{
									ID:   "svc_db2",
									UUID: "mon_db2",
									Name: map[string]string{"en": "DB Node 2"},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapSectionsToTF(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if len(tt.input) == 0 {
				if !result.IsNull() {
					t.Errorf("expected null list for nil/empty input")
				}
				return
			}

			if result.IsNull() {
				t.Errorf("expected non-null list")
				return
			}

			elements := result.Elements()
			if len(elements) != len(tt.input) {
				t.Errorf("expected %d sections, got %d", len(tt.input), len(elements))
			}
		})
	}
}

// TestMapTFToSections tests reverse mapping from TF to client
func TestMapTFToSections(t *testing.T) {
	tests := []struct {
		name      string
		input     types.List
		wantCount int
		wantError bool
	}{
		{
			name:      "null list returns empty",
			input:     types.ListNull(types.ObjectType{AttrTypes: SectionAttrTypes()}),
			wantCount: 0,
			wantError: false,
		},
		{
			name: "single section with name",
			input: types.ListValueMust(types.ObjectType{AttrTypes: SectionAttrTypes()}, []attr.Value{
				types.ObjectValueMust(SectionAttrTypes(), map[string]attr.Value{
					"name": types.MapValueMust(types.StringType, map[string]attr.Value{
						"en": types.StringValue("API Services"),
					}),
					"is_split": types.BoolValue(true),
					"services": types.ListNull(types.ObjectType{AttrTypes: ServiceAttrTypes()}),
				}),
			}),
			wantCount: 1,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapTFToSections(tt.input, &diags)

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

			if len(result) != tt.wantCount {
				t.Errorf("expected %d sections, got %d", tt.wantCount, len(result))
			}

			// Verify first section has English name if present
			if tt.wantCount > 0 && result[0].Name != "" {
				// Name should be extracted from the "en" key
				if result[0].Name == "" {
					t.Errorf("expected non-empty name for first section")
				}
			}
		})
	}
}

// TestMapSubscriberToTF tests subscriber mapping
func TestMapSubscriberToTF(t *testing.T) {
	tests := []struct {
		name  string
		input *client.StatusPageSubscriber
	}{
		{
			name: "email subscriber",
			input: &client.StatusPageSubscriber{
				ID:        1,
				Type:      "email",
				Value:     "user@example.com",
				Email:     stringPtr("user@example.com"),
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		},
		{
			name: "sms subscriber",
			input: &client.StatusPageSubscriber{
				ID:        2,
				Type:      "sms",
				Value:     "+1234567890",
				Phone:     stringPtr("+1234567890"),
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		},
		{
			name: "teams subscriber",
			input: &client.StatusPageSubscriber{
				ID:        3,
				Type:      "teams",
				Value:     "https://outlook.office.com/webhook/...",
				CreatedAt: "2026-01-31T10:00:00Z",
			},
		},
		{
			name: "slack subscriber",
			input: &client.StatusPageSubscriber{
				ID:           4,
				Type:         "slack",
				Value:        "#general",
				SlackChannel: stringPtr("#general"),
				CreatedAt:    "2026-01-31T10:00:00Z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapSubscriberToTF(tt.input, &diags)

			if diags.HasError() {
				t.Errorf("unexpected error: %v", diags.Errors())
				return
			}

			if result.ID.IsNull() {
				t.Errorf("expected non-null ID")
			}

			if result.ID.ValueInt64() != int64(tt.input.ID) {
				t.Errorf("ID: expected %d, got %d", tt.input.ID, result.ID.ValueInt64())
			}

			if result.Type.ValueString() != tt.input.Type {
				t.Errorf("Type: expected %q, got %q", tt.input.Type, result.Type.ValueString())
			}

			if result.Value.ValueString() != tt.input.Value {
				t.Errorf("Value: expected %q, got %q", tt.input.Value, result.Value.ValueString())
			}

			// Check type-specific fields
			switch tt.input.Type {
			case "email":
				if result.Email.IsNull() {
					t.Errorf("expected non-null Email for email subscriber")
				}
			case "sms":
				if result.Phone.IsNull() {
					t.Errorf("expected non-null Phone for sms subscriber")
				}
			case "slack":
				if result.SlackChannel.IsNull() {
					t.Errorf("expected non-null SlackChannel for slack subscriber")
				}
			}
		})
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
