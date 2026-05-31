// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hyperping "github.com/develeap/hyperping-go"

	"github.com/develeap/terraform-provider-hyperping/internal/provider/testutil"
)

// =============================================================================
// extractLocalizedString
// =============================================================================

func TestExtractLocalizedString(t *testing.T) {
	tests := []struct {
		name            string
		m               map[string]string
		configuredLangs []string
		wantNull        bool
		wantValue       string
	}{
		{
			name:            "nil map returns null",
			m:               nil,
			configuredLangs: nil,
			wantNull:        true,
		},
		{
			name:            "empty map returns null",
			m:               map[string]string{},
			configuredLangs: nil,
			wantNull:        true,
		},
		{
			name:            "map with en only returns en value",
			m:               map[string]string{"en": "hello"},
			configuredLangs: nil,
			wantNull:        false,
			wantValue:       "hello",
		},
		{
			name:            "map with en and fr returns en value",
			m:               map[string]string{"en": "hello", "fr": "bonjour"},
			configuredLangs: []string{"en", "fr"},
			wantNull:        false,
			wantValue:       "hello",
		},
		{
			name:            "en empty and fr populated skips en returns fr",
			m:               map[string]string{"en": "", "fr": "bonjour"},
			configuredLangs: []string{"en", "fr"},
			wantNull:        false,
			wantValue:       "bonjour",
		},
		{
			name:            "en empty and fr populated with nil configuredLangs falls back to fr",
			m:               map[string]string{"en": "", "fr": "bonjour"},
			configuredLangs: nil,
			wantNull:        false,
			wantValue:       "bonjour",
		},
		{
			name:            "en empty and fr populated with empty configuredLangs falls back to fr",
			m:               map[string]string{"en": "", "fr": "bonjour"},
			configuredLangs: []string{},
			wantNull:        false,
			wantValue:       "bonjour",
		},
		{
			name:            "only unconfigured languages falls back to any non-empty value",
			m:               map[string]string{"de": "hallo", "es": "hola"},
			configuredLangs: []string{"en", "fr"},
			wantNull:        false,
			// Cannot predict map iteration order, but should be one of de/es
		},
		{
			name:            "all values empty with en key returns empty string",
			m:               map[string]string{"en": "", "fr": ""},
			configuredLangs: nil,
			wantNull:        false,
			wantValue:       "",
		},
		{
			name:            "all values empty without en key returns null",
			m:               map[string]string{"fr": "", "de": ""},
			configuredLangs: nil,
			wantNull:        true,
		},
		{
			name:            "configured lang matches before fallback",
			m:               map[string]string{"fr": "bonjour", "de": "hallo"},
			configuredLangs: []string{"fr"},
			wantNull:        false,
			wantValue:       "bonjour",
		},
		{
			name:            "configured lang first entry is empty second is populated",
			m:               map[string]string{"fr": "", "de": "hallo"},
			configuredLangs: []string{"fr", "de"},
			wantNull:        false,
			wantValue:       "hallo",
		},
		{
			name:            "single non-en language no configured langs",
			m:               map[string]string{"ja": "konnichiwa"},
			configuredLangs: nil,
			wantNull:        false,
			wantValue:       "konnichiwa",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractLocalizedString(tt.m, tt.configuredLangs)

			if tt.wantNull {
				if !result.IsNull() {
					t.Errorf("expected null, got %q", result.ValueString())
				}
				return
			}

			if result.IsNull() {
				t.Fatal("expected non-null result")
			}

			// For the "unconfigured languages" case we cannot predict map iteration order,
			// so we only verify the value is non-null and non-empty.
			if tt.name == "only unconfigured languages falls back to any non-empty value" {
				if result.ValueString() == "" {
					t.Error("expected non-empty fallback value")
				}
				return
			}

			if result.ValueString() != tt.wantValue {
				t.Errorf("got %q, want %q", result.ValueString(), tt.wantValue)
			}
		})
	}
}

// =============================================================================
// filterLocalizedMap
// =============================================================================

func TestFilterLocalizedMap_Coverage(t *testing.T) {
	tests := []struct {
		name            string
		m               map[string]string
		configuredLangs []string
		wantNil         bool
		wantLen         int
		wantKeys        []string
	}{
		{
			name:            "nil configuredLangs returns original map",
			m:               map[string]string{"en": "a", "fr": "b"},
			configuredLangs: nil,
			wantLen:         2,
			wantKeys:        []string{"en", "fr"},
		},
		{
			name:            "empty configuredLangs returns original map",
			m:               map[string]string{"en": "a", "fr": "b"},
			configuredLangs: []string{},
			wantLen:         2,
			wantKeys:        []string{"en", "fr"},
		},
		{
			name:            "nil map with configuredLangs returns nil",
			m:               nil,
			configuredLangs: []string{"en"},
			wantNil:         true,
		},
		{
			name:            "empty map with configuredLangs returns empty",
			m:               map[string]string{},
			configuredLangs: []string{"en"},
			wantLen:         0,
		},
		{
			name:            "filters to subset",
			m:               map[string]string{"en": "a", "fr": "b", "de": "c"},
			configuredLangs: []string{"en", "de"},
			wantLen:         2,
			wantKeys:        []string{"en", "de"},
		},
		{
			name:            "no matching languages returns empty map",
			m:               map[string]string{"en": "a", "fr": "b"},
			configuredLangs: []string{"de", "es"},
			wantLen:         0,
		},
		{
			name:            "single language filter",
			m:               map[string]string{"en": "a", "fr": "b", "de": "c"},
			configuredLangs: []string{"fr"},
			wantLen:         1,
			wantKeys:        []string{"fr"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterLocalizedMap(tt.m, tt.configuredLangs)

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("expected len %d, got %d (map: %v)", tt.wantLen, len(result), result)
			}

			for _, key := range tt.wantKeys {
				if _, ok := result[key]; !ok {
					t.Errorf("expected key %q in result", key)
				}
			}
		})
	}
}

