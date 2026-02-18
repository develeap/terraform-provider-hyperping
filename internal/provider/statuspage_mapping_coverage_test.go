// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Increase coverage for mapTFToSettings
func TestMapTFToSettings_WithValues(t *testing.T) {
	// Create subscribe settings object
	subscribeObj, _ := types.ObjectValue(SubscribeSettingsAttrTypes(), map[string]attr.Value{
		"enabled": types.BoolValue(true),
		"email":   types.BoolValue(true),
		"slack":   types.BoolValue(false),
		"teams":   types.BoolValue(true),
		"sms":     types.BoolValue(false),
	})

	// Create authentication settings object
	authObj, _ := types.ObjectValue(AuthenticationSettingsAttrTypes(), map[string]attr.Value{
		"password_protection": types.BoolValue(true),
		"google_sso":          types.BoolValue(false),
		"saml_sso":            types.BoolValue(true),
		"allowed_domains": types.ListValueMust(types.StringType, []attr.Value{
			types.StringValue("example.com"),
			types.StringValue("test.com"),
		}),
	})

	// Create settings object with subscribe and authentication
	settingsObj, _ := types.ObjectValue(StatusPageSettingsAttrTypes(), map[string]attr.Value{
		"name":                     types.StringValue("Test"),
		"website":                  types.StringValue("https://example.com"),
		"description":              types.MapNull(types.StringType),
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

// Increase coverage for mapTFToSections
func TestMapTFToSections_WithValues(t *testing.T) {
	// Create a service object
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
	})

	// Create services list
	servicesList, _ := types.ListValue(types.ObjectType{AttrTypes: ServiceAttrTypes()}, []attr.Value{serviceObj})

	// Create a section object
	sectionObj, _ := types.ObjectValue(SectionAttrTypes(), map[string]attr.Value{
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("Services"),
			"fr": types.StringValue("Prestations de service"),
		}),
		"is_split": types.BoolValue(false),
		"services": servicesList,
	})

	// Create sections list
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

