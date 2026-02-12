// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIncidentUpdateResource_import(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	// Pre-create an incident to add updates to
	server.incidents["inci_base"] = map[string]interface{}{
		"uuid":        "inci_base",
		"title":       map[string]interface{}{"en": "Import Test Incident"},
		"text":        map[string]interface{}{"en": "Initial description"},
		"type":        "incident",
		"statuspages": []string{"sp_main"},
		"updates":     []interface{}{},
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Step 1: Create the resource
			{
				Config: testAccIncidentUpdateResourceConfig_basic(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "incident_id", "inci_base"),
					tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "text", "We are investigating the issue"),
					tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "type", "investigating"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident_update.test", "id"),
				),
			},
			// Step 2: Import it and verify all fields match
			{
				ResourceName:      "hyperping_incident_update.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIncidentUpdateResource_importAllTypes(t *testing.T) {
	updateTypes := []string{"investigating", "identified", "update", "monitoring", "resolved"}

	for _, updateType := range updateTypes {
		updateType := updateType
		t.Run(updateType, func(t *testing.T) {
			server := newMockIncidentServer(t)
			defer server.Close()

			incidentID := "inci_import_type"
			server.incidents[incidentID] = map[string]interface{}{
				"uuid":        incidentID,
				"title":       map[string]interface{}{"en": "Type Import Test"},
				"text":        map[string]interface{}{"en": "Testing"},
				"type":        "incident",
				"statuspages": []string{"sp_main"},
				"updates":     []interface{}{},
			}

			tfresource.ParallelTest(t, tfresource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []tfresource.TestStep{
					// Create resource
					{
						Config: testAccIncidentUpdateResourceConfig_withTypeAndIncident(server.URL, incidentID, updateType),
						Check: tfresource.ComposeTestCheckFunc(
							tfresource.TestCheckResourceAttr("hyperping_incident_update.test", "type", updateType),
						),
					},
					// Import and verify
					{
						ResourceName:      "hyperping_incident_update.test",
						ImportState:       true,
						ImportStateVerify: true,
					},
				},
			})
		})
	}
}

func TestAccIncidentUpdateResource_importNotFound(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	server.incidents["inci_base"] = map[string]interface{}{
		"uuid":        "inci_base",
		"title":       map[string]interface{}{"en": "Base Incident"},
		"text":        map[string]interface{}{"en": "Initial"},
		"type":        "incident",
		"statuspages": []string{"sp_main"},
		"updates":     []interface{}{},
	}

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource first
			{
				Config: testAccIncidentUpdateResourceConfig_basic(server.URL),
			},
			// Try to import non-existent resource
			{
				ResourceName:  "hyperping_incident_update.test",
				ImportState:   true,
				ImportStateId: "inci_base/upd_nonexistent",
				ExpectError:   regexp.MustCompile("Update not found|Cannot import non-existent"),
			},
		},
	})
}