// =============================================================================
// mapTFToStringMap
// =============================================================================

func TestMapTFToStringMap_Coverage(t *testing.T) {
	tests := []struct {
		name      string
		input     types.Map
		wantNil   bool
		wantLen   int
		wantPairs map[string]string
		wantError bool
	}{
		{
			name:    "null map returns nil",
			input:   types.MapNull(types.StringType),
			wantNil: true,
		},
		{
			name:    "unknown map returns nil",
			input:   types.MapUnknown(types.StringType),
			wantNil: true,
		},
		{
			name: "populated map returns correct entries",
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"en": types.StringValue("English"),
				"fr": types.StringValue("French"),
			}),
			wantLen: 2,
			wantPairs: map[string]string{
				"en": "English",
				"fr": "French",
			},
		},
		{
			name: "null values in map are skipped",
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"en":   types.StringValue("English"),
				"null": types.StringNull(),
			}),
			wantLen: 1,
			wantPairs: map[string]string{
				"en": "English",
			},
		},
		{
			name: "empty string values are preserved",
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"en": types.StringValue(""),
			}),
			wantLen: 1,
			wantPairs: map[string]string{
				"en": "",
			},
		},
		{
			name: "single entry map",
			input: types.MapValueMust(types.StringType, map[string]attr.Value{
				"ja": types.StringValue("Japanese"),
			}),
			wantLen: 1,
			wantPairs: map[string]string{
				"ja": "Japanese",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapTFToStringMap(tt.input, &diags)

			if tt.wantError {
				if !diags.HasError() {
					t.Error("expected diagnostics error")
				}
				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}

			if tt.wantNil {
				if result != nil {
					t.Errorf("expected nil, got %v", result)
				}
				return
			}

			if len(result) != tt.wantLen {
				t.Errorf("expected len %d, got %d", tt.wantLen, len(result))
			}

			for k, v := range tt.wantPairs {
				got, ok := result[k]
				if !ok {
					t.Errorf("expected key %q not found", k)
					continue
				}
				if got != v {
					t.Errorf("key %q: got %q, want %q", k, got, v)
				}
			}
		})
	}
}

// =============================================================================
// mapStringMapToTF
// =============================================================================

func TestMapStringMapToTF_Coverage(t *testing.T) {
	tests := []struct {
		name      string
		input     map[string]string
		wantNull  bool
		wantLen   int
		wantPairs map[string]string
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
			name:    "single entry",
			input:   map[string]string{"en": "hello"},
			wantLen: 1,
			wantPairs: map[string]string{
				"en": "hello",
			},
		},
		{
			name: "multiple entries",
			input: map[string]string{
				"en": "English",
				"fr": "French",
				"de": "German",
			},
			wantLen: 3,
			wantPairs: map[string]string{
				"en": "English",
				"fr": "French",
				"de": "German",
			},
		},
		{
			name:    "entry with empty string value",
			input:   map[string]string{"en": ""},
			wantLen: 1,
			wantPairs: map[string]string{
				"en": "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapStringMapToTF(tt.input, nil)

			if tt.wantNull {
				if !result.IsNull() {
					t.Errorf("expected null, got non-null map with %d elements", len(result.Elements()))
				}
				return
			}

			if result.IsNull() {
				t.Fatal("expected non-null map")
			}

			elements := result.Elements()
			if len(elements) != tt.wantLen {
				t.Errorf("expected %d elements, got %d", tt.wantLen, len(elements))
			}

			for k, wantVal := range tt.wantPairs {
				attrVal, ok := elements[k]
				if !ok {
					t.Errorf("expected key %q not found", k)
					continue
				}
				strVal, ok := attrVal.(types.String)
				if !ok {
					t.Errorf("key %q: expected types.String, got %T", k, attrVal)
					continue
				}
				if strVal.ValueString() != wantVal {
					t.Errorf("key %q: got %q, want %q", k, strVal.ValueString(), wantVal)
				}
			}
		})
	}
}

// =============================================================================
// serviceIDToString (supplementary cases not covered by statuspage_nested_service_test.go)
// =============================================================================

func TestServiceIDToString_SupplementaryCases(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "float64 large value",
			input: float64(9999999),
			want:  "9999999",
		},
		{
			name:  "float64 with fractional part truncated",
			input: float64(123.456),
			want:  "123",
		},
		{
			name:  "int falls through to default",
			input: 42,
			want:  "42",
		},
		{
			name:  "int64 falls through to default",
			input: int64(99),
			want:  "99",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := serviceIDToString(tt.input)
			if got != tt.want {
				t.Errorf("serviceIDToString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// =============================================================================
// mapListToStringSlice
// =============================================================================

func TestMapListToStringSlice_Coverage(t *testing.T) {
	tests := []struct {
		name      string
		input     types.List
		want      []string
		wantError bool
	}{
		{
			name:  "null list returns empty slice",
			input: types.ListNull(types.StringType),
			want:  []string{},
		},
		{
			name:  "unknown list returns empty slice",
			input: types.ListUnknown(types.StringType),
			want:  []string{},
		},
		{
			name: "populated list with multiple values",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("en"),
				types.StringValue("fr"),
				types.StringValue("de"),
			}),
			want: []string{"en", "fr", "de"},
		},
		{
			name: "single element list",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("only"),
			}),
			want: []string{"only"},
		},
		{
			name: "list with empty string element",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue(""),
			}),
			want: []string{""},
		},
		{
			name: "list with null element produces empty string",
			input: types.ListValueMust(types.StringType, []attr.Value{
				types.StringValue("first"),
				types.StringNull(),
				types.StringValue("third"),
			}),
			want: []string{"first", "", "third"},
		},
		{
			name: "invalid element type produces diagnostics error",
			input: types.ListValueMust(types.BoolType, []attr.Value{
				types.BoolValue(true),
			}),
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapListToStringSlice(tt.input, &diags)

			if tt.wantError {
				if !diags.HasError() {
					t.Error("expected diagnostics error")
				}
				return
			}

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}

			if len(result) != len(tt.want) {
				t.Fatalf("expected %d elements, got %d", len(tt.want), len(result))
			}

			for i, wantVal := range tt.want {
				if result[i] != wantVal {
					t.Errorf("index %d: got %q, want %q", i, result[i], wantVal)
				}
			}
		})
	}
}

