// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// Test 1: alerts_wait Extreme Values
// =============================================================================

// TestAccMonitorResource_alertsWaitExtremeValues tests alerts_wait with
// extreme values like 3600 (1 hour) and 7200 (2 hours).
// Note: alerts_wait=0 is treated as null by provider mapping logic (see monitor_resource_edge_cases_test.go).
// This test focuses on high values that represent very long alert delays.
func TestAccMonitorResource_alertsWaitExtremeValues(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test with alerts_wait=60 (baseline: 1 minute)
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 60),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "60"),
				),
			},
			// Test with alerts_wait=3600 (1 hour, very long delay)
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 3600),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "3600"),
				),
			},
			// Test with alerts_wait=7200 (2 hours, extreme delay)
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 7200),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "7200"),
				),
			},
			// Back to moderate value
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 600),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "600"),
				),
			},
		},
	})
}

// =============================================================================
// Test 2: required_keyword Unicode and Long Strings
// =============================================================================

// TestAccMonitorResource_requiredKeywordUnicodeAndLongStrings tests
// required_keyword with Unicode characters and very long strings (1000+ chars)
// to complement the existing requiredKeywordEdgeCases test.
func TestAccMonitorResource_requiredKeywordUnicodeAndLongStrings(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test with Unicode characters (Chinese)
			{
				Config: testAccMonitorResourceConfigWithRequiredKeyword(server.URL, "健康状态"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "required_keyword", "健康状态"),
				),
			},
			// Test with mixed Unicode (emojis)
			{
				Config: testAccMonitorResourceConfigWithRequiredKeyword(server.URL, "✅ Status: OK 🟢"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "required_keyword", "✅ Status: OK 🟢"),
				),
			},
			// Test with very long string (1000+ characters)
			{
				Config: testAccMonitorResourceConfigWithRequiredKeyword(server.URL, generateLongString(1000)),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "required_keyword", generateLongString(1000)),
				),
			},
			// Test removal
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "keyword-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckNoResourceAttr("hyperping_monitor.test", "required_keyword"),
				),
			},
		},
	})
}

// =============================================================================
// Utility Functions
// =============================================================================

func generateLongString(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "A"
	}
	return result
}
