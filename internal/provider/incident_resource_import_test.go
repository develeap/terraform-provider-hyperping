// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIncidentResource_import(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Step 1: Create the resource
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "Test Import Incident"),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Test Import Incident"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "text", "Something went wrong"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "type", "incident"),
					tfresource.TestCheckResourceAttrSet("hyperping_incident.test", "id"),
				),
			},
			// Step 2: Import it and verify all fields match
			{
				ResourceName:      "hyperping_incident.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIncidentResource_importFull(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource with all optional fields
			{
				Config: testAccIncidentResourceConfig_full(server.URL),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "title", "Major Outage"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "text", "We are experiencing a major outage"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "type", "outage"),
					tfresource.TestCheckResourceAttr("hyperping_incident.test", "status_pages.#", "1"),
				),
			},
			// Import and verify zero-drift
			{
				ResourceName:      "hyperping_incident.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIncidentResource_importNotFound(t *testing.T) {
	server := newMockIncidentServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource first
			{
				Config: testAccIncidentResourceConfig_basic(server.URL, "test"),
			},
			// Try to import non-existent resource
			{
				ResourceName:  "hyperping_incident.test",
				ImportState:   true,
				ImportStateId: "inc_nonexistent",
				ExpectError:   regexp.MustCompile("Incident not found|Cannot import non-existent"),
			},
		},
	})
}
