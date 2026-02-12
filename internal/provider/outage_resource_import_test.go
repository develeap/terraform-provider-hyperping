// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOutageResource_import(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Step 1: Create the resource
			{
				Config: testAccOutageResourceConfig_basic(server.URL, startDate, endDate),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor_uuid", "mon_test123"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "start_date", startDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "end_date", endDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "outage_type", "manual"),
					tfresource.TestCheckResourceAttrSet("hyperping_outage.test", "id"),
				),
			},
			// Step 2: Import it and verify all fields match
			{
				ResourceName:      "hyperping_outage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOutageResource_importFull(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-2 * time.Hour).Format(time.RFC3339)
	endDate := now.Add(-1 * time.Hour).Format(time.RFC3339)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource with all optional fields
			{
				Config: testAccOutageResourceConfig_full(server.URL, startDate, endDate),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "monitor_uuid", "mon_test456"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "start_date", startDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "end_date", endDate),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "status_code", "500"),
					tfresource.TestCheckResourceAttr("hyperping_outage.test", "description", "Server error"),
				),
			},
			// Import and verify zero-drift
			{
				ResourceName:      "hyperping_outage.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccOutageResource_importNotFound(t *testing.T) {
	server := newMockOutageServer(t)
	defer server.Close()

	now := time.Now().UTC()
	startDate := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endDate := now.Format(time.RFC3339)

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create resource first
			{
				Config: testAccOutageResourceConfig_basic(server.URL, startDate, endDate),
			},
			// Try to import non-existent resource
			{
				ResourceName:  "hyperping_outage.test",
				ImportState:   true,
				ImportStateId: "out_nonexistent",
				ExpectError:   regexp.MustCompile("Outage not found|Cannot import non-existent"),
			},
		},
	})
}
