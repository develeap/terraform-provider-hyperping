// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccIncidentResource_driftDetection_statusChange tests detection of external status changes
func TestAccIncidentResource_driftDetection_statusChange(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create incident and externally change type
			{
				Config: testAccIncidentResourceConfig_withType(server.URL, "incident"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "type", "incident"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
					// Externally change type to "outage"
					testAccExternallyChangeIncidentType(server, "outage"),
				),
				// Drift detected: type changed externally
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccIncidentResource_driftDetection_titleChange tests detection of external title changes
func TestAccIncidentResource_driftDetection_titleChange(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create incident and externally change title
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "Original Title"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Original Title"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
					// Externally change title
					testAccExternallyChangeIncidentTitle(server, "Externally Modified Title"),
				),
				// Drift detected: title changed externally
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccIncidentResource_driftDetection_textChange tests detection of external text changes
func TestAccIncidentResource_driftDetection_textChange(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create incident and externally change text
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "Test Incident"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "text", "Something went wrong"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
					// Externally change text
					testAccExternallyChangeIncidentText(server, "Text modified externally"),
				),
				// Drift detected: text changed externally
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccIncidentResource_driftDetection_externalDeletion tests detection of external deletion
func TestAccIncidentResource_driftDetection_externalDeletion(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create incident and externally delete it
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "Deletion Test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Deletion Test"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
					// Externally delete the incident
					testAccCheckIncidentDisappears(server),
				),
				// Drift detected: resource no longer exists
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccIncidentResource_driftDetection_statusPagesChange tests detection of status page changes
func TestAccIncidentResource_driftDetection_statusPagesChange(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create incident and externally change status pages
			{
				Config: testAccIncidentResourceConfig_full(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "status_pages.#", "1"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
					// Externally change status pages
					testAccExternallyChangeIncidentStatusPages(server, []string{"sp_external1", "sp_external2"}),
				),
				// Drift detected: status pages changed externally
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Helper functions for external incident state manipulation

// testAccExternallyChangeIncidentType simulates external type change
func testAccExternallyChangeIncidentType(server *mockIncidentServer, newType string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for id, incident := range server.incidents {
			incident["type"] = newType
			server.incidents[id] = incident
		}
		return nil
	}
}

// testAccExternallyChangeIncidentTitle simulates external title change
func testAccExternallyChangeIncidentTitle(server *mockIncidentServer, newTitle string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for id, incident := range server.incidents {
			// Update localized title
			incident["title"] = map[string]interface{}{
				"en": newTitle,
			}
			server.incidents[id] = incident
		}
		return nil
	}
}

// testAccExternallyChangeIncidentText simulates external text change
func testAccExternallyChangeIncidentText(server *mockIncidentServer, newText string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for id, incident := range server.incidents {
			// Update localized text
			incident["text"] = map[string]interface{}{
				"en": newText,
			}
			server.incidents[id] = incident
		}
		return nil
	}
}

// testAccExternallyChangeIncidentStatusPages simulates external status pages change
func testAccExternallyChangeIncidentStatusPages(server *mockIncidentServer, newPages []string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for id, incident := range server.incidents {
			// Convert string slice to interface slice
			pages := make([]interface{}, len(newPages))
			for i, page := range newPages {
				pages[i] = page
			}
			incident["statuspages"] = pages
			server.incidents[id] = incident
		}
		return nil
	}
}

// testAccCheckIncidentDisappears is defined in incident_resource_test.go to avoid duplication
