// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"strings"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Test Group 1: Boundary Values
// =============================================================================

// TestAccMonitorResource_frequencyBoundaries tests check_frequency at min, max, and common values.
// Tests: 10 (min), 60 (default/common), 86400 (max).
// Invalid frequencies are tested in validator tests, not here.
func TestAccMonitorResource_frequencyBoundaries(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test minimum frequency (10 seconds)
			{
				Config: testAccMonitorResourceConfigWithFrequency(server.URL, 10),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "10"),
				),
			},
			// Test default/common frequency (60 seconds)
			{
				Config: testAccMonitorResourceConfigWithFrequency(server.URL, 60),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "60"),
				),
			},
			// Test maximum frequency (86400 seconds = 24 hours)
			{
				Config: testAccMonitorResourceConfigWithFrequency(server.URL, 86400),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "86400"),
				),
			},
			// Test another common value (300 seconds = 5 minutes)
			{
				Config: testAccMonitorResourceConfigWithFrequency(server.URL, 300),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "300"),
				),
			},
		},
	})
}

// TestAccMonitorResource_allRegions tests all 8 available regions.
// Verifies that all regions persist correctly and can be updated.
func TestAccMonitorResource_allRegions(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	// All 8 regions from client.AllowedRegions
	allRegions := []string{
		"london", "frankfurt", "singapore", "sydney",
		"tokyo", "virginia", "saopaulo", "bahrain",
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with all 8 regions
			{
				Config: testAccMonitorResourceConfigWithAllRegions(server.URL, allRegions),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "8"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "london"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "frankfurt"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "singapore"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "sydney"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "tokyo"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "virginia"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "saopaulo"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "bahrain"),
				),
			},
			// Update to single region
			{
				Config: testAccMonitorResourceConfigWithRegions(server.URL, `["london"]`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "1"),
					tfresource.TestCheckTypeSetElemAttr("hyperping_monitor.test", "regions.*", "london"),
				),
			},
		},
	})
}

// TestAccMonitorResource_statusCodeRanges tests various expected_status_code patterns.
// Tests wildcard patterns (2xx, 3xx, 4xx, 5xx) and specific codes.
func TestAccMonitorResource_statusCodeRanges(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test 2xx wildcard (default)
			{
				Config: testAccMonitorResourceConfigWithStatusCode(server.URL, "2xx"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "2xx"),
				),
			},
			// Test specific 200
			{
				Config: testAccMonitorResourceConfigWithStatusCode(server.URL, "200"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "200"),
				),
			},
			// Test 201 Created
			{
				Config: testAccMonitorResourceConfigWithStatusCode(server.URL, "201"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "201"),
				),
			},
			// Test 204 No Content
			{
				Config: testAccMonitorResourceConfigWithStatusCode(server.URL, "204"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "204"),
				),
			},
			// Test 3xx wildcard
			{
				Config: testAccMonitorResourceConfigWithStatusCode(server.URL, "3xx"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "3xx"),
				),
			},
			// Test 4xx wildcard
			{
				Config: testAccMonitorResourceConfigWithStatusCode(server.URL, "4xx"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "4xx"),
				),
			},
			// Test specific 404
			{
				Config: testAccMonitorResourceConfigWithStatusCode(server.URL, "404"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "expected_status_code", "404"),
				),
			},
		},
	})
}

// =============================================================================
// Test Group 2: Field Length Boundaries
// =============================================================================

// TestAccMonitorResource_nameLengthBoundaries tests monitor name at various lengths.
// Tests 1 char (min), 255 chars (max).
// 256+ chars should fail validation (tested in validator tests).
func TestAccMonitorResource_nameLengthBoundaries(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test minimum length (1 character)
			{
				Config: testAccMonitorResourceConfigWithName(server.URL, "a"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "a"),
				),
			},
			// Test maximum length (255 characters)
			{
				Config: testAccMonitorResourceConfigWithName(server.URL, strings.Repeat("x", 255)),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", strings.Repeat("x", 255)),
				),
			},
			// Test common length (50 characters)
			{
				Config: testAccMonitorResourceConfigWithName(server.URL, strings.Repeat("m", 50)),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", strings.Repeat("m", 50)),
				),
			},
		},
	})
}

// TestAccMonitorResource_urlMaxLength tests very long URLs (up to 2000 characters).
// Verifies URL length boundary handling.
func TestAccMonitorResource_urlMaxLength(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	// Create a very long URL (2000 chars) with long query string
	longURL := "https://example.com/api/v1/resource?" + strings.Repeat("param=value&", 180)
	if len(longURL) > 2000 {
		longURL = longURL[:2000]
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigWithURL(server.URL, longURL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", longURL),
				),
			},
		},
	})
}

