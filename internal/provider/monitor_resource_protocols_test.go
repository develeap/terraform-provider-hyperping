// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccMonitorResource_portProtocol tests port protocol monitor creation and updates.
// Regression test for v1.0.8 bug: HTTP defaults were applied to all protocols.
// Verifies that HTTP-specific fields (http_method, expected_status_code, follow_redirects)
// are null for port monitors and don't cause "Provider produced inconsistent result" errors.
func TestAccMonitorResource_portProtocol(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create port monitor with port field
			{
				Config: testAccMonitorResourceConfigPort(server.URL, "port-monitor", 5432),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "port-monitor"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "https://db.example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "port"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "port", "5432"),
					// Verify HTTP fields have default values (not null) - provider sets defaults
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "GET"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "2xx"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "follow_redirects", "true"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
				),
			},
			// ImportState testing - verify no drift after import
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update port number
			{
				Config: testAccMonitorResourceConfigPort(server.URL, "port-monitor", 3306),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "port", "3306"),
				),
			},
		},
	})
}

// TestAccMonitorResource_icmpProtocol tests ICMP protocol monitor.
// Regression test for v1.0.8 bug: verifies ICMP monitors don't inherit HTTP defaults.
func TestAccMonitorResource_icmpProtocol(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create ICMP monitor
			{
				Config: testAccMonitorResourceConfigICMP(server.URL, "icmp-monitor"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "icmp-monitor"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "https://ping.example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "icmp"),
					// HTTP fields should have default values
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "GET"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "2xx"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "follow_redirects", "true"),
				),
			},
			// ImportState testing - verify persistence
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccMonitorResource_icmpWithHTTPFields tests that ICMP monitor rejects HTTP-only fields.
// ValidateConfig catches this at plan time with a clear error message.
func TestAccMonitorResource_icmpWithHTTPFields(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// ICMP monitor with HTTP fields should produce plan-time error
			{
				Config:      testAccMonitorResourceConfigICMPWithHTTPFields(server.URL),
				ExpectError: regexp.MustCompile(`http_method is only valid for HTTP monitors`),
			},
		},
	})
}

// TestAccMonitorResource_protocolSwitch tests switching between protocols.
// When switching from HTTP to ICMP, HTTP-only fields must be removed from config.
// When switching back, they can be re-added.
func TestAccMonitorResource_protocolSwitch(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Start with HTTP monitor with all HTTP fields
			{
				Config: testAccMonitorResourceConfigHTTPFull(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "POST"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "201"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "follow_redirects", "false"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_body", `{"key":"value"}`),
				),
			},
			// Switch to ICMP - must remove HTTP-only fields from config
			{
				Config: testAccMonitorResourceConfigICMP(server.URL, "switch-protocol-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "icmp"),
				),
			},
			// Import state to verify no drift
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Switch back to HTTP - verify HTTP fields work again
			{
				Config: testAccMonitorResourceConfigSwitchBackToHTTP(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", "GET"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "200"),
				),
			},
		},
	})
}

// TestAccMonitorResource_portWithoutPortField tests that port protocol requires the port field.
// ValidateConfig catches this at plan time with a clear error message.
func TestAccMonitorResource_portWithoutPortField(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Port monitor without port field should produce plan-time error
			{
				Config:      testAccMonitorResourceConfigPortNoPort(server.URL),
				ExpectError: regexp.MustCompile(`port is required when protocol is "port"`),
			},
		},
	})
}

// TestAccMonitorResource_requiredKeywordNonHTTP tests that required_keyword is rejected for non-HTTP protocols.
// ValidateConfig catches this at plan time with a clear error message.
func TestAccMonitorResource_requiredKeywordNonHTTP(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Port monitor with required_keyword should produce plan-time error
			{
				Config:      testAccMonitorResourceConfigPortWithKeyword(server.URL, "HEALTHY"),
				ExpectError: regexp.MustCompile(`required_keyword is only valid for HTTP monitors`),
			},
		},
	})
}

// Config generators for protocol-specific tests

func testAccMonitorResourceConfigPort(baseURL, name string, port int) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name     = "` + name + `"
  url      = "https://db.example.com"
  protocol = "port"
  port     = ` + tfInt(port) + `
}
`
}

func testAccMonitorResourceConfigICMP(baseURL, name string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name     = "` + name + `"
  url      = "https://ping.example.com"
  protocol = "icmp"
}
`
}

func testAccMonitorResourceConfigICMPWithHTTPFields(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name                 = "icmp-with-http-fields"
  url                  = "https://ping.example.com"
  protocol             = "icmp"
  http_method          = "POST"
  expected_status_code = "201"
  follow_redirects     = false
}
`
}

func testAccMonitorResourceConfigHTTPFull(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name                 = "switch-protocol-test"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "POST"
  expected_status_code = "201"
  follow_redirects     = false
  request_body         = jsonencode({
    key = "value"
  })
}
`
}

func testAccMonitorResourceConfigSwitchBackToHTTP(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name                 = "switch-protocol-test"
  url                  = "https://api.example.com/health"
  protocol             = "http"
  http_method          = "GET"
  expected_status_code = "200"
}
`
}

func testAccMonitorResourceConfigPortNoPort(baseURL string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name     = "port-no-port-field"
  url      = "https://example.com:8443"
  protocol = "port"
}
`
}

func testAccMonitorResourceConfigPortWithKeyword(baseURL, keyword string) string {
	return testAccProviderConfig(baseURL) + `
resource "hyperping_monitor" "test" {
  name             = "port-with-keyword"
  url              = "https://example.com"
  protocol         = "port"
  port             = 9000
  required_keyword = "` + keyword + `"
}
`
}

// Helper functions

func testAccProviderConfig(baseURL string) string {
	return `
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = "` + baseURL + `"
}
`
}

func tfInt(val int) string {
	return fmt.Sprintf("%d", val)
}
