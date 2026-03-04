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
// the full range of valid enumerated values, including edge cases.
// Valid values: -1 (disabled), 0, 1, 2, 3, 5, 10, 30, 60 (minutes).
// Note: alerts_wait=0 is treated as null by provider mapping logic.
func TestAccMonitorResource_alertsWaitExtremeValues(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Test with alerts_wait=1 (1 minute, smallest positive)
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 1),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "1"),
				),
			},
			// Test with alerts_wait=60 (1 hour, largest valid value)
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 60),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "60"),
				),
			},
			// Test with alerts_wait=-1 (disabled)
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, -1),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "-1"),
				),
			},
			// Back to moderate value
			{
				Config: testAccMonitorResourceConfigWithAlertsWait(server.URL, 5),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "alerts_wait", "5"),
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
