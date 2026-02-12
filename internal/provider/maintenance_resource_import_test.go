// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMaintenanceResource_importBasic(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_import_123",
		Name:      "Import Test",
		Title:     "Import Test Title",
		Text:      "Testing import functionality",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			// Step 1: Create the resource
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "id", "mw_import_123"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Import Test"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "title", "Import Test Title"),
				),
			},
			// Step 2: Import it and verify all fields match
			{
				ResourceName:      "hyperping_maintenance.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notification_minutes",
					"notification_option",
				},
			},
		},
	})
}

func TestAccMaintenanceResource_importFull(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(48, 4)

	fixture := &maintenanceTestFixture{
		UUID:                "mw_import_full_123",
		Name:                "Full Import Test",
		Title:               "Full Import Title",
		Text:                "Testing full import",
		StartDate:           startStr,
		EndDate:             endStr,
		Monitors:            []string{"mon_123", "mon_456"},
		StatusPages:         []string{"sp_main"},
		NotificationOption:  "scheduled",
		NotificationMinutes: 120,
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			// Create resource with all optional fields
			{
				Config: generateMaintenanceConfigFull(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors, fixture.StatusPages, fixture.NotificationOption, fixture.NotificationMinutes),
				Check: tfresource.ComposeTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "id", "mw_import_full_123"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_option", "scheduled"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_minutes", "120"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "monitors.#", "2"),
				),
			},
			// Import and verify zero-drift
			{
				ResourceName:      "hyperping_maintenance.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"notification_minutes",
					"notification_option",
				},
			},
		},
	})
}

func TestAccMaintenanceResource_importNotFound(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_test_123",
		Name:      "Test",
		Title:     "Test Title",
		Text:      "Test description",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	server := newSimpleMaintenanceServer(fixture)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			// Create resource first
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
			},
			// Try to import non-existent resource
			{
				ResourceName:  "hyperping_maintenance.test",
				ImportState:   true,
				ImportStateId: "mw_nonexistent",
				ExpectError:   regexp.MustCompile("Maintenance not found|Cannot import non-existent"),
			},
		},
	})
}
