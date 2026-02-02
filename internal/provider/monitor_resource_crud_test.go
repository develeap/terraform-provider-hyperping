// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMonitorResource_basic(t *testing.T) {
	// Create mock server
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create and Read testing
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "test-monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "test-monitor"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "https://example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "GET"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "60"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "2xx"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "follow_redirects", "true"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "paused", "false"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update and Read testing
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "updated-monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "updated-monitor"),
				),
			},
		},
	})
}

func TestAccMonitorResource_full(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigFull(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "full-monitor"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "https://api.example.com/health"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "POST"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "300"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "201"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "follow_redirects", "false"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "paused", "false"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_headers.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_body", `{"key":"value"}`),
				),
			},
		},
	})
}

func TestAccMonitorResource_paused(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigPaused(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "paused-monitor"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "paused", "true"),
				),
			},
			// Unpause
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "paused-monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "paused", "false"),
				),
			},
		},
	})
}

func TestAccMonitorResource_disappears(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "disappear-monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
					// Simulate resource being deleted outside Terraform
					testAccCheckMonitorDisappears(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMonitorResource_updateAllFields(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with minimal config
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "update-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "update-test"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "GET"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "60"),
				),
			},
			// Update all fields
			{
				Config: testAccMonitorResourceConfigUpdateAll(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "updated-all-fields"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "https://updated.example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "PUT"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "120"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "204"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "follow_redirects", "false"),
				),
			},
		},
	})
}

func TestAccMonitorResource_headersUpdate(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with headers
			{
				Config: testAccMonitorResourceConfigWithHeaders(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_headers.#", "2"),
				),
			},
			// Update headers
			{
				Config: testAccMonitorResourceConfigWithUpdatedHeaders(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_headers.#", "1"),
				),
			},
		},
	})
}

func TestAccMonitorResource_bodyUpdate(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with body
			{
				Config: testAccMonitorResourceConfigWithBody(server.URL, `{"initial":"data"}`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_body", `{"initial":"data"}`),
				),
			},
			// Update body
			{
				Config: testAccMonitorResourceConfigWithBody(server.URL, `{"updated":"payload"}`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_body", `{"updated":"payload"}`),
				),
			},
		},
	})
}

func TestAccMonitorResource_regionsUpdate(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with specific regions
			{
				Config: testAccMonitorResourceConfigWithRegions(server.URL, `["london", "virginia"]`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "2"),
				),
			},
			// Update regions
			{
				Config: testAccMonitorResourceConfigWithRegions(server.URL, `["frankfurt", "singapore", "tokyo"]`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "3"),
				),
			},
		},
	})
}

func TestAccMonitorResource_pauseUnpause(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create unpaused
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "pause-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "paused", "false"),
				),
			},
			// Pause via update
			{
				Config: testAccMonitorResourceConfigWithPaused(server.URL, "pause-test", true),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "paused", "true"),
				),
			},
			// Unpause via update
			{
				Config: testAccMonitorResourceConfigWithPaused(server.URL, "pause-test", false),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "paused", "false"),
				),
			},
		},
	})
}

func TestAccMonitorResource_removeOptionalFields(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with all optional fields
			{
				Config: testAccMonitorResourceConfigAllOptional(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_headers.#", "1"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_body", `{"test":"data"}`),
				),
			},
			// Remove all optional fields
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "remove-optional-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "remove-optional-test"),
				),
			},
		},
	})
}

func TestAccMonitorResource_clearRegions(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with regions
			{
				Config: testAccMonitorResourceConfigWithRegions(server.URL, `["london", "virginia"]`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "2"),
				),
			},
			// Clear regions
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "regions-test"),
			},
		},
	})
}

func TestAccMonitorResource_clearHeaders(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with headers
			{
				Config: testAccMonitorResourceConfigWithHeaders(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_headers.#", "2"),
				),
			},
			// Clear headers
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "headers-test"),
			},
		},
	})
}

func TestAccMonitorResource_clearBody(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with body
			{
				Config: testAccMonitorResourceConfigWithBody(server.URL, `{"initial":"data"}`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_body", `{"initial":"data"}`),
				),
			},
			// Clear body
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "body-test"),
			},
		},
	})
}