// =============================================================================
// Retained coverage tests from original file (mapTFToSettings, mapTFToSections,
// mapTFToServices, mapSettingsToTFWithFilter edge cases, subscriber mapping)
// =============================================================================

func TestMapTFToSettings_WithValues(t *testing.T) {
	subscribeObj, _ := types.ObjectValue(SubscribeSettingsAttrTypes(), map[string]attr.Value{
		"enabled": types.BoolValue(true),
		"email":   types.BoolValue(true),
		"slack":   types.BoolValue(false),
		"teams":   types.BoolValue(true),
		"sms":     types.BoolValue(false),
	})

	authObj, _ := types.ObjectValue(AuthenticationSettingsAttrTypes(), map[string]attr.Value{
		"password_protection": types.BoolValue(true),
		"google_sso":          types.BoolValue(false),
		"saml_sso":            types.BoolValue(true),
		"allowed_domains": types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("example.com"),
			types.StringValue("test.com"),
		}),
		"sso_connection_uuid": types.StringNull(),
	})

	settingsObj, _ := types.ObjectValue(StatusPageSettingsAttrTypes(), map[string]attr.Value{
		"name":                     types.StringValue("Test"),
		"website":                  types.StringValue("https://example.com"),
		"description":              types.StringNull(),
		"languages":                types.ListNull(types.StringType),
		"default_language":         types.StringValue("en"),
		"theme":                    types.StringValue("light"),
		"font":                     types.StringValue("inter"),
		"accent_color":             types.StringValue("#3B82F6"),
		"auto_refresh":             types.BoolValue(true),
		"banner_header":            types.BoolValue(false),
		"logo":                     types.StringNull(),
		"logo_height":              types.StringValue("40px"),
		"favicon":                  types.StringNull(),
		"hide_powered_by":          types.BoolValue(false),
		"hide_from_search_engines": types.BoolValue(false),
		"google_analytics":         types.StringNull(),
		"subscribe":                subscribeObj,
		"authentication":           authObj,
	})

	var diags diag.Diagnostics
	subscribe, auth := mapTFToSettings(context.Background(), settingsObj, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags.Errors())
	}

	if subscribe == nil {
		t.Error("expected non-nil subscribe")
	} else {
		if subscribe.Enabled == nil || !*subscribe.Enabled {
			t.Error("subscribe.Enabled should be true")
		}
		if subscribe.Email == nil || !*subscribe.Email {
			t.Error("subscribe.Email should be true")
		}
		if subscribe.Slack == nil || *subscribe.Slack {
			t.Error("subscribe.Slack should be false")
		}
	}

	if auth == nil {
		t.Error("expected non-nil auth")
	} else {
		if auth.PasswordProtection == nil || !*auth.PasswordProtection {
			t.Error("auth.PasswordProtection should be true")
		}
		if len(auth.AllowedDomains) != 2 {
			t.Errorf("expected 2 allowed domains, got %d", len(auth.AllowedDomains))
		}
	}
}

func TestMapTFToSections_WithValues(t *testing.T) {
	serviceObj, _ := types.ObjectValue(ServiceAttrTypes(), map[string]attr.Value{
		"id":   types.StringValue("svc_1"),
		"uuid": types.StringValue("mon_123"),
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("API Gateway"),
			"fr": types.StringValue("Passerelle API"),
		}),
		"is_group":            types.BoolValue(false),
		"show_uptime":         types.BoolValue(true),
		"show_response_times": types.BoolValue(true),
		"description":         types.MapNull(types.StringType),
		"services":            types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}),
	})

	servicesList, _ := types.ListValue(types.ObjectType{AttrTypes: ServiceAttrTypes()}, []attr.Value{serviceObj})

	sectionObj, _ := types.ObjectValue(SectionAttrTypes(), map[string]attr.Value{
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("Services"),
			"fr": types.StringValue("Prestations de service"),
		}),
		"is_split": types.BoolValue(false),
		"services": servicesList,
	})

	sectionsList, _ := types.ListValue(types.ObjectType{AttrTypes: SectionAttrTypes()}, []attr.Value{sectionObj})

	var diags diag.Diagnostics
	sections := mapTFToSections(sectionsList, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags.Errors())
	}

	if len(sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(sections))
	}

	section := sections[0]
	if section.Name != "Services" {
		t.Errorf("section name = %q, want 'Services'", section.Name)
	}

	if len(section.Services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(section.Services))
	}
}

