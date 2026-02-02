// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"
	"time"

	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccMaintenanceWindowsDataSource_basic(t *testing.T) {
	server := newMockMaintenanceServer(t)
	defer server.Close()

	now := time.Now().UTC()

	// Pre-create maintenance windows
	for i := 1; i <= 2; i++ {
		id := fmt.Sprintf("mw%d", i)
		start := now.Add(time.Duration(24*i) * time.Hour).Format(time.RFC3339)
		end := now.Add(time.Duration(24*i+2) * time.Hour).Format(time.RFC3339)

		server.maintenanceWindows[id] = map[string]interface{}{
			"uuid":       id,
			"name":       fmt.Sprintf("Maintenance %d", i),
			"title":      map[string]interface{}{"en": fmt.Sprintf("Title %d", i)},
			"text":       map[string]interface{}{"en": fmt.Sprintf("Description %d", i)},
			"start_date": start,
			"end_date":   end,
			"monitors":   []string{"mon_123"},
		}
	}

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMaintenanceWindowsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_maintenance_windows.all", "maintenance_windows.#", "2"),
				),
			},
		},
	})
}

func TestAccMaintenanceWindowsDataSource_empty(t *testing.T) {
	server := newMockMaintenanceServer(t)
	defer server.Close()

	tfresource.Test(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMaintenanceWindowsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					tfresource.TestCheckResourceAttr("data.hyperping_maintenance_windows.all", "maintenance_windows.#", "0"),
				),
			},
		},
	})
}

func testAccMaintenanceWindowsDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_maintenance_windows" "all" {
}
`, baseURL)
}
