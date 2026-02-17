// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// makeStringPtr is a test helper that returns a pointer to a *string field.
func makeStringPtr() **string {
	var s *string
	return &s
}

// makeBoolPtr is a test helper that returns a pointer to a *bool field.
func makeBoolPtr() **bool {
	var b *bool
	return &b
}

// TestPopulateStringSettings verifies that all 9 string fields are populated
// when the attrs contain non-null values, and skipped when null.
func TestPopulateStringSettings(t *testing.T) {
	t.Run("all string fields populated when non-null", func(t *testing.T) {
		website := makeStringPtr()
		theme := makeStringPtr()
		font := makeStringPtr()
		accentColor := makeStringPtr()
		logo := makeStringPtr()
		logoHeight := makeStringPtr()
		favicon := makeStringPtr()
		googleAnalytics := makeStringPtr()
		defaultLanguage := makeStringPtr()

		target := &statusPageSettingsTarget{
			Website:         website,
			Theme:           theme,
			Font:            font,
			AccentColor:     accentColor,
			Logo:            logo,
			LogoHeight:      logoHeight,
			Favicon:         favicon,
			GoogleAnalytics: googleAnalytics,
			DefaultLanguage: defaultLanguage,
		}

		attrs := map[string]attr.Value{
			"website":          types.StringValue("https://example.com"),
			"theme":            types.StringValue("dark"),
			"font":             types.StringValue("Roboto"),
			"accent_color":     types.StringValue("#0066cc"),
			"logo":             types.StringValue("https://example.com/logo.png"),
			"logo_height":      types.StringValue("40px"),
			"favicon":          types.StringValue("https://example.com/favicon.ico"),
			"google_analytics": types.StringValue("UA-12345"),
			"default_language": types.StringValue("fr"),
		}

		populateStringSettings(attrs, target)

		checkStringField(t, "website", *target.Website, "https://example.com")
		checkStringField(t, "theme", *target.Theme, "dark")
		checkStringField(t, "font", *target.Font, "Roboto")
		checkStringField(t, "accent_color", *target.AccentColor, "#0066cc")
		checkStringField(t, "logo", *target.Logo, "https://example.com/logo.png")
		checkStringField(t, "logo_height", *target.LogoHeight, "40px")
		checkStringField(t, "favicon", *target.Favicon, "https://example.com/favicon.ico")
		checkStringField(t, "google_analytics", *target.GoogleAnalytics, "UA-12345")
		checkStringField(t, "default_language", *target.DefaultLanguage, "fr")
	})

	t.Run("string fields skipped when null", func(t *testing.T) {
		website := makeStringPtr()
		theme := makeStringPtr()
		target := &statusPageSettingsTarget{
			Website: website,
			Theme:   theme,
		}

		attrs := map[string]attr.Value{
			"website": types.StringNull(),
			"theme":   types.StringNull(),
		}

		populateStringSettings(attrs, target)

		if *target.Website != nil {
			t.Errorf("expected website to be nil when null, got %v", *target.Website)
		}
		if *target.Theme != nil {
			t.Errorf("expected theme to be nil when null, got %v", *target.Theme)
		}
	})

	t.Run("string fields skipped when key missing from attrs", func(t *testing.T) {
		website := makeStringPtr()
		target := &statusPageSettingsTarget{
			Website: website,
		}

		// Empty attrs — no keys present
		attrs := map[string]attr.Value{}

		populateStringSettings(attrs, target)

		if *target.Website != nil {
			t.Errorf("expected website to be nil when key missing, got %v", *target.Website)
		}
	})
}

// checkStringField is a test helper that checks a *string pointer equals the expected value.
func checkStringField(t *testing.T, name string, got *string, want string) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: expected %q but got nil", name, want)
		return
	}
	if *got != want {
		t.Errorf("%s: expected %q but got %q", name, want, *got)
	}
}

// TestPopulateBoolSettings verifies that all 4 bool fields are populated
// when attrs contain non-null values, and skipped when null.
func TestPopulateBoolSettings(t *testing.T) {
	t.Run("all bool fields populated when non-null", func(t *testing.T) {
		autoRefresh := makeBoolPtr()
		bannerHeader := makeBoolPtr()
		hidePoweredBy := makeBoolPtr()
		hideFromSearchEngines := makeBoolPtr()

		target := &statusPageSettingsTarget{
			AutoRefresh:           autoRefresh,
			BannerHeader:          bannerHeader,
			HidePoweredBy:         hidePoweredBy,
			HideFromSearchEngines: hideFromSearchEngines,
		}

		attrs := map[string]attr.Value{
			"auto_refresh":             types.BoolValue(true),
			"banner_header":            types.BoolValue(false),
			"hide_powered_by":          types.BoolValue(true),
			"hide_from_search_engines": types.BoolValue(false),
		}

		populateBoolSettings(attrs, target)

		checkBoolField(t, "auto_refresh", *target.AutoRefresh, true)
		checkBoolField(t, "banner_header", *target.BannerHeader, false)
		checkBoolField(t, "hide_powered_by", *target.HidePoweredBy, true)
		checkBoolField(t, "hide_from_search_engines", *target.HideFromSearchEngines, false)
	})

	t.Run("bool fields skipped when null", func(t *testing.T) {
		autoRefresh := makeBoolPtr()
		bannerHeader := makeBoolPtr()
		hidePoweredBy := makeBoolPtr()
		hideFromSearchEngines := makeBoolPtr()

		target := &statusPageSettingsTarget{
			AutoRefresh:           autoRefresh,
			BannerHeader:          bannerHeader,
			HidePoweredBy:         hidePoweredBy,
			HideFromSearchEngines: hideFromSearchEngines,
		}

		attrs := map[string]attr.Value{
			"auto_refresh":             types.BoolNull(),
			"banner_header":            types.BoolNull(),
			"hide_powered_by":          types.BoolNull(),
			"hide_from_search_engines": types.BoolNull(),
		}

		populateBoolSettings(attrs, target)

		if *target.AutoRefresh != nil {
			t.Errorf("expected auto_refresh to be nil when null, got %v", *target.AutoRefresh)
		}
		if *target.BannerHeader != nil {
			t.Errorf("expected banner_header to be nil when null, got %v", *target.BannerHeader)
		}
		if *target.HidePoweredBy != nil {
			t.Errorf("expected hide_powered_by to be nil when null, got %v", *target.HidePoweredBy)
		}
		if *target.HideFromSearchEngines != nil {
			t.Errorf("expected hide_from_search_engines to be nil when null, got %v", *target.HideFromSearchEngines)
		}
	})

	t.Run("bool fields skipped when key missing from attrs", func(t *testing.T) {
		autoRefresh := makeBoolPtr()
		target := &statusPageSettingsTarget{
			AutoRefresh: autoRefresh,
		}

		attrs := map[string]attr.Value{}

		populateBoolSettings(attrs, target)

		if *target.AutoRefresh != nil {
			t.Errorf("expected auto_refresh to be nil when key missing, got %v", *target.AutoRefresh)
		}
	})
}