func TestMapTFToServices_WithValues(t *testing.T) {
	service1, _ := types.ObjectValue(ServiceAttrTypes(), map[string]attr.Value{
		"id":   types.StringValue("svc_1"),
		"uuid": types.StringValue("mon_123"),
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("API"),
		}),
		"is_group":            types.BoolValue(false),
		"show_uptime":         types.BoolValue(true),
		"show_response_times": types.BoolValue(false),
		"description":         types.MapNull(types.StringType),
		"services":            types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}),
	})

	nestedSvc, _ := types.ObjectValue(NestedServiceAttrTypes(), map[string]attr.Value{
		"id":   types.StringValue("nsvc_1"),
		"uuid": types.StringValue("mon_456"),
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("Primary DB"),
		}),
		"is_group":            types.BoolValue(false),
		"show_uptime":         types.BoolValue(false),
		"show_response_times": types.BoolValue(false),
		"description":         types.MapNull(types.StringType),
	})
	nestedSvcList, _ := types.ListValue(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}, []attr.Value{nestedSvc})

	service2, _ := types.ObjectValue(ServiceAttrTypes(), map[string]attr.Value{
		"id":   types.StringValue("svc_2"),
		"uuid": types.StringNull(),
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("Database Group"),
		}),
		"is_group":            types.BoolValue(true),
		"show_uptime":         types.BoolValue(false),
		"show_response_times": types.BoolValue(true),
		"description":         types.MapNull(types.StringType),
		"services":            nestedSvcList,
	})

	servicesList, _ := types.ListValue(types.ObjectType{AttrTypes: ServiceAttrTypes()}, []attr.Value{service1, service2})

	var diags diag.Diagnostics
	services := mapTFToServices(servicesList, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags.Errors())
	}

	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}

	if services[0].MonitorUUID == nil || *services[0].MonitorUUID != "mon_123" {
		t.Errorf("service[0].MonitorUUID = %v, want 'mon_123'", services[0].MonitorUUID)
	}

	if services[1].MonitorUUID != nil {
		t.Errorf("service[1].MonitorUUID should be nil for groups, got %v", *services[1].MonitorUUID)
	}
	if services[1].IsGroup == nil || !*services[1].IsGroup {
		t.Error("service[1].IsGroup should be true")
	}
}

func TestMapSettingsToTFWithFilter_EdgeCases(t *testing.T) {
	t.Run("with null logo and favicon", func(t *testing.T) {
		settings := hyperping.StatusPageSettings{
			Name:            "Test",
			Website:         "https://example.com",
			Description:     map[string]string{"en": "English"},
			Languages:       []string{"en"},
			DefaultLanguage: "en",
			Theme:           "light",
			Font:            "inter",
			AccentColor:     "#3B82F6",
			AutoRefresh:     true,
			BannerHeader:    false,
			Logo:            nil,
			LogoHeight:      "40px",
			Favicon:         nil,
			HidePoweredBy:   false,
			Subscribe: hyperping.StatusPageSubscribeSettings{
				Enabled: true,
				Email:   true,
			},
			Authentication: hyperping.StatusPageAuthenticationSettings{
				AllowedDomains: []string{},
			},
		}

		var diags diag.Diagnostics
		result := mapSettingsToTFWithFilter(settings, nil, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}

		if result.IsNull() {
			t.Error("expected non-null settings object")
		}

		attrs := result.Attributes()
		logo, _ := attrs["logo"].(types.String)
		if !logo.IsNull() {
			t.Error("expected null logo")
		}
	})

	t.Run("with empty strings for optionals", func(t *testing.T) {
		emptyStr := ""
		settings := hyperping.StatusPageSettings{
			Name:            "Test",
			Website:         "https://example.com",
			Description:     map[string]string{},
			Languages:       []string{},
			DefaultLanguage: "en",
			Theme:           "light",
			Font:            "inter",
			AccentColor:     "#3B82F6",
			Logo:            &emptyStr,
			LogoHeight:      "40px",
			Favicon:         &emptyStr,
			GoogleAnalytics: &emptyStr,
			Subscribe:       hyperping.StatusPageSubscribeSettings{},
			Authentication: hyperping.StatusPageAuthenticationSettings{
				AllowedDomains: []string{},
			},
		}

		var diags diag.Diagnostics
		result := mapSettingsToTFWithFilter(settings, nil, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}

		attrs := result.Attributes()
		logo, _ := attrs["logo"].(types.String)
		if !logo.IsNull() {
			t.Error("expected null logo for empty string")
		}
	})
}

func TestMapTFToStringMap_InvalidElement(t *testing.T) {
	tfMap := types.MapValueMust(types.StringType, map[string]attr.Value{
		"en":   types.StringValue("English"),
		"null": types.StringNull(),
	})

	var diags diag.Diagnostics
	result := mapTFToStringMap(tfMap, &diags)

	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags.Errors())
	}

	if _, ok := result["null"]; ok {
		t.Error("expected null value to be skipped")
	}

	if result["en"] != "English" {
		t.Errorf("en = %q, want 'English'", result["en"])
	}
}

func TestMapListToStringSlice_WithInvalidElement(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		emptyList, _ := types.ListValue(types.StringType, []attr.Value{})

		var diags diag.Diagnostics
		result := mapListToStringSlice(emptyList, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}

		if len(result) != 0 {
			t.Errorf("expected empty result, got %d elements", len(result))
		}
	})

	t.Run("single element", func(t *testing.T) {
		singleList, _ := types.ListValue(types.StringType, []attr.Value{
			types.StringValue("only-one"),
		})

		var diags diag.Diagnostics
		result := mapListToStringSlice(singleList, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}

		if len(result) != 1 || result[0] != "only-one" {
			t.Errorf("expected [only-one], got %v", result)
		}
	})
}

func TestMapSubscriberToTF_EmptyOptionals(t *testing.T) {
	emptyStr := ""
	subscriber := &hyperping.StatusPageSubscriber{
		ID:           123,
		Type:         "webhook",
		Value:        "https://webhook.example.com",
		Email:        &emptyStr,
		Phone:        &emptyStr,
		SlackChannel: &emptyStr,
		CreatedAt:    "2026-01-01T00:00:00Z",
	}

	var diags diag.Diagnostics
	result := mapSubscriberToTF(subscriber, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags.Errors())
	}

	if !result.Email.IsNull() {
		t.Error("expected null email for empty string")
	}
	if !result.Phone.IsNull() {
		t.Error("expected null phone for empty string")
	}
	if !result.SlackChannel.IsNull() {
		t.Error("expected null slack_channel for empty string")
	}
}

