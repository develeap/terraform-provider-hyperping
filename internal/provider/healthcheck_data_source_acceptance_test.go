// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHealthcheckDataSource_basic(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	// Pre-create a healthcheck
	server.healthchecks["hc_test1"] = map[string]interface{}{
		"uuid":             "hc_test1",
		"name":             "Test Healthcheck",
		"periodValue":      60,
		"periodType":       "seconds",
		"period":           60,
		"gracePeriodValue": 300,
		"gracePeriodType":  "seconds",
		"gracePeriod":      300,
		"isPaused":         false,
		"isDown":           false,
		"pingUrl":          "https://ping.hyperping.io/hc_test1",
		"createdAt":        "2026-01-01T00:00:00Z",
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccHealthcheckDataSourceConfig(server.URL, "hc_test1"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "id", "hc_test1"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "name", "Test Healthcheck"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "period_value", "60"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "period_type", "seconds"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "period", "60"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "grace_period_value", "300"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "grace_period_type", "seconds"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "grace_period", "300"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "is_paused", "false"),
					tfresource.TestCheckResourceAttrSet("data.hyperping_healthcheck.test", "ping_url"),
				),
			},
		},
	})
}

func TestAccHealthcheckDataSource_withEscalationPolicy(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	// Pre-create a healthcheck with escalation policy
	server.healthchecks["hc_ep1"] = map[string]interface{}{
		"uuid":             "hc_ep1",
		"name":             "EP Healthcheck",
		"periodValue":      120,
		"periodType":       "seconds",
		"period":           120,
		"gracePeriodValue": 600,
		"gracePeriodType":  "seconds",
		"gracePeriod":      600,
		"isPaused":         false,
		"isDown":           false,
		"pingUrl":          "https://ping.hyperping.io/hc_ep1",
		"createdAt":        "2026-01-01T00:00:00Z",
		"escalationPolicy": map[string]interface{}{
			"uuid": "ep_test123",
		},
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccHealthcheckDataSourceConfig(server.URL, "hc_ep1"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "id", "hc_ep1"),
					tfresource.TestCheckResourceAttr("data.hyperping_healthcheck.test", "escalation_policy", "ep_test123"),
				),
			},
		},
	})
}

// Helper functions

func testAccHealthcheckDataSourceConfig(baseURL, id string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_healthcheck" "test" {
  id = %[2]q
}
`, baseURL, id)
}
