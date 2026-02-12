// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMonitorResource_import(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Step 1: Create the resource
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "test-import"),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "test-import"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "https://example.com"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
				),
			},
			// Step 2: Import it and verify all fields match
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMonitorResource_importFull(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource with all optional fields
			{
				Config: testAccMonitorResourceConfigFull(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "full-monitor"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "https://api.example.com/health"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "POST"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "300"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "201"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "follow_redirects", "false"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_headers.#", "2"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "request_body"),
				),
			},
			// Import and verify zero-drift (all fields should match exactly)
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccMonitorResource_importNotFound(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource first
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "test"),
			},
			// Try to import non-existent resource
			{
				ResourceName:  "hyperping_monitor.test",
				ImportState:   true,
				ImportStateId: "mon_nonexistent",
				ExpectError:   regexp.MustCompile("Monitor not found|Cannot import non-existent"),
			},
		},
	})
}
