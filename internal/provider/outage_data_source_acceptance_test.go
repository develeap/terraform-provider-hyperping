// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOutageDataSource_basic(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	// Pre-create an outage (use camelCase to match API)
	server.outages["out_test1"] = map[string]interface{}{
		"uuid":        "out_test1",
		"monitor":     map[string]interface{}{"uuid": "mon_test1"},
		"startDate":   startDate,
		"endDate":     endDate,
		"outageType":  "manual",
		"isResolved":  true,
		"statusCode":  500,
		"description": "Test outage",
	}

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccOutageDataSourceConfig(server.URL, "out_test1"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_outage.test", "id", "out_test1"),
					tfresource.TestCheckResourceAttr("data.hyperping_outage.test", "monitor_uuid", "mon_test1"),
					tfresource.TestCheckResourceAttr("data.hyperping_outage.test", "outage_type", "manual"),
					tfresource.TestCheckResourceAttr("data.hyperping_outage.test", "is_resolved", "true"),
					tfresource.TestCheckResourceAttr("data.hyperping_outage.test", "status_code", "500"),
				),
			},
		},
	})
}

func TestAccOutageDataSource_ongoing(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-2 * time.Hour).Format(time.RFC3339)

	// Pre-create an ongoing outage
	server.outages["out_ongoing"] = map[string]interface{}{
		"uuid":         "out_ongoing",
		"monitor_uuid": "mon_test2",
		"start_date":   startDate,
		"outage_type":  "manual",
		"is_resolved":  false,
	}

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccOutageDataSourceConfig(server.URL, "out_ongoing"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_outage.test", "id", "out_ongoing"),
					tfresource.TestCheckResourceAttr("data.hyperping_outage.test", "is_resolved", "false"),
				),
			},
		},
	})
}

func testAccOutageDataSourceConfig(baseURL, id string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_outage" "test" {
  id = %[2]q
}
`, baseURL, id)
}
