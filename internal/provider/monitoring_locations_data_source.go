// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &MonitoringLocationsDataSource{}

// NewMonitoringLocationsDataSource creates a new monitoring locations data source.
func NewMonitoringLocationsDataSource() datasource.DataSource {
	return &MonitoringLocationsDataSource{}
}

// MonitoringLocationsDataSource returns available monitoring regions.
// This is a static data source that does not make API calls.
type MonitoringLocationsDataSource struct{}

// MonitoringLocationsDataSourceModel describes the data source data model.
type MonitoringLocationsDataSourceModel struct {
	Locations []MonitoringLocationModel `tfsdk:"locations"`
	IDs       types.List                `tfsdk:"ids"`
}

// MonitoringLocationModel describes a single monitoring location.
type MonitoringLocationModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Continent   types.String `tfsdk:"continent"`
	CloudRegion types.String `tfsdk:"cloud_region"`
}

// monitoringLocationMetadata holds enrichment data for a monitoring region.
type monitoringLocationMetadata struct {
	Name        string
	Continent   string
	CloudRegion string
}

// monitoringLocations maps region codes from client.AllowedRegions to metadata.
var monitoringLocations = map[string]monitoringLocationMetadata{
	"london":    {Name: "London, UK", Continent: "Europe", CloudRegion: "eu-west-2"},
	"frankfurt": {Name: "Frankfurt, DE", Continent: "Europe", CloudRegion: "eu-central-1"},
	"singapore": {Name: "Singapore", Continent: "Asia Pacific", CloudRegion: "ap-southeast-1"},
	"sydney":    {Name: "Sydney, AU", Continent: "Asia Pacific", CloudRegion: "ap-southeast-2"},
	"tokyo":     {Name: "Tokyo, JP", Continent: "Asia Pacific", CloudRegion: "ap-northeast-1"},
	"virginia":  {Name: "Virginia, US", Continent: "North America", CloudRegion: "us-east-1"},
	"saopaulo":  {Name: "Sao Paulo, BR", Continent: "South America", CloudRegion: "sa-east-1"},
	"bahrain":   {Name: "Bahrain, ME", Continent: "Middle East", CloudRegion: "me-south-1"},
}

// Metadata returns the data source type name.
func (d *MonitoringLocationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitoring_locations"
}

// Schema defines the schema for the data source.
func (d *MonitoringLocationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns available monitoring locations (regions) for Hyperping monitors. " +
			"This data source is static and does not require an API call. " +
			"The region list reflects the current provider version.",

		Attributes: map[string]schema.Attribute{
			"locations": schema.ListNestedAttribute{
				MarkdownDescription: "List of available monitoring locations with metadata.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Region code (e.g., `london`). Use this value in `hyperping_monitor.regions`.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Display name of the location (e.g., `London, UK`).",
							Computed:            true,
						},
						"continent": schema.StringAttribute{
							MarkdownDescription: "Continent grouping (e.g., `Europe`, `Asia Pacific`).",
							Computed:            true,
						},
						"cloud_region": schema.StringAttribute{
							MarkdownDescription: "Approximate cloud provider region identifier (e.g., `eu-west-2`).",
							Computed:            true,
						},
					},
				},
			},
			"ids": schema.ListAttribute{
				MarkdownDescription: "List of region codes. Convenient for `for_each` patterns or as input to `hyperping_monitor.regions`.",
				Computed:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

// Read populates the data source model with static location data.
func (d *MonitoringLocationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model MonitoringLocationsDataSourceModel

	// Build locations list in a deterministic order from client.AllowedRegions
	model.Locations = make([]MonitoringLocationModel, len(client.AllowedRegions))
	ids := make([]string, len(client.AllowedRegions))

	for i, regionID := range client.AllowedRegions {
		meta, ok := monitoringLocations[regionID]
		if !ok {
			resp.Diagnostics.Append(diag.NewWarningDiagnostic(
				"Missing Region Metadata",
				fmt.Sprintf("Region %q from client.AllowedRegions has no metadata entry. "+
					"It will appear with empty name, continent, and cloud_region.", regionID),
			))
		}
		model.Locations[i] = MonitoringLocationModel{
			ID:          types.StringValue(regionID),
			Name:        types.StringValue(meta.Name),
			Continent:   types.StringValue(meta.Continent),
			CloudRegion: types.StringValue(meta.CloudRegion),
		}
		ids[i] = regionID
	}

	// Build the ids list
	idsList, diags := types.ListValueFrom(ctx, types.StringType, ids)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	model.IDs = idsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &model)...)
}