func TestMapTFToSections_NullAndUnknown(t *testing.T) {
	t.Run("null list", func(t *testing.T) {
		nullList := types.ListNull(types.ObjectType{AttrTypes: SectionAttrTypes()})
		var diags diag.Diagnostics
		result := mapTFToSections(nullList, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if result != nil {
			t.Error("expected nil result for null list")
		}
	})

	t.Run("unknown list", func(t *testing.T) {
		unknownList := types.ListUnknown(types.ObjectType{AttrTypes: SectionAttrTypes()})
		var diags diag.Diagnostics
		result := mapTFToSections(unknownList, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if result != nil {
			t.Error("expected nil result for unknown list")
		}
	})

	t.Run("invalid element type", func(t *testing.T) {
		invalidList, _ := types.ListValue(types.StringType, []attr.Value{
			types.StringValue("invalid"),
		})
		var diags diag.Diagnostics
		result := mapTFToSections(invalidList, &diags)

		if !diags.HasError() {
			t.Error("expected error for invalid element type")
		}
		if len(result) != 0 {
			t.Error("expected empty result when element is invalid")
		}
	})

	t.Run("section with name without en key", func(t *testing.T) {
		sectionObj, _ := types.ObjectValue(SectionAttrTypes(), map[string]attr.Value{
			"name": types.MapValueMust(types.StringType, map[string]attr.Value{
				"fr": types.StringValue("Francais"),
				"de": types.StringValue("Deutsch"),
			}),
			"is_split": types.BoolValue(false),
			"services": types.ListNull(types.ObjectType{AttrTypes: ServiceAttrTypes()}),
		})

		sectionsList, _ := types.ListValue(types.ObjectType{AttrTypes: SectionAttrTypes()}, []attr.Value{sectionObj})

		var diags diag.Diagnostics
		sections := mapTFToSections(sectionsList, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}

		if len(sections) != 1 {
			t.Fatalf("expected 1 section, got %d", len(sections))
		}

		if sections[0].Name == "" {
			t.Error("expected section name to be set to first available language")
		}
	})
}

func TestMapTFToStringMap_NullAndUnknown(t *testing.T) {
	t.Run("null map", func(t *testing.T) {
		nullMap := types.MapNull(types.StringType)
		var diags diag.Diagnostics
		result := mapTFToStringMap(nullMap, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if result != nil {
			t.Error("expected nil result for null map")
		}
	})

	t.Run("unknown map", func(t *testing.T) {
		unknownMap := types.MapUnknown(types.StringType)
		var diags diag.Diagnostics
		result := mapTFToStringMap(unknownMap, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if result != nil {
			t.Error("expected nil result for unknown map")
		}
	})

	t.Run("map with non-string value type", func(t *testing.T) {
		tfMap := types.MapValueMust(types.StringType, map[string]attr.Value{
			"en":   types.StringValue("English"),
			"fr":   types.StringValue("French"),
			"null": types.StringNull(),
		})

		var diags diag.Diagnostics
		result := mapTFToStringMap(tfMap, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}

		if _, ok := result["null"]; ok {
			t.Error("expected null values to be skipped")
		}

		if len(result) != 2 {
			t.Errorf("expected 2 entries (skipping null), got %d", len(result))
		}
	})
}

func TestMapListToStringSlice_NullAndUnknown(t *testing.T) {
	t.Run("null list", func(t *testing.T) {
		nullList := types.ListNull(types.StringType)
		var diags diag.Diagnostics
		result := mapListToStringSlice(nullList, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if len(result) != 0 {
			t.Errorf("expected empty result for null list, got %v", result)
		}
	})

	t.Run("unknown list", func(t *testing.T) {
		unknownList := types.ListUnknown(types.StringType)
		var diags diag.Diagnostics
		result := mapListToStringSlice(unknownList, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		if len(result) != 0 {
			t.Errorf("expected empty result for unknown list, got %v", result)
		}
	})

	t.Run("list with null element", func(t *testing.T) {
		listWithNull, _ := types.ListValue(types.StringType, []attr.Value{
			types.StringValue("first"),
			types.StringNull(),
			types.StringValue("third"),
		})
		var diags diag.Diagnostics
		result := mapListToStringSlice(listWithNull, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}

		if len(result) != 3 {
			t.Errorf("expected 3 elements, got %d", len(result))
		}
		if result[0] != "first" || result[1] != "" || result[2] != "third" {
			t.Errorf("expected [first, '', third], got %v", result)
		}
	})

	t.Run("list with invalid element type", func(t *testing.T) {
		invalidList, _ := types.ListValue(types.BoolType, []attr.Value{
			types.BoolValue(true),
		})
		var diags diag.Diagnostics
		result := mapListToStringSlice(invalidList, &diags)

		if !diags.HasError() {
			t.Error("expected error for invalid element type")
		}

		if len(result) != 0 {
			t.Errorf("expected empty result after error, got %d elements", len(result))
		}
	})
}

// =============================================================================
// mapTFToSettings edge cases
// =============================================================================

func TestMapTFToSettings_NullObject(t *testing.T) {
	nullObj := types.ObjectNull(StatusPageSettingsAttrTypes())
	var diags diag.Diagnostics
	subscribe, auth := mapTFToSettings(context.Background(), nullObj, &diags)

	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags.Errors())
	}
	if subscribe != nil {
		t.Error("expected nil subscribe for null object")
	}
	if auth != nil {
		t.Error("expected nil auth for null object")
	}
}

func TestMapTFToSettings_UnknownObject(t *testing.T) {
	unknownObj := types.ObjectUnknown(StatusPageSettingsAttrTypes())
	var diags diag.Diagnostics
	subscribe, auth := mapTFToSettings(context.Background(), unknownObj, &diags)

	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags.Errors())
	}
	if subscribe != nil {
		t.Error("expected nil subscribe for unknown object")
	}
	if auth != nil {
		t.Error("expected nil auth for unknown object")
	}
}

// =============================================================================
// mapTFToServices validation branches
// =============================================================================

func TestMapTFToServices_NonGroupWithoutUUID(t *testing.T) {
	// Non-group service without UUID should produce a validation error
	svc, _ := types.ObjectValue(ServiceAttrTypes(), map[string]attr.Value{
		"id":                  types.StringNull(),
		"uuid":                types.StringNull(),
		"name":                types.MapNull(types.StringType),
		"is_group":            types.BoolValue(false),
		"show_uptime":         types.BoolNull(),
		"show_response_times": types.BoolNull(),
		"description":         types.MapNull(types.StringType),
		"services":            types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}),
	})

	list, _ := types.ListValue(types.ObjectType{AttrTypes: ServiceAttrTypes()}, []attr.Value{svc})

	var diags diag.Diagnostics
	services := mapTFToServices(list, &diags)

	if !diags.HasError() {
		t.Fatal("expected error for non-group service without UUID")
	}

	if len(services) != 1 {
		t.Errorf("expected 1 service (still appended), got %d", len(services))
	}
}