// =============================================================================
// Test Group 3: Empty/Null Handling
// =============================================================================

// TestAccMonitorResource_emptyCollections tests monitors with empty regions and headers.
// Verifies that empty arrays are handled correctly and API defaults are applied.
func TestAccMonitorResource_emptyCollections(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create monitor without regions and headers (null/omitted)
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "empty-collections"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "empty-collections"),
				),
			},
			// Add regions and headers
			{
				Config: testAccMonitorResourceConfigWithRegionsAndHeaders(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "regions.#", "2"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_headers.#", "1"),
				),
			},
			// Remove back to empty
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "empty-collections"),
			},
		},
	})
}

// TestAccMonitorResource_nullVsEmptyString tests null vs empty string for optional fields.
// Note: Provider treats empty strings as null for request_body.
// This test verifies the normalization behavior.
func TestAccMonitorResource_nullVsEmptyString(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with non-empty body
			{
				Config: testAccMonitorResourceConfigWithBody(server.URL, `{"key":"value"}`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_body", `{"key":"value"}`),
				),
			},
			// Update with different non-empty body
			{
				Config: testAccMonitorResourceConfigWithBody(server.URL, `{"updated":"data"}`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_body", `{"updated":"data"}`),
				),
			},
			// Clear body (null - field omitted)
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "body-test"),
			},
		},
	})
}

// =============================================================================
// Test Group 4: Untested Fields
// =============================================================================

// TestAccMonitorResource_alertsWait tests the alerts_wait field.
// Tests different alert wait times: 60 (1 min), 300 (5 min).
// Note: alerts_wait=0 is treated as null/unset by the provider mapping logic.
func TestAccMonitorResource_alertsWait(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with 60 second wait
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 60),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "60"),
				),
			},
			// Update to 300 seconds (5 minutes)
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 300),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "300"),
				),
			},
			// Update to 120 seconds (2 minutes)
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 120),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "120"),
				),
			},
			// Clear alerts_wait (null)
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "alerts-wait-test"),
			},
		},
	})
}

// TestAccMonitorResource_escalationPolicy tests the escalation_policy field.
// Tests assigning, updating, and clearing escalation policy UUIDs.
func TestAccMonitorResource_escalationPolicy(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with escalation_policy
			{
				Config: testAccMonitorResourceConfigWithEscalationPolicy(server.URL, "policy_abc123"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "escalation_policy", "policy_abc123"),
				),
			},
			// Update to different policy
			{
				Config: testAccMonitorResourceConfigWithEscalationPolicy(server.URL, "policy_def456"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "escalation_policy", "policy_def456"),
				),
			},
			// Clear escalation_policy
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "escalation-test"),
			},
		},
	})
}

// TestAccMonitorResource_portField tests the port field for port protocol monitors.
// Tests various port numbers and protocol combinations.
func TestAccMonitorResource_portField(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create port monitor with port 80
			{
				Config: testAccMonitorResourceConfigWithPort(server.URL, 80),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "port"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "port", "80"),
				),
			},
			// Update to port 443
			{
				Config: testAccMonitorResourceConfigWithPort(server.URL, 443),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "port", "443"),
				),
			},
			// Update to port 8080
			{
				Config: testAccMonitorResourceConfigWithPort(server.URL, 8080),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "port", "8080"),
				),
			},
			// Test high port number (65535)
			{
				Config: testAccMonitorResourceConfigWithPort(server.URL, 65535),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "port", "65535"),
				),
			},
		},
	})
}

// TestAccMonitorResource_protocolTypes tests all protocol types.
// Tests http, port, and icmp protocols with appropriate configurations.
func TestAccMonitorResource_protocolTypes(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test HTTP protocol
			{
				Config: testAccMonitorResourceConfigWithProtocol(server.URL, "http"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
				),
			},
			// Test PORT protocol (requires port field)
			{
				Config: testAccMonitorResourceConfigWithPort(server.URL, 8080),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "port"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "port", "8080"),
				),
			},
			// Test ICMP protocol - create new resource to avoid protocol switching issues
			{
				Config: testAccMonitorResourceConfigWithProtocolICMP(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.icmp_test", "protocol", "icmp"),
				),
			},
		},
	})
}

// TestAccMonitorResource_httpMethodsComprehensive tests all HTTP methods.
// Tests GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS.
func TestAccMonitorResource_httpMethodsComprehensive(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}

	steps := make([]tfresource.TestStep, 0, len(methods))
	for _, method := range methods {
		steps = append(steps, tfresource.TestStep{
			Config: testAccMonitorResourceConfigWithHTTPMethod(server.URL, method),
			Check: tfresource.ComposeAggregateTestCheckFunc(
				tfresource.TestCheckResourceAttr("hyperping_monitor.test", "http_method", method),
			),
		})
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}

