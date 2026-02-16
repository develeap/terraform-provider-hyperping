// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccMaintenanceResource_driftDetection_timeChange tests detection of external time changes
func TestAccMaintenanceResource_driftDetection_timeChange(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)
	newStartStr, newEndStr := generateMaintenanceTimeRange(48, 4)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_drift_time_123",
		Name:      "Time Drift Test",
		Title:     "Time Drift Test",
		Text:      "Testing time drift detection",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)
	server := httptest.NewServer(mock.handler())
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			// Create maintenance window and externally change times
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "start_date", startStr),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "end_date", endStr),
					tfresource.TestCheckResourceAttrSet("hyperping_maintenance.test", "id"),
					// Externally change times
					testAccExternallyChangeMaintenanceTime(mock, newStartStr, newEndStr),
				),
				// Drift detected: times changed externally
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMaintenanceResource_driftDetection_titleChange tests detection of external title changes
func TestAccMaintenanceResource_driftDetection_titleChange(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_drift_title_123",
		Name:      "Title Drift Test",
		Title:     "Original Title",
		Text:      "Testing title drift detection",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)
	server := httptest.NewServer(mock.handler())
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			// Create maintenance window and externally change title
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "title", "Original Title"),
					tfresource.TestCheckResourceAttrSet("hyperping_maintenance.test", "id"),
					// Externally change title
					testAccExternallyChangeMaintenanceTitle(mock, "Externally Modified Title"),
				),
				// Drift detected: title changed externally
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMaintenanceResource_driftDetection_monitorsChange tests detection of external monitors changes
func TestAccMaintenanceResource_driftDetection_monitorsChange(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_drift_monitors_123",
		Name:      "Monitors Drift Test",
		Title:     "Monitors Drift Test",
		Text:      "Testing monitors drift detection",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)
	server := httptest.NewServer(mock.handler())
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			// Create maintenance window and externally change monitors
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "monitors.#", "1"),
					tfresource.TestCheckResourceAttrSet("hyperping_maintenance.test", "id"),
					// Externally change monitors
					testAccExternallyChangeMaintenanceMonitors(mock, []string{"mon_123", "mon_456", "mon_789"}),
				),
				// Drift detected: monitors changed externally
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMaintenanceResource_driftDetection_externalDeletion tests detection of external deletion
func TestAccMaintenanceResource_driftDetection_externalDeletion(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:      "mw_drift_deletion_123",
		Name:      "Deletion Drift Test",
		Title:     "Deletion Drift Test",
		Text:      "Testing deletion drift detection",
		StartDate: startStr,
		EndDate:   endStr,
		Monitors:  []string{"mon_123"},
	}

	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)
	server := httptest.NewServer(mock.handler())
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			// Create maintenance window and externally delete it
			{
				Config: generateMaintenanceConfig(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "name", "Deletion Drift Test"),
					tfresource.TestCheckResourceAttrSet("hyperping_maintenance.test", "id"),
					// Externally delete the maintenance window
					testAccCheckMaintenanceDisappears(mock),
				),
				// Drift detected: resource no longer exists
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMaintenanceResource_driftDetection_notificationChange tests detection of external notification changes
func TestAccMaintenanceResource_driftDetection_notificationChange(t *testing.T) {
	startStr, endStr := generateMaintenanceTimeRange(24, 2)

	fixture := &maintenanceTestFixture{
		UUID:                "mw_drift_notification_123",
		Name:                "Notification Drift Test",
		Title:               "Notification Drift Test",
		Text:                "Testing notification drift detection",
		StartDate:           startStr,
		EndDate:             endStr,
		Monitors:            []string{"mon_123"},
		NotificationOption:  "scheduled",
		NotificationMinutes: 60,
	}

	mock := newMaintenanceMockServer()
	mock.addFixture(fixture)
	server := httptest.NewServer(mock.handler())
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"hyperping": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []tfresource.TestStep{
			// Create maintenance window and externally change notification settings
			{
				Config: generateMaintenanceConfigFull(server.URL, fixture.Name, fixture.Title, fixture.Text, startStr, endStr, fixture.Monitors, fixture.StatusPages, fixture.NotificationOption, fixture.NotificationMinutes),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_option", "scheduled"),
					tfresource.TestCheckResourceAttr("hyperping_maintenance.test", "notification_minutes", "60"),
					tfresource.TestCheckResourceAttrSet("hyperping_maintenance.test", "id"),
					// Externally change notification settings
					testAccExternallyChangeMaintenanceNotification(mock, "instant", 120),
				),
				// Drift detected: notification settings changed externally
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Helper functions for external maintenance state manipulation

// testAccExternallyChangeMaintenanceTime simulates external time change
func testAccExternallyChangeMaintenanceTime(mock *maintenanceMockServer, newStart, newEnd string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, fixture := range mock.fixtures {
			fixture.StartDate = newStart
			fixture.EndDate = newEnd
		}
		return nil
	}
}

// testAccExternallyChangeMaintenanceTitle simulates external title change
func testAccExternallyChangeMaintenanceTitle(mock *maintenanceMockServer, newTitle string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, fixture := range mock.fixtures {
			fixture.Title = newTitle
		}
		return nil
	}
}

// testAccExternallyChangeMaintenanceMonitors simulates external monitors change
func testAccExternallyChangeMaintenanceMonitors(mock *maintenanceMockServer, newMonitors []string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, fixture := range mock.fixtures {
			fixture.Monitors = newMonitors
		}
		return nil
	}
}

// testAccExternallyChangeMaintenanceNotification simulates external notification change
func testAccExternallyChangeMaintenanceNotification(mock *maintenanceMockServer, newOption string, newMinutes int) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, fixture := range mock.fixtures {
			fixture.NotificationOption = newOption
			fixture.NotificationMinutes = newMinutes
		}
		return nil
	}
}

// testAccCheckMaintenanceDisappears simulates external deletion of maintenance window
func testAccCheckMaintenanceDisappears(mock *maintenanceMockServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Mark all maintenance windows as deleted
		for uuid := range mock.fixtures {
			mock.setDeleted(uuid, true)
		}
		return nil
	}
}
