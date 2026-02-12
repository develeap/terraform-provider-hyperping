// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHealthcheckResource_import(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Step 1: Create the resource
			{
				Config: testAccHealthcheckResourceConfig_basic(server.URL, "test-import"),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "name", "test-import"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_value", "60"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_value", "300"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_type", "seconds"),
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "ping_url"),
				),
			},
			// Step 2: Import it and verify all fields match
			{
				ResourceName:      "hyperping_healthcheck.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHealthcheckResource_importFull(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource with all optional fields
			{
				Config: testAccHealthcheckResourceConfig_full(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "name", "full-healthcheck"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_value", "300"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_value", "600"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "grace_period_type", "seconds"),
					tfresource.TestCheckResourceAttr("hyperping_healthcheck.test", "is_paused", "false"),
					tfresource.TestCheckResourceAttrSet("hyperping_healthcheck.test", "escalation_policy"),
				),
			},
			// Import and verify zero-drift
			{
				ResourceName:      "hyperping_healthcheck.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccHealthcheckResource_importNotFound(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource first
			{
				Config: testAccHealthcheckResourceConfig_basic(server.URL, "test"),
			},
			// Try to import non-existent resource
			{
				ResourceName:  "hyperping_healthcheck.test",
				ImportState:   true,
				ImportStateId: "hc_nonexistent",
				ExpectError:   regexp.MustCompile("Healthcheck not found|Cannot import non-existent"),
			},
		},
	})
}
