// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIncidentsDataSource_basic(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Pre-create incidents
	for i := 1; i <= 3; i++ {
		id := fmt.Sprintf("inci%d", i)
		server.incidents[id] = map[string]interface{}{
			"uuid":        id,
			"title":       map[string]interface{}{"en": fmt.Sprintf("Incident %d", i)},
			"text":        map[string]interface{}{"en": fmt.Sprintf("Description %d", i)},
			"type":        "incident",
			"statuspages": []string{"sp_main"},
		}
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_incidents.all", "incidents.#", "3"),
				),
			},
		},
	})
}

func TestAccIncidentsDataSource_empty(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_incidents.all", "incidents.#", "0"),
				),
			},
		},
	})
}

func testAccIncidentsDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_incidents" "all" {
}
`, baseURL)
}