func TestMapTFToServices_GroupWithEmptyServices(t *testing.T) {
	// Group service with empty services list should produce a validation error
	svc, _ := types.ObjectValue(ServiceAttrTypes(), map[string]attr.Value{
		"id":                  types.StringNull(),
		"uuid":                types.StringNull(),
		"name":                types.MapNull(types.StringType),
		"is_group":            types.BoolValue(true),
		"show_uptime":         types.BoolNull(),
		"show_response_times": types.BoolNull(),
		"description":         types.MapNull(types.StringType),
		"services":            types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}),
	})

	list, _ := types.ListValue(types.ObjectType{AttrTypes: ServiceAttrTypes()}, []attr.Value{svc})

	var diags diag.Diagnostics
	services := mapTFToServices(list, &diags)

	if !diags.HasError() {
		t.Fatal("expected error for group service with empty services")
	}

	if len(services) != 1 {
		t.Errorf("expected 1 service (still appended), got %d", len(services))
	}
}

func TestMapTFToServices_NullList(t *testing.T) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: ServiceAttrTypes()})
	var diags diag.Diagnostics
	services := mapTFToServices(nullList, &diags)

	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags.Errors())
	}
	if services != nil {
		t.Error("expected nil for null list")
	}
}

func TestMapTFToNestedServices_NullList(t *testing.T) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()})
	var diags diag.Diagnostics
	services := mapTFToNestedServices(nullList, &diags)

	if diags.HasError() {
		t.Errorf("unexpected error: %v", diags.Errors())
	}
	if services != nil {
		t.Error("expected nil for null list")
	}
}

// =============================================================================
// mapSettingsToTFWithFilter with non-nil non-empty optional fields
// =============================================================================

func TestMapSettingsToTFWithFilter_WithOptionalValues(t *testing.T) {
	logoVal := "https://example.com/logo.png"
	faviconVal := "https://example.com/favicon.ico"
	gaVal := "UA-12345678-1"
	settings := hyperping.StatusPageSettings{
		Name:            "Test",
		Website:         "https://example.com",
		Description:     map[string]string{"en": "Test page"},
		Languages:       []string{"en", "fr"},
		DefaultLanguage: "en",
		Theme:           "light",
		Font:            "inter",
		AccentColor:     "#3B82F6",
		AutoRefresh:     true,
		BannerHeader:    false,
		Logo:            &logoVal,
		LogoHeight:      "40px",
		Favicon:         &faviconVal,
		HidePoweredBy:   false,
		GoogleAnalytics: &gaVal,
		Subscribe: hyperping.StatusPageSubscribeSettings{
			Enabled: true,
			Email:   true,
		},
		Authentication: hyperping.StatusPageAuthenticationSettings{
			AllowedDomains: []string{"example.com"},
		},
	}

	var diags diag.Diagnostics
	result := mapSettingsToTFWithFilter(settings, []string{"en"}, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags.Errors())
	}

	attrs := result.Attributes()

	logo, _ := attrs["logo"].(types.String)
	if logo.IsNull() || logo.ValueString() != logoVal {
		t.Errorf("logo = %q, want %q", logo.ValueString(), logoVal)
	}

	favicon, _ := attrs["favicon"].(types.String)
	if favicon.IsNull() || favicon.ValueString() != faviconVal {
		t.Errorf("favicon = %q, want %q", favicon.ValueString(), faviconVal)
	}

	ga, _ := attrs["google_analytics"].(types.String)
	if ga.IsNull() || ga.ValueString() != gaVal {
		t.Errorf("google_analytics = %q, want %q", ga.ValueString(), gaVal)
	}
}

// =============================================================================
// isAllowedBaseURL edge case (port in domain)
// =============================================================================

func TestIsAllowedBaseURL_PortVariants(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		want    bool
	}{
		{"https with port", "https://api.hyperping.io:443", true},
		{"localhost with path", "http://localhost/api/v1", true},
		{"127.0.0.1 with path", "http://127.0.0.1/api", true},
		{"http non-localhost with port", "http://api.example.com:8080", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isAllowedBaseURL(tt.baseURL, true)
			if got != tt.want {
				t.Errorf("isAllowedBaseURL(%q, true) = %v, want %v", tt.baseURL, got, tt.want)
			}
		})
	}
}

// =============================================================================
// AttrTypes count assertions
// =============================================================================

func TestServiceAttrTypes_Count(t *testing.T) {
	attrs := ServiceAttrTypes()
	expectedKeys := []string{"id", "uuid", "name", "is_group", "show_uptime", "show_response_times", "description", "services"}

	if len(attrs) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d: %v", len(expectedKeys), len(attrs), keysOf(attrs))
	}

	for _, key := range expectedKeys {
		if _, ok := attrs[key]; !ok {
			t.Errorf("missing expected key %q", key)
		}
	}
}

func TestAuthenticationSettingsAttrTypes_Count(t *testing.T) {
	attrs := AuthenticationSettingsAttrTypes()
	expectedKeys := []string{"password_protection", "google_sso", "saml_sso", "allowed_domains", "sso_connection_uuid"}

	if len(attrs) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d", len(expectedKeys), len(attrs))
	}

	for _, key := range expectedKeys {
		if _, ok := attrs[key]; !ok {
			t.Errorf("missing expected key %q", key)
		}
	}
}

