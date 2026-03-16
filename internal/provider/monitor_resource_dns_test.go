// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccMonitorResource_dnsBasic tests basic DNS monitor creation with defaults.
func TestAccMonitorResource_dnsBasic(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigDNSBasic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "dns-basic"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "dns"),
					// dns_record_type should be computed to "A" (API default)
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", "A"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccMonitorResource_dnsFull tests DNS monitor creation with all DNS fields.
func TestAccMonitorResource_dnsFull(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigDNSFull(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "dns-full"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "dns"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", "CNAME"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_nameserver", "8.8.8.8"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_expected_answer", "www.example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "300"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccMonitorResource_dnsUpdate tests updating DNS monitor fields.
func TestAccMonitorResource_dnsUpdate(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with full DNS config
			{
				Config: testAccMonitorResourceConfigDNSFull(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", "CNAME"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_nameserver", "8.8.8.8"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_expected_answer", "www.example.com"),
				),
			},
			// Update all DNS fields
			{
				Config: testAccMonitorResourceConfigDNSUpdate(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "dns-updated"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", "MX"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_nameserver", "1.1.1.1"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_expected_answer", "mail.example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "600"),
				),
			},
		},
	})
}

// TestAccMonitorResource_dnsFieldsOnHTTP tests that DNS fields are rejected on HTTP protocol.
func TestAccMonitorResource_dnsFieldsOnHTTP(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "hyperping_monitor" "test" {
  name            = "http-with-dns-fields"
  url             = "https://example.com"
  protocol        = "http"
  dns_record_type = "A"
}
`,
				ExpectError: regexp.MustCompile(`dns_record_type is only valid for DNS monitors`),
			},
		},
	})
}

// TestAccMonitorResource_dnsFieldsOnICMP tests that DNS fields are rejected on ICMP protocol.
func TestAccMonitorResource_dnsFieldsOnICMP(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "hyperping_monitor" "test" {
  name           = "icmp-with-dns-fields"
  url            = "https://example.com"
  protocol       = "icmp"
  dns_nameserver = "8.8.8.8"
}
`,
				ExpectError: regexp.MustCompile(`dns_nameserver is only valid for DNS monitors`),
			},
		},
	})
}

// TestAccMonitorResource_dnsFieldsOnPort tests that DNS fields are rejected on port protocol.
func TestAccMonitorResource_dnsFieldsOnPort(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "hyperping_monitor" "test" {
  name                = "port-with-dns-fields"
  url                 = "https://example.com"
  protocol            = "port"
  port                = 5432
  dns_expected_answer = "127.0.0.1"
}
`,
				ExpectError: regexp.MustCompile(`dns_expected_answer is only valid for DNS monitors`),
			},
		},
	})
}

// TestAccMonitorResource_dnsWithHTTPFields tests that HTTP fields are rejected on DNS protocol.
func TestAccMonitorResource_dnsWithHTTPFields(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "hyperping_monitor" "test" {
  name        = "dns-with-http-fields"
  url         = "example.com"
  protocol    = "dns"
  http_method = "POST"
}
`,
				ExpectError: regexp.MustCompile(`http_method is only valid for HTTP monitors`),
			},
		},
	})
}

