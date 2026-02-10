// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIncidentDataSource_basic(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Pre-create an incident
	server.incidents["inci_test1"] = map[string]interface{}{
		"uuid":        "inci_test1",
		"title":       map[string]interface{}{"en": "Test Incident"},
		"text":        map[string]interface{}{"en": "Test description"},
		"type":        "incident",
		"statuspages": []string{"sp_main"},
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentDataSourceConfig(server.URL, "inci_test1"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_incident.test", "id", "inci_test1"),
					tfresource.TestCheckResourceAttr("data.hyperping_incident.test", "title", "Test Incident"),
					tfresource.TestCheckResourceAttr("data.hyperping_incident.test", "text", "Test description"),
					tfresource.TestCheckResourceAttr("data.hyperping_incident.test", "type", "incident"),
				),
			},
		},
	})
}

func TestAccIncidentDataSource_withComponents(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Pre-create an incident with affected components
	server.incidents["inci_comp"] = map[string]interface{}{
		"uuid":               "inci_comp",
		"title":              map[string]interface{}{"en": "Component Incident"},
		"text":               map[string]interface{}{"en": "Components affected"},
		"type":               "outage",
		"statuspages":        []string{"sp_main"},
		"affectedComponents": []interface{}{"comp_1", "comp_2"},
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentDataSourceConfig(server.URL, "inci_comp"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_incident.test", "id", "inci_comp"),
					tfresource.TestCheckResourceAttr("data.hyperping_incident.test", "type", "outage"),
					tfresource.TestCheckResourceAttr("data.hyperping_incident.test", "affected_components.#", "2"),
				),
			},
		},
	})
}

func testAccIncidentDataSourceConfig(baseURL, id string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_incident" "test" {
  id = %[2]q
}
`, baseURL, id)
}
