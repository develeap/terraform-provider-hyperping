// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOutagesDataSource_basic(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()

	// Pre-create outages
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("out%d", i)
		server.outages[id] = map[string]interface{}{
			"uuid":         id,
			"monitor_uuid": fmt.Sprintf("mon%d", i),
			"start_date":   now.Add(time.Duration(-i) * time.Hour).Format(time.RFC3339),
			"end_date":     now.Format(time.RFC3339),
			"outage_type":  "manual",
			"is_resolved":  true,
		}
	}

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccOutagesDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_outages.all", "outages.#", "3"),
				),
			},
		},
	})
}

func TestAccOutagesDataSource_empty(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccOutagesDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_outages.all", "outages.#", "0"),
				),
			},
		},
	})
}

func testAccOutagesDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_outages" "all" {
}
`, baseURL)
}
