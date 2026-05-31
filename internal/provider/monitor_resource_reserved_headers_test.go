// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccMonitorResource_reservedHeadersRejected verifies end-to-end that
// every framing/connection-control header on the reserved-name banlist is
// rejected at plan time with a "Reserved Header Name" diagnostic, and that
// header names with leading/trailing/internal whitespace or non-token
// characters are rejected with an "Invalid Header Name" diagnostic.
//
// This guards against regressions where a header is silently dropped from
// the validator map or where the whitespace-trim/token-grammar check is
// removed, which would re-open the smuggling/bypass vectors documented on
// the reservedHeaderNames map.
func TestAccMonitorResource_reservedHeadersRejected(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	reservedCases := []struct {
		name       string
		headerName string
	}{
		{"Host", "Host"},
		{"Transfer-Encoding", "Transfer-Encoding"},
		{"Content-Length", "Content-Length"},
		{"Connection", "Connection"},
		{"Upgrade", "Upgrade"},
		{"TE", "TE"},
		{"Trailer", "Trailer"},
		{"Expect", "Expect"},
	}

	for _, tc := range reservedCases {
		t.Run("reserved/"+tc.name, func(t *testing.T) {
			tfresource.ParallelTest(t, tfresource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []tfresource.TestStep{
					{
						Config:      testAccMonitorResourceConfigWithSingleHeader(server.URL, tc.headerName, "value"),
						ExpectError: regexp.MustCompile(`Reserved Header Name`),
					},
				},
			})
		})
	}

	invalidNameCases := []struct {
		name       string
		headerName string
	}{
		{"trailing-space", "Host "},
		{"leading-space", " Host"},
		{"internal-space", "X-Foo Host"},
	}

	for _, tc := range invalidNameCases {
		t.Run("invalid/"+tc.name, func(t *testing.T) {
			tfresource.ParallelTest(t, tfresource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []tfresource.TestStep{
					{
						Config:      testAccMonitorResourceConfigWithSingleHeader(server.URL, tc.headerName, "value"),
						ExpectError: regexp.MustCompile(`Invalid Header Name`),
					},
				},
			})
		})
	}
}

// testAccMonitorResourceConfigWithSingleHeader returns an HCL fragment that
// declares a monitor with exactly one request_header entry. Used by the
// reserved-header acceptance test to drive each banned name through plan in
// isolation.
func testAccMonitorResourceConfigWithSingleHeader(baseURL, headerName, headerValue string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_monitor" "test" {
  name = "reserved-header-test"
  url  = "https://example.com"
  request_headers = [
    { name = %[2]q, value = %[3]q }
  ]
}
`, baseURL, headerName, headerValue)
}
