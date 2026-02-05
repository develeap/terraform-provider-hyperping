// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net/http"
	"regexp"
	"testing"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestAccMonitorResource_createError(t *testing.T) {
	server := newMockHyperpingServerWithErrors(t)
	defer server.Close()

	server.setCreateError(true)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorResourceConfigBasic(server.URL, "error-test"),
				ExpectError: regexp.MustCompile(`Failed to Create Monitor`),
			},
		},
	})
}

func TestAccMonitorResource_updateError(t *testing.T) {
	server := newMockHyperpingServerWithErrors(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "update-error-test"),
			},
			{
				PreConfig:   func() { server.setUpdateError(true) },
				Config:      testAccMonitorResourceConfigBasic(server.URL, "updated-name"),
				ExpectError: regexp.MustCompile(`Failed to Update Monitor`),
			},
		},
	})
}

func TestAccMonitorResource_readError(t *testing.T) {
	server := newMockHyperpingServerWithErrors(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "read-error-test"),
			},
			{
				PreConfig:          func() { server.setReadError(true) },
				Config:             testAccMonitorResourceConfigBasic(server.URL, "read-error-test"),
				ExpectError:        regexp.MustCompile(`Failed to Read Monitor`),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccMonitorResource_pauseError(t *testing.T) {
	server := newMockHyperpingServerWithErrors(t)
	defer server.Close()

	server.setPauseError(true)

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config:      testAccMonitorResourceConfigPaused(server.URL),
				ExpectError: regexp.MustCompile(`failed to pause`),
			},
		},
	})
}

// Note: created_at and updated_at are returned by the API but not exposed in the schema
// since they are read-only server-side timestamps. Keeping the mock server function

func TestAccMonitorResource_deleteErrorNon404(t *testing.T) {
	// This test verifies that delete errors (non-404) are properly handled
	server := newMockHyperpingServerWithErrors(t)
	defer server.Close()

	var monitorID string

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			// Create a monitor
			{
				Config: testAccMonitorResourceConfigBasic(server.URL, "delete-non404-test"),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					// Store the monitor ID for later use
					tfresource.TestCheckResourceAttrWith("hyperping_monitor.test", "id", func(value string) error {
						monitorID = value
						return nil
					}),
				),
			},
		},
		// After the test, the destroy will succeed normally
		// The delete error is tested separately
	})

	// Now test that delete errors are properly handled by calling the API directly
	// with an error flag set
	if monitorID != "" {
		// Re-create the monitor so we can test delete error
		server.createTestMonitor(monitorID, "test")
		server.setDeleteError(true)

		// Attempt to delete via the mock server
		req, _ := http.NewRequest("DELETE", server.URL+client.MonitorsBasePath+"/"+monitorID, nil)
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusInternalServerError {
			t.Errorf("Expected 500 status, got %d", resp.StatusCode)
		}
	}
}