// =============================================================================
// SSO connection UUID mapping tests
// =============================================================================

func TestExtractAuthSettings_SSOConnectionUUID(t *testing.T) {
	t.Run("non-null sso_connection_uuid extracted", func(t *testing.T) {
		domainsList, _ := types.ListValue(types.StringType, []attr.Value{})
		obj := buildAuthObj(t, map[string]attr.Value{
			"password_protection": types.BoolValue(false),
			"google_sso":          types.BoolValue(false),
			"saml_sso":            types.BoolValue(false),
			"allowed_domains":     domainsList,
			"sso_connection_uuid": types.StringValue("uuid-123"),
		})

		var d diag.Diagnostics
		result := extractAuthSettings(context.Background(), obj, &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result.SSOConnectionUUID == nil {
			t.Fatal("expected non-nil SSOConnectionUUID")
		}
		if *result.SSOConnectionUUID != "uuid-123" {
			t.Errorf("expected 'uuid-123', got %q", *result.SSOConnectionUUID)
		}
	})

	t.Run("null sso_connection_uuid produces nil", func(t *testing.T) {
		domainsList, _ := types.ListValue(types.StringType, []attr.Value{})
		obj := buildAuthObj(t, map[string]attr.Value{
			"password_protection": types.BoolValue(false),
			"google_sso":          types.BoolValue(false),
			"saml_sso":            types.BoolValue(false),
			"allowed_domains":     domainsList,
			"sso_connection_uuid": types.StringNull(),
		})

		var d diag.Diagnostics
		result := extractAuthSettings(context.Background(), obj, &d)

		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result.SSOConnectionUUID != nil {
			t.Errorf("expected nil SSOConnectionUUID, got %q", *result.SSOConnectionUUID)
		}
	})
}

func TestMapSettingsToTFWithFilter_SSOConnectionUUID(t *testing.T) {
	t.Run("non-nil sso_connection_uuid mapped", func(t *testing.T) {
		val := "uuid-abc"
		settings := hyperping.StatusPageSettings{
			Name:            "Test",
			DefaultLanguage: "en",
			Theme:           "light",
			Font:            "inter",
			AccentColor:     "#3B82F6",
			LogoHeight:      "40px",
			Authentication: hyperping.StatusPageAuthenticationSettings{
				AllowedDomains:    []string{},
				SSOConnectionUUID: &val,
			},
			Subscribe: hyperping.StatusPageSubscribeSettings{},
		}

		var diags diag.Diagnostics
		result := mapSettingsToTFWithFilter(settings, nil, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}

		attrs := result.Attributes()
		authObj, _ := attrs["authentication"].(types.Object)
		authAttrs := authObj.Attributes()
		ssoVal, _ := authAttrs["sso_connection_uuid"].(types.String)
		if ssoVal.IsNull() {
			t.Error("expected non-null sso_connection_uuid")
		}
		if ssoVal.ValueString() != "uuid-abc" {
			t.Errorf("expected 'uuid-abc', got %q", ssoVal.ValueString())
		}
	})

	t.Run("nil sso_connection_uuid mapped as null", func(t *testing.T) {
		settings := hyperping.StatusPageSettings{
			Name:            "Test",
			DefaultLanguage: "en",
			Theme:           "light",
			Font:            "inter",
			AccentColor:     "#3B82F6",
			LogoHeight:      "40px",
			Authentication: hyperping.StatusPageAuthenticationSettings{
				AllowedDomains: []string{},
			},
			Subscribe: hyperping.StatusPageSubscribeSettings{},
		}

		var diags diag.Diagnostics
		result := mapSettingsToTFWithFilter(settings, nil, &diags)
		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}

		attrs := result.Attributes()
		authObj, _ := attrs["authentication"].(types.Object)
		authAttrs := authObj.Attributes()
		ssoVal, _ := authAttrs["sso_connection_uuid"].(types.String)
		if !ssoVal.IsNull() {
			t.Errorf("expected null sso_connection_uuid, got %q", ssoVal.ValueString())
		}
	})
}

// =============================================================================
// Service description mapping tests
// =============================================================================

func TestMapServiceToTFWithFilter_Description(t *testing.T) {
	t.Run("description populated", func(t *testing.T) {
		svc := hyperping.StatusPageService{
			ID:          testutil.Ptr(hyperping.FlexibleString("svc_1")),
			UUID:        "mon_1",
			Name:        map[string]string{"en": "API"},
			Description: map[string]string{"en": "My API"},
		}

		var d diag.Diagnostics
		tfObj := mapServiceToTFWithFilter(svc, nil, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}

		attrs := tfObj.Attributes()
		descMap, ok := attrs["description"].(types.Map)
		if !ok {
			t.Fatal("expected description to be types.Map")
		}
		if descMap.IsNull() {
			t.Fatal("expected non-null description")
		}
		elems := descMap.Elements()
		enVal, _ := elems["en"].(types.String)
		if enVal.ValueString() != "My API" {
			t.Errorf("expected 'My API', got %q", enVal.ValueString())
		}
	})

	t.Run("nil description", func(t *testing.T) {
		svc := hyperping.StatusPageService{
			ID:   testutil.Ptr(hyperping.FlexibleString("svc_1")),
			UUID: "mon_1",
			Name: map[string]string{"en": "API"},
		}

		var d diag.Diagnostics
		tfObj := mapServiceToTFWithFilter(svc, nil, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}

		attrs := tfObj.Attributes()
		descMap, _ := attrs["description"].(types.Map)
		if !descMap.IsNull() {
			t.Error("expected null description for nil input")
		}
	})
}

