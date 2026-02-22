// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccStatusPageResource_import(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Step 1: Create the resource
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "name", "Test Status Page"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "hosted_subdomain", "test-status"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage.test", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_statuspage.test", "url"),
				),
			},
			// Step 2: Import it and verify all fields match
			{
				ResourceName:      "hyperping_statuspage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStatusPageResource_importFull(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource with all optional fields
			{
				Config: testAccStatusPageResourceConfig_full(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "name", "Production Status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "hosted_subdomain", "prod-status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.theme", "dark"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.font", "Inter"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.accent_color", "#0066cc"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.languages.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.description", "Production system status"),
					tfresource.TestCheckResourceAttr("hyperping_statuspage.test", "settings.subscribe.enabled", "true"),
				),
			},
			// Import and verify zero-drift (with ignored fields)
			{
				ResourceName:      "hyperping_statuspage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccStatusPageResource_importNotFound(t *testing.T) {
	server := newMockStatusPageServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource first
			{
				Config: testAccStatusPageResourceConfig_basic(server.URL),
			},
			// Try to import non-existent resource
			{
				ResourceName:  "hyperping_statuspage.test",
				ImportState:   true,
				ImportStateId: "sp_nonexistent",
				ExpectError:   regexp.MustCompile("Status page not found|Cannot import non-existent"),
			},
		},
	})
}
