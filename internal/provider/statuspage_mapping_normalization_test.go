// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

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