func TestMapNestedServicesToTF_Description(t *testing.T) {
	t.Run("description populated", func(t *testing.T) {
		svc := hyperping.StatusPageService{
			ID:          testutil.Ptr(hyperping.FlexibleString("svc_1")),
			UUID:        "mon_1",
			Name:        map[string]string{"en": "Nested"},
			Description: map[string]string{"en": "Nested svc"},
		}

		var d diag.Diagnostics
		result := mapNestedServicesToTF([]hyperping.StatusPageService{svc}, nil, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}

		obj, _ := result.Elements()[0].(types.Object)
		attrs := obj.Attributes()
		descMap, _ := attrs["description"].(types.Map)
		if descMap.IsNull() {
			t.Fatal("expected non-null description")
		}
		elems := descMap.Elements()
		enVal, _ := elems["en"].(types.String)
		if enVal.ValueString() != "Nested svc" {
			t.Errorf("expected 'Nested svc', got %q", enVal.ValueString())
		}
	})

	t.Run("nil description", func(t *testing.T) {
		svc := hyperping.StatusPageService{
			ID:   testutil.Ptr(hyperping.FlexibleString("svc_1")),
			UUID: "mon_1",
			Name: map[string]string{"en": "Nested"},
		}

		var d diag.Diagnostics
		result := mapNestedServicesToTF([]hyperping.StatusPageService{svc}, nil, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}

		obj, _ := result.Elements()[0].(types.Object)
		attrs := obj.Attributes()
		descMap, _ := attrs["description"].(types.Map)
		if !descMap.IsNull() {
			t.Error("expected null description for nil input")
		}
	})
}

func TestMapTFToService_Description(t *testing.T) {
	t.Run("en description extracted", func(t *testing.T) {
		obj := types.ObjectValueMust(ServiceAttrTypes(), map[string]attr.Value{
			"id":   types.StringNull(),
			"uuid": types.StringValue("mon_1"),
			"name": types.MapValueMust(types.StringType, map[string]attr.Value{
				"en": types.StringValue("API"),
			}),
			"is_group":            types.BoolValue(false),
			"show_uptime":         types.BoolNull(),
			"show_response_times": types.BoolNull(),
			"description": types.MapValueMust(types.StringType, map[string]attr.Value{
				"en": types.StringValue("English desc"),
			}),
			"services": types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}),
		})

		var d diag.Diagnostics
		result := mapTFToService(obj, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if descStr, ok := result.Description.(*string); !ok || descStr == nil || *descStr != "English desc" {
			t.Errorf("expected 'English desc', got %v", result.Description)
		}
	})

	t.Run("fallback to non-en language", func(t *testing.T) {
		obj := types.ObjectValueMust(ServiceAttrTypes(), map[string]attr.Value{
			"id":                  types.StringNull(),
			"uuid":                types.StringValue("mon_1"),
			"name":                types.MapNull(types.StringType),
			"is_group":            types.BoolValue(false),
			"show_uptime":         types.BoolNull(),
			"show_response_times": types.BoolNull(),
			"description": types.MapValueMust(types.StringType, map[string]attr.Value{
				"fr": types.StringValue("French desc"),
			}),
			"services": types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}),
		})

		var d diag.Diagnostics
		result := mapTFToService(obj, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if descStr, ok := result.Description.(*string); !ok || descStr == nil || *descStr != "French desc" {
			t.Errorf("expected 'French desc', got %v", result.Description)
		}
	})

	t.Run("null description produces nil", func(t *testing.T) {
		obj := types.ObjectValueMust(ServiceAttrTypes(), map[string]attr.Value{
			"id":                  types.StringNull(),
			"uuid":                types.StringValue("mon_1"),
			"name":                types.MapNull(types.StringType),
			"is_group":            types.BoolValue(false),
			"show_uptime":         types.BoolNull(),
			"show_response_times": types.BoolNull(),
			"description":         types.MapNull(types.StringType),
			"services":            types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}),
		})

		var d diag.Diagnostics
		result := mapTFToService(obj, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result.Description != nil {
			t.Errorf("expected nil description, got %q", result.Description)
		}
	})
}

func TestMapTFToNestedServices_Description(t *testing.T) {
	t.Run("description extracted", func(t *testing.T) {
		obj := types.ObjectValueMust(NestedServiceAttrTypes(), map[string]attr.Value{
			"id":                  types.StringNull(),
			"uuid":                types.StringValue("mon_1"),
			"name":                types.MapNull(types.StringType),
			"is_group":            types.BoolValue(false),
			"show_uptime":         types.BoolValue(false),
			"show_response_times": types.BoolValue(false),
			"description": types.MapValueMust(types.StringType, map[string]attr.Value{
				"en": types.StringValue("Nested desc"),
			}),
		})
		list, _ := types.ListValue(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}, []attr.Value{obj})

		var d diag.Diagnostics
		result := mapTFToNestedServices(list, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if len(result) != 1 {
			t.Fatalf("expected 1 service, got %d", len(result))
		}
		if descMap, ok := result[0].Description.(map[string]string); !ok || descMap["en"] != "Nested desc" {
			t.Errorf("expected 'Nested desc', got %v", result[0].Description)
		}
	})

	t.Run("null description produces nil", func(t *testing.T) {
		obj := types.ObjectValueMust(NestedServiceAttrTypes(), map[string]attr.Value{
			"id":                  types.StringNull(),
			"uuid":                types.StringValue("mon_1"),
			"name":                types.MapNull(types.StringType),
			"is_group":            types.BoolValue(false),
			"show_uptime":         types.BoolValue(false),
			"show_response_times": types.BoolValue(false),
			"description":         types.MapNull(types.StringType),
		})
		list, _ := types.ListValue(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}, []attr.Value{obj})

		var d diag.Diagnostics
		result := mapTFToNestedServices(list, &d)
		if d.HasError() {
			t.Fatalf("unexpected error: %v", d.Errors())
		}
		if result[0].Description != nil {
			t.Errorf("expected nil description, got %q", result[0].Description)
		}
	})
}
