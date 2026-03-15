// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusPageResource_basic(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "name", "Test Status Page"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "hosted_subdomain", "test-status"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage.test", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage.test", "url"),
				),
			},
			{
				ResourceName:      "hyperping_statuspage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStatusPageResource_full(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_full(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "name", "Production Status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "hosted_subdomain", "prod-status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "dark"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.font", "Inter"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.accent_color", "#0066cc"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.languages.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.description", "Production system status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.enabled", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.email", "true"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.authentication.password_protection", "false"),
				),
			},
		},
	})
}

func TestAccStatusPageResource_withSections(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with sections
			{
				Config: testAccStatusPageResourceConfig_withSections(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.name.en", "API Services"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.1.name.en", "Databases"),
				),
			},
			// Update sections
			{
				Config: testAccStatusPageResourceConfig_withUpdatedSections(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "sections.0.name.en", "All Services"),
				),
			},
		},
	})
}

func TestAccStatusPageResource_updateSettings(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with default settings
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "system"),
				),
			},
			// Update theme
			{
				Config: testAccStatusPageResourceConfig_withTheme(server.URL, "dark"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "dark"),
				),
			},
			// Update to light theme
			{
				Config: testAccStatusPageResourceConfig_withTheme(server.URL, "light"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "light"),
				),
			},
		},
	})
}

func TestAccStatusPageResource_disappears(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					testAccCheckStatusPageDisappears(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// T19: StatusPage languages rejects invalid language code
func TestAccStatusPageResource_invalidLanguageCode(t *testing.T) {
	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccStatusPageConfig_invalidLanguage("http://localhost:9999"),
				ExpectError: regexp.MustCompile(`(?i)value must be one of`),
			},
		},
	})
}

// T20: StatusPage default_language rejects invalid value
func TestAccStatusPageResource_invalidDefaultLanguage(t *testing.T) {
	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccStatusPageConfig_invalidDefaultLanguage("http://localhost:9999"),
				ExpectError: regexp.MustCompile(`(?i)value must be one of`),
			},
		},
	})
}

// T21: StatusPage languages accepts all valid language codes
func TestAccStatusPageResource_validLanguageCodes(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccStatusPageConfig_multipleValidLanguages(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.languages.#", "3"),
				),
			},
		},
	})
}

// Config helpers for language validation tests

func testAccStatusPageConfig_invalidLanguage(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Language Test"
  hosted_subdomain = "lang-test"

  settings = {
    name      = "Language Test"
    languages = ["en", "xx"]
  }
}
`, baseURL)
}

func testAccStatusPageConfig_invalidDefaultLanguage(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Default Language Test"
  hosted_subdomain = "default-lang-test"

  settings = {
    name             = "Default Language Test"
    languages        = ["en"]
    default_language = "xx"
  }
}
`, baseURL)
}

func testAccStatusPageConfig_multipleValidLanguages(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_statuspage" "test" {
  name             = "Multi Language Test"
  hosted_subdomain = "multi-lang-test"

  settings = {
    name      = "Multi Language Test"
    languages = ["en", "fr", "de"]
  }
}
`, baseURL)
}