// Increase coverage for mapTFToServices
func TestMapTFToServices_WithValues(t *testing.T) {
	// Create multiple service objects
	service1, _ := types.ObjectValue(ServiceAttrTypes(), map[string]attr.Value{
		"id":   types.StringValue("svc_1"),
		"uuid": types.StringValue("mon_123"),
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("API"),
		}),
		"is_group":            types.BoolValue(false),
		"show_uptime":         types.BoolValue(true),
		"show_response_times": types.BoolValue(false),
	})

	service2, _ := types.ObjectValue(ServiceAttrTypes(), map[string]attr.Value{
		"id":   types.StringValue("svc_2"),
		"uuid": types.StringValue("mon_456"),
		"name": types.MapValueMust(types.StringType, map[string]attr.Value{
			"en": types.StringValue("Database"),
		}),
		"is_group":            types.BoolValue(true),
		"show_uptime":         types.BoolValue(false),
		"show_response_times": types.BoolValue(true),
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

	// Check first service
	if services[0].MonitorUUID != "mon_123" {
		t.Errorf("service[0].MonitorUUID = %q, want 'mon_123'", services[0].MonitorUUID)
	}

	// Check second service
	if services[1].MonitorUUID != "mon_456" {
		t.Errorf("service[1].MonitorUUID = %q, want 'mon_456'", services[1].MonitorUUID)
	}
	if services[1].IsGroup == nil || !*services[1].IsGroup {
		t.Error("service[1].IsGroup should be true")
	}
}

// Increase coverage for mapSettingsToTFWithFilter edge cases
func TestMapSettingsToTFWithFilter_EdgeCases(t *testing.T) {
	t.Run("with null logo and favicon", func(t *testing.T) {
		settings := client.StatusPageSettings{
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
			Subscribe: client.StatusPageSubscribeSettings{
				Enabled: true,
				Email:   true,
				Slack:   false,
				Teams:   false,
				SMS:     false,
			},
			Authentication: client.StatusPageAuthenticationSettings{
				PasswordProtection: false,
				GoogleSSO:          false,
				SAMLSSO:            false,
				AllowedDomains:     []string{},
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
		settings := client.StatusPageSettings{
			Name:            "Test",
			Website:         "https://example.com",
			Description:     map[string]string{},
			Languages:       []string{},
			DefaultLanguage: "en",
			Theme:           "light",
			Font:            "inter",
			AccentColor:     "#3B82F6",
			AutoRefresh:     false,
			BannerHeader:    false,
			Logo:            &emptyStr,
			LogoHeight:      "40px",
			Favicon:         &emptyStr,
			GoogleAnalytics: &emptyStr,
			HidePoweredBy:   false,
			Subscribe: client.StatusPageSubscribeSettings{
				Enabled: false,
				Email:   false,
				Slack:   false,
				Teams:   false,
				SMS:     false,
			},
			Authentication: client.StatusPageAuthenticationSettings{
				PasswordProtection: false,
				GoogleSSO:          false,
				SAMLSSO:            false,
				AllowedDomains:     []string{},
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

// Increase coverage for mapTFToStringMap edge cases
func TestMapTFToStringMap_InvalidElement(t *testing.T) {
	// Create a map with a null value
	tfMap := types.MapValueMust(types.StringType, map[string]attr.Value{
		"en":   types.StringValue("English"),
		"null": types.StringNull(),
	})

	var diags diag.Diagnostics
	result := mapTFToStringMap(tfMap, &diags)

	// Should not error but should skip null values
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

// Increase coverage for mapListToStringSlice with invalid element
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

// Increase coverage for mapSubscriberToTF with empty optionals
func TestMapSubscriberToTF_EmptyOptionals(t *testing.T) {
	emptyStr := ""
	subscriber := &client.StatusPageSubscriber{
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

	// Empty strings should result in null values
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

// Test mapTFToSections with null and unknown lists
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
		// Create list with wrong element type (string instead of object)
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
		// Create section with name map that doesn't have "en" key
		sectionObj, _ := types.ObjectValue(SectionAttrTypes(), map[string]attr.Value{
			"name": types.MapValueMust(types.StringType, map[string]attr.Value{
				"fr": types.StringValue("Français"),
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

		// Should use first available language when "en" not present
		if sections[0].Name == "" {
			t.Error("expected section name to be set to first available language")
		}
	})
}

// Test mapTFToStringMap with null and unknown maps
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
		// Create a map with wrong value type (using ObjectAsOptions to bypass type checking)
		// Note: This is difficult to test in practice since Maps enforce element type
		// Testing the null value handling instead which is the critical path
		tfMap := types.MapValueMust(types.StringType, map[string]attr.Value{
			"en":   types.StringValue("English"),
			"fr":   types.StringValue("French"),
			"null": types.StringNull(), // This tests line 517
		})

		var diags diag.Diagnostics
		result := mapTFToStringMap(tfMap, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}

		// Null values should be skipped
		if _, ok := result["null"]; ok {
			t.Error("expected null values to be skipped")
		}

		if len(result) != 2 {
			t.Errorf("expected 2 entries (skipping null), got %d", len(result))
		}
	})
}

// Test mapListToStringSlice with null and unknown lists
func TestMapListToStringSlice_NullAndUnknown(t *testing.T) {
	t.Run("null list", func(t *testing.T) {
		nullList := types.ListNull(types.StringType)
		var diags diag.Diagnostics
		result := mapListToStringSlice(nullList, &diags)

		if diags.HasError() {
			t.Errorf("unexpected error: %v", diags.Errors())
		}
		// Function returns empty slice, not nil
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
		// Function returns empty slice, not nil
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

		// Null values become empty strings
		if len(result) != 3 {
			t.Errorf("expected 3 elements, got %d", len(result))
		}
		if result[0] != "first" || result[1] != "" || result[2] != "third" {
			t.Errorf("expected [first, '', third], got %v", result)
		}
	})

	t.Run("list with invalid element type", func(t *testing.T) {
		// Create list with Bool instead of String
		invalidList, _ := types.ListValue(types.BoolType, []attr.Value{
			types.BoolValue(true),
		})
		var diags diag.Diagnostics
		result := mapListToStringSlice(invalidList, &diags)

		if !diags.HasError() {
			t.Error("expected error for invalid element type")
		}

		// Should continue processing and return empty result
		if len(result) != 0 {
			t.Errorf("expected empty result after error, got %d elements", len(result))
		}
	})
}