// TestAccMonitorResource_dnsWithPort tests that port is rejected on DNS protocol.
func TestAccMonitorResource_dnsWithPort(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "hyperping_monitor" "test" {
  name     = "dns-with-port"
  url      = "example.com"
  protocol = "dns"
  port     = 53
}
`,
				ExpectError: regexp.MustCompile(`port is not valid when protocol is "dns"`),
			},
		},
	})
}

// TestAccMonitorResource_dnsRecordTypeNotLeakedToHTTP verifies that updating
// an HTTP monitor does not cause dns_record_type to appear in its state.
func TestAccMonitorResource_dnsRecordTypeNotLeakedToHTTP(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create a basic HTTP monitor
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "no-leak-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_record_type"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_nameserver"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_expected_answer"),
				),
			},
			// Update the HTTP monitor — DNS fields must remain absent
			{
				Config: testAccMonitorResourceConfigUpdateAll(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "updated-all-fields"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_record_type"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_nameserver"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_expected_answer"),
				),
			},
		},
	})
}

// TestAccMonitorResource_dnsToHTTPSwitch tests protocol switch from DNS to HTTP.
// DNS fields should be cleared when switching to HTTP.
func TestAccMonitorResource_dnsToHTTPSwitch(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Start with a full DNS monitor
			{
				Config: testAccMonitorResourceConfigDNSFull(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "dns"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", "CNAME"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_nameserver", "8.8.8.8"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_expected_answer", "www.example.com"),
				),
			},
			// Switch to HTTP — remove DNS fields, add HTTP URL
			{
				Config: testAccMonitorResourceConfigSwitchDNSToHTTP(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "https://example.com"),
					// DNS fields must not appear in state
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_record_type"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_nameserver"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_expected_answer"),
				),
			},
		},
	})
}

// TestAccMonitorResource_httpToDNSSwitch tests protocol switch from HTTP to DNS.
func TestAccMonitorResource_httpToDNSSwitch(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Start with an HTTP monitor
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "switch-to-dns"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "http"),
				),
			},
			// Switch to DNS
			{
				Config: testAccMonitorResourceConfigDNSFull(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "dns"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", "CNAME"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_nameserver", "8.8.8.8"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_expected_answer", "www.example.com"),
				),
			},
			// Import and verify DNS fields persist
			{
				ResourceName:      "hyperping_monitor.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// TestAccMonitorResource_dnsRemoveOptionalFields tests removing optional DNS fields.
// dns_nameserver and dns_expected_answer can be removed while keeping dns_record_type.
func TestAccMonitorResource_dnsRemoveOptionalFields(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create with all DNS fields
			{
				Config: testAccMonitorResourceConfigDNSFull(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", "CNAME"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_nameserver", "8.8.8.8"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_expected_answer", "www.example.com"),
				),
			},
			// Remove nameserver and expected_answer, keep record_type
			{
				Config: testAccMonitorResourceConfigDNSRecordTypeOnly(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "dns"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", "AAAA"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_nameserver"),
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "dns_expected_answer"),
				),
			},
		},
	})
}

// TestAccMonitorResource_dnsAllRecordTypes tests that each supported DNS record type is accepted.
func TestAccMonitorResource_dnsAllRecordTypes(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	// Test a subset of record types to keep the test fast
	recordTypes := []string{"A", "AAAA", "CNAME", "MX", "NS", "TXT"}

	steps := make([]tfresource.TestStep, len(recordTypes))
	for i, rt := range recordTypes {
		rt := rt // capture loop variable
		steps[i] = tfresource.TestStep{
			Config: testAccMonitorResourceConfigDNSWithRecordType(server.URL, rt),
			Check: tfresource.ComposeAggregateTestCheckFunc(
				tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "dns"),
				tfresource.TestCheckResourceAttr("hyperping_monitor.test", "dns_record_type", rt),
			),
		}
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps:                    steps,
	})
}

// TestAccMonitorResource_dnsInvalidRecordType tests that an invalid DNS record type is rejected.
func TestAccMonitorResource_dnsInvalidRecordType(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigDNSWithRecordType(server.URL, "INVALID"),
				ExpectError: regexp.MustCompile(
					`value must be one of`,
				),
			},
		},
	})
}

// TestAccMonitorResource_dnsURLAcceptsBaredomain verifies DNS monitors accept bare domains.
func TestAccMonitorResource_dnsURLAcceptsBaredomain(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigDNSBasic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "url", "example.com"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "protocol", "dns"),
				),
			},
		},
	})
}

// TestAccMonitorResource_httpURLRejectsBareDomain verifies HTTP monitors reject bare domains.
func TestAccMonitorResource_httpURLRejectsBareDomain(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccProviderConfig(server.URL) + `
resource "hyperping_monitor" "test" {
  name     = "http-bare-domain"
  url      = "example.com"
  protocol = "http"
}
`,
				ExpectError: regexp.MustCompile(`Invalid URL Format`),
			},
		},
	})
}

// TestAccMonitorResource_dnsAllFieldsCrossValidation tests that ALL DNS fields
// are rejected on ALL non-DNS protocols (combinatorial coverage).
func TestAccMonitorResource_dnsAllFieldsCrossValidation(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tests := []struct {
		name     string
		config   string
		errMatch string
	}{
		{
			name: "http_with_dns_nameserver",
			config: `
resource "hyperping_monitor" "test" {
  name           = "cross-val-test"
  url            = "https://example.com"
  protocol       = "http"
  dns_nameserver = "8.8.8.8"
}`,
			errMatch: `dns_nameserver is only valid for DNS monitors`,
		},
		{
			name: "http_with_dns_expected_answer",
			config: `
resource "hyperping_monitor" "test" {
  name                = "cross-val-test"
  url                 = "https://example.com"
  protocol            = "http"
  dns_expected_answer = "127.0.0.1"
}`,
			errMatch: `dns_expected_answer is only valid for DNS monitors`,
		},
		{
			name: "icmp_with_dns_record_type",
			config: `
resource "hyperping_monitor" "test" {
  name            = "cross-val-test"
  url             = "https://example.com"
  protocol        = "icmp"
  dns_record_type = "A"
}`,
			errMatch: `dns_record_type is only valid for DNS monitors`,
		},
		{
			name: "icmp_with_dns_expected_answer",
			config: `
resource "hyperping_monitor" "test" {
  name                = "cross-val-test"
  url                 = "https://example.com"
  protocol            = "icmp"
  dns_expected_answer = "10.0.0.1"
}`,
			errMatch: `dns_expected_answer is only valid for DNS monitors`,
		},
		{
			name: "port_with_dns_record_type",
			config: `
resource "hyperping_monitor" "test" {
  name            = "cross-val-test"
  url             = "https://example.com"
  protocol        = "port"
  port            = 443
  dns_record_type = "MX"
}`,
			errMatch: `dns_record_type is only valid for DNS monitors`,
		},
		{
			name: "port_with_dns_nameserver",
			config: `
resource "hyperping_monitor" "test" {
  name           = "cross-val-test"
  url            = "https://example.com"
  protocol       = "port"
  port           = 443
  dns_nameserver = "1.1.1.1"
}`,
			errMatch: `dns_nameserver is only valid for DNS monitors`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tfresource.Test(t, tfresource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []tfresource.TestStep{
					{
						Config:      testAccProviderConfig(server.URL) + tc.config,
						ExpectError: regexp.MustCompile(tc.errMatch),
					},
				},
			})
		})
	}
}
