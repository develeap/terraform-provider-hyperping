// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	tfresource "github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// T31: Location metadata correctness (unit test)
func TestMonitoringLocations_MetadataCorrectness(t *testing.T) {
	tests := []struct {
		id          string
		name        string
		continent   string
		cloudRegion string
	}{
		{"london", "London, UK", "Europe", "eu-west-2"},
		{"frankfurt", "Frankfurt, DE", "Europe", "eu-central-1"},
		{"singapore", "Singapore", "Asia Pacific", "ap-southeast-1"},
		{"sydney", "Sydney, AU", "Asia Pacific", "ap-southeast-2"},
		{"tokyo", "Tokyo, JP", "Asia Pacific", "ap-northeast-1"},
		{"virginia", "Virginia, US", "North America", "us-east-1"},
		{"saopaulo", "Sao Paulo, BR", "South America", "sa-east-1"},
		{"bahrain", "Bahrain, ME", "Middle East", "me-south-1"},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			meta, ok := monitoringLocations[tt.id]
			if !ok {
				t.Fatalf("missing monitoring location for region %q", tt.id)
			}
			if meta.Name != tt.name {
				t.Errorf("expected name %q, got %q", tt.name, meta.Name)
			}
			if meta.Continent != tt.continent {
				t.Errorf("expected continent %q, got %q", tt.continent, meta.Continent)
			}
			if meta.CloudRegion != tt.cloudRegion {
				t.Errorf("expected cloud_region %q, got %q", tt.cloudRegion, meta.CloudRegion)
			}
		})
	}
}

// T32: All client.AllowedRegions are covered (invariant test)
func TestMonitoringLocations_AllRegionsCovered(t *testing.T) {
	// Every entry in client.AllowedRegions must have a corresponding monitoring location
	for _, region := range client.AllowedRegions {
		if _, ok := monitoringLocations[region]; !ok {
			t.Errorf("client.AllowedRegions contains %q but monitoringLocations does not", region)
		}
	}

	// No extra entries in monitoringLocations beyond AllowedRegions
	allowedSet := make(map[string]bool, len(client.AllowedRegions))
	for _, r := range client.AllowedRegions {
		allowedSet[r] = true
	}
	for id := range monitoringLocations {
		if !allowedSet[id] {
			t.Errorf("monitoringLocations contains %q but client.AllowedRegions does not", id)
		}
	}

	// Verify 1:1 count
	if len(monitoringLocations) != len(client.AllowedRegions) {
		t.Errorf("expected %d locations, got %d", len(client.AllowedRegions), len(monitoringLocations))
	}
}

// T33: Data source Metadata returns correct type name
func TestMonitoringLocationsDataSource_Metadata(t *testing.T) {
	d := &MonitoringLocationsDataSource{}

	req := datasource.MetadataRequest{ProviderTypeName: "hyperping"}
	resp := &datasource.MetadataResponse{}

	d.Metadata(context.Background(), req, resp)

	if resp.TypeName != "hyperping_monitoring_locations" {
		t.Errorf("expected type name 'hyperping_monitoring_locations', got %q", resp.TypeName)
	}
}

// T34: Data source Schema has required attributes
func TestMonitoringLocationsDataSource_Schema(t *testing.T) {
	d := &MonitoringLocationsDataSource{}

	req := datasource.SchemaRequest{}
	resp := &datasource.SchemaResponse{}

	d.Schema(context.Background(), req, resp)

	if _, ok := resp.Schema.Attributes["locations"]; !ok {
		t.Error("schema missing 'locations' attribute")
	}
	if _, ok := resp.Schema.Attributes["ids"]; !ok {
		t.Error("schema missing 'ids' attribute")
	}
}

// T29 + T30: Acceptance test for monitoring locations data source
func TestAccMonitoringLocationsDataSource_basic(t *testing.T) {
	server := newMinimalMockServer(t)
	defer server.Close()

	tfresource.ParallelTest(t, tfresource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []tfresource.TestStep{
			{
				Config: testAccMonitoringLocationsDataSourceConfig(server.URL),
				Check: tfresource.ComposeAggregateTestCheckFunc(
					// T29: All 8 locations returned
					tfresource.TestCheckResourceAttr("data.hyperping_monitoring_locations.all", "locations.#", "8"),
					tfresource.TestCheckResourceAttr("data.hyperping_monitoring_locations.all", "ids.#", "8"),
					// T30: Known IDs present
					tfresource.TestCheckTypeSetElemAttr("data.hyperping_monitoring_locations.all", "ids.*", "london"),
					tfresource.TestCheckTypeSetElemAttr("data.hyperping_monitoring_locations.all", "ids.*", "frankfurt"),
					tfresource.TestCheckTypeSetElemAttr("data.hyperping_monitoring_locations.all", "ids.*", "singapore"),
					tfresource.TestCheckTypeSetElemAttr("data.hyperping_monitoring_locations.all", "ids.*", "sydney"),
					tfresource.TestCheckTypeSetElemAttr("data.hyperping_monitoring_locations.all", "ids.*", "tokyo"),
					tfresource.TestCheckTypeSetElemAttr("data.hyperping_monitoring_locations.all", "ids.*", "virginia"),
					tfresource.TestCheckTypeSetElemAttr("data.hyperping_monitoring_locations.all", "ids.*", "saopaulo"),
					tfresource.TestCheckTypeSetElemAttr("data.hyperping_monitoring_locations.all", "ids.*", "bahrain"),
				),
			},
		},
	})
}

func testAccMonitoringLocationsDataSourceConfig(baseURL string) string {
	return fmt.Sprintf(`
provider "hyperping" {
  api_key  = "test_api_key"
  base_url = %[1]q
}

data "hyperping_monitoring_locations" "all" {
}
`, baseURL)
}
