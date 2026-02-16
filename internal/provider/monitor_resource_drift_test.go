// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// TestAccMonitorResource_driftDetection_externalPause tests detection of external pause state changes
func TestAccMonitorResource_driftDetection_externalPause(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create active monitor, externally pause it, and detect drift
			{
				Config: testAccMonitorResourceConfigWithPaused(server.URL, "drift-pause-test", false),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "paused", "false"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
					// Externally pause the monitor
					testAccExternallyPauseMonitor(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMonitorResource_driftDetection_nameChange tests detection of external name changes
func TestAccMonitorResource_driftDetection_nameChange(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create monitor, externally change name, and detect drift
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "original-name"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "original-name"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
					// Externally change name
					testAccExternallyChangeMonitorName(server, "externally-modified-name"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMonitorResource_driftDetection_frequencyChange tests detection of external frequency changes
func TestAccMonitorResource_driftDetection_frequencyChange(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create monitor, externally change frequency, and detect drift
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "frequency-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "check_frequency", "60"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
					// Externally change frequency
					testAccExternallyChangeMonitorFrequency(server, 300),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMonitorResource_driftDetection_externalDeletion tests detection of external deletion
func TestAccMonitorResource_driftDetection_externalDeletion(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create monitor, externally delete it, and detect drift
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "deletion-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "deletion-test"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
					// Externally delete the monitor
					testAccCheckMonitorDisappears(server),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccMonitorResource_driftDetection_requiredKeyword tests write-only field consistency
func TestAccMonitorResource_driftDetection_requiredKeyword(t *testing.T) {
	server := newMockHyperpingServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create monitor with required_keyword and externally change it
			// No drift should be detected because API doesn't return this field
			{
				Config: testAccMonitorResourceConfigWithRequiredKeyword(server.URL, "healthy"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "name", "keyword-test"),
					tfresource.TestCheckResourceAttr("hyperping_monitor.test", "required_keyword", "healthy"),
					tfresource.TestCheckResourceAttrSet("hyperping_monitor.test", "id"),
					// Externally change required_keyword (write-only field)
					testAccExternallyChangeMonitorKeyword(server, "status:ok"),
				),
				// No drift detected because API doesn't return this write-only field
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

// Helper functions for external state manipulation

// testAccExternallyPauseMonitor simulates external pause operation
func testAccExternallyPauseMonitor(server *mockHyperpingServer) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for id, monitor := range server.monitors {
			monitor["paused"] = true
			server.monitors[id] = monitor
		}
		return nil
	}
}

// testAccExternallyChangeMonitorName simulates external name change
func testAccExternallyChangeMonitorName(server *mockHyperpingServer, newName string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for id, monitor := range server.monitors {
			monitor["name"] = newName
			server.monitors[id] = monitor
		}
		return nil
	}
}

// testAccExternallyChangeMonitorFrequency simulates external frequency change
func testAccExternallyChangeMonitorFrequency(server *mockHyperpingServer, newFrequency int) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for id, monitor := range server.monitors {
			monitor["check_frequency"] = newFrequency
			server.monitors[id] = monitor
		}
		return nil
	}
}

// testAccExternallyChangeMonitorKeyword simulates external keyword change
// Note: This field is write-only in the API, so changes won't be visible to Terraform
func testAccExternallyChangeMonitorKeyword(server *mockHyperpingServer, newKeyword string) tfresource.TestCheckFunc {
	return func(s *terraform.State) error {
		for id, monitor := range server.monitors {
			monitor["required_keyword"] = newKeyword
			server.monitors[id] = monitor
		}
		return nil
	}
}