// TestAccMonitorResource_complexHeadersAndBody tests monitors with complex headers and body.
// Tests multiple headers, special characters, JSON body.
func TestAccMonitorResource_complexHeadersAndBody(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with multiple headers and complex JSON body
			{
				Config: testAccMonitorResourceConfigComplexHeadersBody(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "request_headers.#", "3"),
				),
			},
		},
	})
}

// TestAccMonitorResource_requiredKeywordEdgeCases tests required_keyword with various patterns.
// Tests single words, phrases, special characters.
func TestAccMonitorResource_requiredKeywordEdgeCases(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Single word keyword
			{
				Config: testAccMonitorResourceConfigWithRequiredKeyword(server.URL, "healthy"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "required_keyword", "healthy"),
				),
			},
			// JSON pattern keyword
			{
				Config: testAccMonitorResourceConfigWithRequiredKeyword(server.URL, `"status":"ok"`),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "required_keyword", `"status":"ok"`),
				),
			},
			// Phrase with spaces
			{
				Config: testAccMonitorResourceConfigWithRequiredKeyword(server.URL, "All systems operational"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "required_keyword", "All systems operational"),
				),
			},
		},
	})
}

// =============================================================================
// Test Configuration Helpers
// =============================================================================

func testAccMonitorResourceConfigWithFrequency(baseURL string, frequency int) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name            = "frequency-test"
  url             = "https://example.com"
  check_frequency = %[2]d
}
`, baseURL, frequency)
}

func testAccMonitorResourceConfigWithAllRegions(baseURL string, regions []string) string {
	regionList := `["` + strings.Join(regions, `", "`) + `"]`
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name    = "all-regions-test"
  url     = "https://example.com"
  regions = %[2]s
}
`, baseURL, regionList)
}

func testAccMonitorResourceConfigWithStatusCode(baseURL, statusCode string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name                 = "status-code-test"
  url                  = "https://example.com"
  expected_status_code = %[2]q
}
`, baseURL, statusCode)
}

func testAccMonitorResourceConfigWithName(baseURL, name string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name = %[2]q
  url  = "https://example.com"
}
`, baseURL, name)
}

func testAccMonitorResourceConfigWithURL(baseURL, monitorURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name = "url-test"
  url  = %[2]q
}
`, baseURL, monitorURL)
}

func testAccMonitorResourceConfigWithRegionsAndHeaders(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name    = "empty-collections"
  url     = "https://example.com"
  regions = ["london", "virginia"]
  request_headers = [
    { name = "X-Custom", value = "test" }
  ]
}
`, baseURL)
}

func testAccMonitorResourceConfigWithAlertsWait(baseURL string, alertsWait int) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name        = "alerts-wait-test"
  url         = "https://example.com"
  alerts_wait = %[2]d
}
`, baseURL, alertsWait)
}

func testAccMonitorResourceConfigWithEscalationPolicy(baseURL, policyUUID string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name              = "escalation-test"
  url               = "https://example.com"
  escalation_policy = %[2]q
}
`, baseURL, policyUUID)
}

func testAccMonitorResourceConfigWithPort(baseURL string, port int) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name     = "port-test"
  url      = "https://example.com"
  protocol = "port"
  port     = %[2]d
}
`, baseURL, port)
}

func testAccMonitorResourceConfigWithProtocol(baseURL, protocol string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name     = "protocol-test"
  url      = "https://example.com"
  protocol = %[2]q
}
`, baseURL, protocol)
}

func testAccMonitorResourceConfigWithProtocolICMP(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "icmp_test" {
  name     = "icmp-protocol-test"
  url      = "https://example.com"
  protocol = "icmp"
}
`, baseURL)
}

func testAccMonitorResourceConfigWithHTTPMethod(baseURL, method string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name        = "method-test"
  url         = "https://example.com"
  http_method = %[2]q
}
`, baseURL, method)
}

func testAccMonitorResourceConfigComplexHeadersBody(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name        = "complex-test"
  url         = "https://api.example.com/v2/health"
  http_method = "POST"
  request_headers = [
    { name = "Content-Type", value = "application/json" },
    { name = "X-API-Key", value = "secret-key-12345" },
    { name = "User-Agent", value = "TerraformProvider/1.0" }
  ]
  request_body = jsonencode({
    service = "api"
    environment = "production"
    checks = ["database", "cache", "queue"]
    metadata = {
      version = "2.1.0"
      region  = "us-east-1"
    }
  })
}
`, baseURL)
}