// checkBoolField is a test helper that checks a *bool pointer equals the expected value.
func checkBoolField(t *testing.T, name string, got *bool, want bool) {
	t.Helper()
	if got == nil {
		t.Errorf("%s: expected %v but got nil", name, want)
		return
	}
	if *got != want {
		t.Errorf("%s: expected %v but got %v", name, want, *got)
	}
}

// TestPopulateCollectionSettings verifies that description map and languages list are populated correctly.
func TestPopulateCollectionSettings(t *testing.T) {
	t.Run("description map populated when non-null", func(t *testing.T) {
		var desc map[string]string
		var langs []string

		target := &statusPageSettingsTarget{
			Description: &desc,
			Languages:   &langs,
		}

		descMap, _ := types.MapValue(types.StringType, map[string]attr.Value{
			"en": types.StringValue("English description"),
			"fr": types.StringValue("Description française"),
		})

		attrs := map[string]attr.Value{
			"description": descMap,
			"languages":   types.ListNull(types.StringType),
		}

		var diags diag.Diagnostics
		populateCollectionSettings(attrs, target, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}
		if len(desc) != 2 {
			t.Errorf("expected 2 description entries, got %d", len(desc))
		}
		if desc["en"] != "English description" {
			t.Errorf("expected en=%q, got %q", "English description", desc["en"])
		}
		if desc["fr"] != "Description française" {
			t.Errorf("expected fr=%q, got %q", "Description française", desc["fr"])
		}
	})

	t.Run("languages list populated when non-null", func(t *testing.T) {
		var desc map[string]string
		var langs []string

		target := &statusPageSettingsTarget{
			Description: &desc,
			Languages:   &langs,
		}

		langsList, _ := types.ListValue(types.StringType, []attr.Value{
			types.StringValue("en"),
			types.StringValue("fr"),
			types.StringValue("de"),
		})

		attrs := map[string]attr.Value{
			"description": types.MapNull(types.StringType),
			"languages":   langsList,
		}

		var diags diag.Diagnostics
		populateCollectionSettings(attrs, target, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}
		if len(langs) != 3 {
			t.Errorf("expected 3 languages, got %d", len(langs))
		}
	})

	t.Run("empty description map not assigned", func(t *testing.T) {
		var desc map[string]string
		var langs []string

		target := &statusPageSettingsTarget{
			Description: &desc,
			Languages:   &langs,
		}

		emptyMap, _ := types.MapValue(types.StringType, map[string]attr.Value{})

		attrs := map[string]attr.Value{
			"description": emptyMap,
			"languages":   types.ListNull(types.StringType),
		}

		var diags diag.Diagnostics
		populateCollectionSettings(attrs, target, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}
		// Empty map should not be assigned (len check guards it)
		if desc != nil {
			t.Errorf("expected description to remain nil for empty map, got %v", desc)
		}
	})

	t.Run("empty languages list not assigned", func(t *testing.T) {
		var desc map[string]string
		var langs []string

		target := &statusPageSettingsTarget{
			Description: &desc,
			Languages:   &langs,
		}

		emptyList, _ := types.ListValue(types.StringType, []attr.Value{})

		attrs := map[string]attr.Value{
			"description": types.MapNull(types.StringType),
			"languages":   emptyList,
		}

		var diags diag.Diagnostics
		populateCollectionSettings(attrs, target, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}
		// Empty list should not be assigned (len check guards it)
		if langs != nil {
			t.Errorf("expected languages to remain nil for empty list, got %v", langs)
		}
	})

	t.Run("both null does not modify target", func(t *testing.T) {
		var desc map[string]string
		var langs []string

		target := &statusPageSettingsTarget{
			Description: &desc,
			Languages:   &langs,
		}

		attrs := map[string]attr.Value{
			"description": types.MapNull(types.StringType),
			"languages":   types.ListNull(types.StringType),
		}

		var diags diag.Diagnostics
		populateCollectionSettings(attrs, target, &diags)

		if diags.HasError() {
			t.Fatalf("unexpected error: %v", diags.Errors())
		}
		if desc != nil {
			t.Errorf("expected description nil, got %v", desc)
		}
		if langs != nil {
			t.Errorf("expected languages nil, got %v", langs)
		}
	})
}
