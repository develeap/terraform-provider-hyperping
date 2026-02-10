// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccHealthchecksDataSource_basic(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	// Pre-create healthchecks
	server.healthchecks["hc1"] = map[string]interface{}{
		"uuid":             "hc1",
		"name":             "Healthcheck 1",
		"periodValue":      60,
		"periodType":       "seconds",
		"period":           60,
		"gracePeriodValue": 300,
		"gracePeriodType":  "seconds",
		"gracePeriod":      300,
		"isPaused":         false,
		"isDown":           false,
		"pingUrl":          "https://ping.hyperping.io/hc1",
		"createdAt":        "2026-01-01T00:00:00Z",
	}
	server.healthchecks["hc2"] = map[string]interface{}{
		"uuid":             "hc2",
		"name":             "Healthcheck 2",
		"periodValue":      120,
		"periodType":       "seconds",
		"period":           120,
		"gracePeriodValue": 600,
		"gracePeriodType":  "seconds",
		"gracePeriod":      600,
		"isPaused":         false,
		"isDown":           false,
		"pingUrl":          "https://ping.hyperping.io/hc2",
		"createdAt":        "2026-01-01T00:00:00Z",
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccHealthchecksDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_healthchecks.all", "healthchecks.#", "2"),
				),
			},
		},
	})
}

func TestAccHealthchecksDataSource_empty(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	// No healthchecks created

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccHealthchecksDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_healthchecks.all", "healthchecks.#", "0"),
				),
			},
		},
	})
}

func TestAccHealthchecksDataSource_many(t *testing.T) {
	server := newMockHealthcheckServer(t)
	defer server.Close()

	// Create 5 healthchecks
	for i := 1; i <= 5; i++ {
		id := fmt.Sprintf("hc%d", i)
		server.healthchecks[id] = map[string]interface{}{
			"uuid":             id,
			"name":             fmt.Sprintf("Healthcheck %d", i),
			"periodValue":      60 * i,
			"periodType":       "seconds",
			"period":           60 * i,
			"gracePeriodValue": 300,
			"gracePeriodType":  "seconds",
			"gracePeriod":      300,
			"isPaused":         false,
			"isDown":           false,
			"pingUrl":          fmt.Sprintf("https://ping.hyperping.io/%s", id),
			"createdAt":        "2026-01-01T00:00:00Z",
		}
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccHealthchecksDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_healthchecks.all", "healthchecks.#", "5"),
				),
			},
		},
	})
}

// Helper functions

func testAccHealthchecksDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_healthchecks" "all" {
}
`, baseURL)
}
