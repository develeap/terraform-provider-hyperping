// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
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
				ImportStateVerifyIgnore: []string{
					"settings.description.%",
					"settings.description.fr",
					"settings.description.de",
					"settings.description.ru",
					"settings.description.nl",
				},
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
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.description.en", "Production system status"),
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
