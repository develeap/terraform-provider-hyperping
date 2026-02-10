// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccIncidentUpdateResource_basic(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Pre-create an incident to add updates to
	server.incidents["inci_base"] = map[string]interface{}{
		"uuid":        "inci_base",
		"title":       map[string]interface{}{"en": "Base Incident"},
		"text":        map[string]interface{}{"en": "Initial description"},
		"type":        "incident",
		"statuspages": []string{"sp_main"},
		"updates":     []interface{}{},
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create and Read testing
			{
				Config: testAccIncidentUpdateResourceConfig_basic(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "incident_id", "inci_base"),
					tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "text", "We are investigating the issue"),
					tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "type", "investigating"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident_update.test", "id"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident_update.test", "date"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "hyperping_incident_update.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIncidentUpdateResource_allTypes(t *testing.T) {
	updateTypes := []string{"investigating", "identified", "update", "monitoring", "resolved"}

	for _, updateType := range updateTypes {
		updateType := updateType // Capture range variable
		t.Run(updateType, func(t *testing.T) {
			// Each subtest gets its own mock server to avoid conflicts
			server := newMockIncidentServer(t)
			defer server.Close()

			// Pre-create a unique incident for this subtest
			incidentID := fmt.Sprintf("inci_%s", updateType)
			server.incidents[incidentID] = map[string]interface{}{
				"uuid":        incidentID,
				"title":       map[string]interface{}{"en": fmt.Sprintf("Test %s", updateType)},
				"text":        map[string]interface{}{"en": "Testing update type"},
				"type":        "incident",
				"statuspages": []string{"sp_main"},
				"updates":     []interface{}{},
			}

			tfresource.ParallelTest(t, tfresource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []tfresource.TestStep{
					{
						Config: testAccIncidentUpdateResourceConfig_withTypeAndIncident(server.URL, incidentID, updateType),
						Check: tfresource.ComposeAggregateTestCheckFunc(
							tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "incident_id", incidentID),
							tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "type", updateType),
						),
					},
				},
			})
		})
	}
}

func TestAccIncidentUpdateResource_disappears(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Pre-create an incident
	server.incidents["inci_disappear"] = map[string]interface{}{
		"uuid":        "inci_disappear",
		"title":       map[string]interface{}{"en": "Disappear Test"},
		"text":        map[string]interface{}{"en": "Test"},
		"type":        "incident",
		"statuspages": []string{"sp_main"},
		"updates":     []interface{}{},
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccIncidentUpdateResourceConfig_disappear(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttrSet("hyperping_incident_update.test", "id"),
					testAccCheckIncidentUpdateDisappears(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIncidentUpdateResource_createError(t *testing.T) {
	server := newMockIncidentServerWithErrors(t)
	defer server.Close()

	// Pre-create an incident
	server.incidents["inci_error"] = map[string]interface{}{
		"uuid":        "inci_error",
		"title":       map[string]interface{}{"en": "Error Test"},
		"text":        map[string]interface{}{"en": "Test"},
		"type":        "incident",
		"statuspages": []string{"sp_main"},
		"updates":     []interface{}{},
	}

	server.setUpdateError(true)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccIncidentUpdateResourceConfig_basic(server.URL),
				ExpectError: regexp.MustCompile(`Error creating incident update`),
			},
		},
	})
}

// Helper functions

func testAccIncidentUpdateResourceConfig_basic(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident_update" "test" {
  incident_id = "inci_base"
  text        = "We are investigating the issue"
  type        = "investigating"
}
`, baseURL)
}

func testAccIncidentUpdateResourceConfig_withType(baseURL, updateType string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident_update" "test" {
  incident_id = "inci_types"
  text        = "Update of type %[2]s"
  type        = %[2]q
}
`, baseURL, updateType)
}

func testAccIncidentUpdateResourceConfig_withTypeAndIncident(baseURL, incidentID, updateType string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident_update" "test" {
  incident_id = %[2]q
  text        = "Update of type %[3]s"
  type        = %[3]q
}
`, baseURL, incidentID, updateType)
}

func testAccIncidentUpdateResourceConfig_disappear(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

resource "hyperping_incident_update" "test" {
  incident_id = "inci_disappear"
  text        = "This will disappear"
  type        = "update"
}
`, baseURL)
}

func testAccCheckIncidentUpdateDisappears(server *mockIncidentServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Clear all updates from the incident
		if incident, ok := server.incidents["inci_disappear"]; ok {
			incident["updates"] = []interface{}{}
			server.incidents["inci_disappear"] = incident
		}
		return nil
	}
}
