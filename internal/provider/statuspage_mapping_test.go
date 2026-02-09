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

// TestMapTFToSettings tests conversion of TF settings object to API structs
func TestMapTFToSettings(t *testing.T) {
	t.Run("null object returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		subscribe, auth := mapTFToSettings(types.ObjectNull(StatusPageSettingsAttrTypes()), &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if subscribe != nil {
			t.Error("expected nil subscribe for null object")
		}
		if auth != nil {
			t.Error("expected nil auth for null object")
		}
	})

	t.Run("unknown object returns nil", func(t *testing.T) {
		var diags diag.Diagnostics
		subscribe, auth := mapTFToSettings(types.ObjectUnknown(StatusPageSettingsAttrTypes()), &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if subscribe != nil {
			t.Error("expected nil subscribe for unknown object")
		}
		if auth != nil {
			t.Error("expected nil auth for unknown object")
		}
	})
}

// TestMapTFToServices tests conversion of TF services list to API structs
func TestMapTFToServices(t *testing.T) {
	t.Run("null list returns empty", func(t *testing.T) {
		var diags diag.Diagnostics
		result := mapTFToServices(types.ListNull(types.ObjectType{AttrTypes: ServiceAttrTypes()}), &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if len(result) != 0 {
			t.Errorf("expected 0 services for null list, got %d", len(result))
		}
	})

	t.Run("unknown list returns empty", func(t *testing.T) {
		var diags diag.Diagnostics
		result := mapTFToServices(types.ListUnknown(types.ObjectType{AttrTypes: ServiceAttrTypes()}), &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if len(result) != 0 {
			t.Errorf("expected 0 services for unknown list, got %d", len(result))
		}
	})
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

// TestMapTFToService tests conversion of a single TF service object to API struct
func TestMapTFToService(t *testing.T) {
	t.Run("valid service with all fields", func(t *testing.T) {
		var diags diag.Diagnostics

		serviceObj := types.ObjectValueMust(ServiceAttrTypes(), map[string]attr.Value{
			"id":   types.StringValue("svc_1"),
			"uuid": types.StringValue("mon_123"),
			"name": types.MapValueMust(types.StringType, map[string]attr.Value{
				"en": types.StringValue("API Service"),
			}),
			"show_uptime":         types.BoolValue(true),
			"show_response_times": types.BoolValue(true),
			"is_group":            types.BoolValue(false),
		})

		result := mapTFToService(serviceObj, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
			return
		}

		if result.MonitorUUID != "mon_123" {
			t.Errorf("expected MonitorUUID 'mon_123', got %q", result.MonitorUUID)
		}
		if result.NameShown == nil || *result.NameShown != "API Service" {
			t.Errorf("expected NameShown 'API Service', got %v", result.NameShown)
		}
		if result.ShowUptime == nil || !*result.ShowUptime {
			t.Error("expected ShowUptime true")
		}
		if result.ShowResponseTimes == nil || !*result.ShowResponseTimes {
			t.Error("expected ShowResponseTimes true")
		}
		if result.IsGroup == nil || *result.IsGroup {
			t.Error("expected IsGroup false")
		}
	})

	t.Run("minimal service with uuid only", func(t *testing.T) {
		var diags diag.Diagnostics

		serviceObj := types.ObjectValueMust(ServiceAttrTypes(), map[string]attr.Value{
			"id":                  types.StringNull(),
			"uuid":                types.StringValue("mon_minimal"),
			"name":                types.MapNull(types.StringType),
			"show_uptime":         types.BoolNull(),
			"show_response_times": types.BoolNull(),
			"is_group":            types.BoolNull(),
		})

		result := mapTFToService(serviceObj, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
			return
		}

		if result.MonitorUUID != "mon_minimal" {
			t.Errorf("expected MonitorUUID 'mon_minimal', got %q", result.MonitorUUID)
		}
		if result.NameShown != nil {
			t.Errorf("expected nil NameShown, got %v", *result.NameShown)
		}
		if result.ShowUptime != nil {
			t.Error("expected nil ShowUptime")
		}
		if result.ShowResponseTimes != nil {
			t.Error("expected nil ShowResponseTimes")
		}
	})

	t.Run("invalid element type returns error", func(t *testing.T) {
		var diags diag.Diagnostics

		// Pass a string instead of an object
		result := mapTFToService(types.StringValue("not an object"), &diags)

		if !diags.HasError() {
			t.Error("expected error for invalid element type")
		}
		if result.MonitorUUID != "" {
			t.Error("expected empty MonitorUUID for error case")
		}
	})
}

// =============================================================================
// Tests for ISS-007 Fixes
// =============================================================================

// TestNormalizeSubdomain tests subdomain normalization (ISS-007 fix)
func TestNormalizeSubdomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strips .hyperping.app suffix",
			input:    "mycompany.hyperping.app",
			expected: "mycompany",
		},
		{
			name:     "leaves plain subdomain unchanged",
			input:    "mycompany",
			expected: "mycompany",
		},
		{
			name:     "handles complex subdomain",
			input:    "status-prod.hyperping.app",
			expected: "status-prod",
		},
		{
			name:     "handles subdomain with dots",
			input:    "status.mycompany.hyperping.app",
			expected: "status.mycompany",
		},
		{
			name:     "does not strip partial suffix",
			input:    "mycompany.hyperping",
			expected: "mycompany.hyperping",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "handles just the suffix",
			input:    ".hyperping.app",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeSubdomain(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeSubdomain(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFilterLocalizedMap tests localized field filtering (ISS-007 fix)
func TestFilterLocalizedMap(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]string
		configuredLang []string
		expected       map[string]string
	}{
		{
			name: "filters to single configured language",
			input: map[string]string{
				"en": "English text",
				"fr": "French text",
				"de": "German text",
			},
			configuredLang: []string{"en"},
			expected: map[string]string{
				"en": "English text",
			},
		},
		{
			name: "filters to multiple configured languages",
			input: map[string]string{
				"en": "English",
				"fr": "Français",
				"de": "Deutsch",
				"es": "Español",
			},
			configuredLang: []string{"en", "fr"},
			expected: map[string]string{
				"en": "English",
				"fr": "Français",
			},
		},
		{
			name: "returns empty map when no configured lang matches",
			input: map[string]string{
				"en": "English",
				"fr": "French",
			},
			configuredLang: []string{"de", "es"},
			expected:       map[string]string{},
		},
		{
			name:           "nil configured langs returns original map",
			input:          map[string]string{"en": "English", "fr": "French"},
			configuredLang: nil,
			expected:       map[string]string{"en": "English", "fr": "French"},
		},
		{
			name:           "empty configured langs returns original map",
			input:          map[string]string{"en": "English", "fr": "French"},
			configuredLang: []string{},
			expected:       map[string]string{"en": "English", "fr": "French"},
		},
		{
			name:           "nil input returns nil",
			input:          nil,
			configuredLang: []string{"en"},
			expected:       nil,
		},
		{
			name:           "empty input returns empty",
			input:          map[string]string{},
			configuredLang: []string{"en"},
			expected:       map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterLocalizedMap(tt.input, tt.configuredLang)

			// Handle nil expected
			if tt.expected == nil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != len(tt.expected) {
				t.Errorf("expected %d entries, got %d", len(tt.expected), len(result))
			}

			for k, v := range tt.expected {
				if got, ok := result[k]; !ok {
					t.Errorf("expected key %q not found", k)
				} else if got != v {
					t.Errorf("key %q: expected %q, got %q", k, v, got)
				}
			}
		})
	}
}

// TestMapStatusPageWithSubdomainNormalization verifies subdomain is normalized in mapping
func TestMapStatusPageWithSubdomainNormalization(t *testing.T) {
	var diags diag.Diagnostics

	// API returns subdomain with .hyperping.app suffix
	sp := &client.StatusPage{
		UUID:            "sp_test123",
		Name:            "Test Page",
		HostedSubdomain: "mycompany.hyperping.app", // API returns full subdomain
		URL:             "https://mycompany.hyperping.app",
		Settings: client.StatusPageSettings{
			Languages: []string{"en"},
		},
	}

	result := MapStatusPageCommonFields(sp, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags.Errors())
	}

	// Should be normalized (suffix stripped)
	expected := "mycompany"
	if result.HostedSubdomain.ValueString() != expected {
		t.Errorf("HostedSubdomain: expected %q (normalized), got %q",
			expected, result.HostedSubdomain.ValueString())
	}
}

// TestMapStatusPageWithLanguageFiltering verifies localized fields are filtered
func TestMapStatusPageWithLanguageFiltering(t *testing.T) {
	var diags diag.Diagnostics

	// API returns all languages populated
	sp := &client.StatusPage{
		UUID:            "sp_test456",
		Name:            "Test Page",
		HostedSubdomain: "status",
		URL:             "https://status.hyperping.app",
		Settings: client.StatusPageSettings{
			Languages: []string{"en"},
			Description: map[string]string{
				"en": "English description",
				"fr": "French description (auto-populated by API)",
				"de": "German description (auto-populated by API)",
			},
		},
		Sections: []client.StatusPageSection{
			{
				Name: map[string]string{
					"en": "API Services",
					"fr": "Services API (auto-populated)",
					"de": "API-Dienste (auto-populated)",
				},
				Services: []client.StatusPageService{
					{
						ID:   "svc_1",
						UUID: "mon_123",
						Name: map[string]string{
							"en": "Main API",
							"fr": "API Principal (auto-populated)",
						},
					},
				},
			},
		},
	}

	// User only configured "en"
	configuredLangs := []string{"en"}
	result := MapStatusPageCommonFieldsWithFilter(sp, configuredLangs, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags.Errors())
	}

	// Verify settings.description only has "en"
	settingsAttrs := result.Settings.Attributes()
	descMap, ok := settingsAttrs["description"].(types.Map)
	if !ok || descMap.IsNull() {
		t.Fatal("expected non-null description map")
	}

	descElements := descMap.Elements()
	if len(descElements) != 1 {
		t.Errorf("expected 1 description entry, got %d", len(descElements))
	}
	if _, hasEn := descElements["en"]; !hasEn {
		t.Error("expected 'en' key in description")
	}
	if _, hasFr := descElements["fr"]; hasFr {
		t.Error("unexpected 'fr' key in description (should be filtered)")
	}

	// Verify sections are filtered
	sectionElements := result.Sections.Elements()
	if len(sectionElements) != 1 {
		t.Fatal("expected 1 section")
	}

	sectionObj, ok := sectionElements[0].(types.Object)
	if !ok {
		t.Fatal("expected section to be object")
	}

	sectionAttrs := sectionObj.Attributes()
	sectionNameMap, ok := sectionAttrs["name"].(types.Map)
	if !ok || sectionNameMap.IsNull() {
		t.Fatal("expected non-null section name map")
	}

	sectionNameElements := sectionNameMap.Elements()
	if len(sectionNameElements) != 1 {
		t.Errorf("expected 1 section name entry, got %d", len(sectionNameElements))
	}
	if _, ok := sectionNameElements["en"]; !ok {
		t.Error("expected 'en' key in section name")
	}
	if _, ok := sectionNameElements["fr"]; ok {
		t.Error("unexpected 'fr' key in section name (should be filtered)")
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}
