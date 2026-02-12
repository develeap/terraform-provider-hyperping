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
var (
	_ datasource.DataSource              = &OutagesDataSource{}
	_ datasource.DataSourceWithConfigure = &OutagesDataSource{}
)

// NewOutagesDataSource creates a new outages data source.
func NewOutagesDataSource() datasource.DataSource {
	return &OutagesDataSource{}
}

// OutagesDataSource defines the data source implementation.
type OutagesDataSource struct {
	client client.OutageAPI
}

// OutagesDataSourceModel describes the data source data model.
type OutagesDataSourceModel struct {
	Outages []OutageDataModel  `tfsdk:"outages"`
	Filter  *OutageFilterModel `tfsdk:"filter"`
}

// OutageDataModel describes a single outage in the list data source.
type OutageDataModel struct {
	ID               types.String `tfsdk:"id"`
	MonitorUUID      types.String `tfsdk:"monitor_uuid"`
	StartDate        types.String `tfsdk:"start_date"`
	EndDate          types.String `tfsdk:"end_date"`
	StatusCode       types.Int64  `tfsdk:"status_code"`
	Description      types.String `tfsdk:"description"`
	OutageType       types.String `tfsdk:"outage_type"`
	IsResolved       types.Bool   `tfsdk:"is_resolved"`
	DurationMs       types.Int64  `tfsdk:"duration_ms"`
	DetectedLocation types.String `tfsdk:"detected_location"`
	Monitor          types.Object `tfsdk:"monitor"`
	AcknowledgedBy   types.Object `tfsdk:"acknowledged_by"`
}

// Metadata returns the data source type name.
func (d *OutagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_outages"
}

// Schema defines the schema for the data source.
func (d *OutagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of all Hyperping outages.",

		Attributes: map[string]schema.Attribute{
			"filter": OutageFilterSchema(),
			"outages": schema.ListNestedAttribute{
				MarkdownDescription: "List of outages.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier (UUID) of the outage.",
							Computed:            true,
						},
						"monitor_uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the monitor associated with this outage.",
							Computed:            true,
						},
						"start_date": schema.StringAttribute{
							MarkdownDescription: "The start date of the outage in ISO 8601 format.",
							Computed:            true,
						},
						"end_date": schema.StringAttribute{
							MarkdownDescription: "The end date of the outage in ISO 8601 format. Null if ongoing.",
							Computed:            true,
						},
						"status_code": schema.Int64Attribute{
							MarkdownDescription: "The HTTP status code that caused the outage.",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Description of the outage.",
							Computed:            true,
						},
						"outage_type": schema.StringAttribute{
							MarkdownDescription: "The type of outage (e.g. `manual` or `automatic`).",
							Computed:            true,
						},
						"is_resolved": schema.BoolAttribute{
							MarkdownDescription: "Whether the outage is resolved.",
							Computed:            true,
						},
						"duration_ms": schema.Int64Attribute{
							MarkdownDescription: "Duration of the outage in milliseconds.",
							Computed:            true,
						},
						"detected_location": schema.StringAttribute{
							MarkdownDescription: "The location that detected the outage.",
							Computed:            true,
						},
						"monitor": schema.SingleNestedAttribute{
							MarkdownDescription: "The monitor associated with this outage.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"uuid": schema.StringAttribute{
									MarkdownDescription: "The UUID of the monitor.",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "The name of the monitor.",
									Computed:            true,
								},
								"url": schema.StringAttribute{
									MarkdownDescription: "The URL of the monitor.",
									Computed:            true,
								},
								"protocol": schema.StringAttribute{
									MarkdownDescription: "The protocol of the monitor.",
									Computed:            true,
								},
							},
						},
						"acknowledged_by": schema.SingleNestedAttribute{
							MarkdownDescription: "The user who acknowledged this outage. Null if not acknowledged.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"uuid": schema.StringAttribute{
									MarkdownDescription: "The UUID of the user.",
									Computed:            true,
								},
								"email": schema.StringAttribute{
									MarkdownDescription: "The email of the user.",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "The name of the user.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *OutagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = c
}

// Read refreshes the Terraform state with the latest data.
func (d *OutagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config OutagesDataSourceModel

	// Get configuration (includes filter if provided)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outages, err := d.client.ListOutages(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading outages",
			fmt.Sprintf("Could not list outages: %s", err),
		)
		return
	}

	// Apply client-side filtering if filter provided
	var filteredOutages []client.Outage
	if config.Filter != nil {
		for _, outage := range outages {
			if d.shouldIncludeOutage(&outage, config.Filter, &resp.Diagnostics) {
				if resp.Diagnostics.HasError() {
					return
				}
				filteredOutages = append(filteredOutages, outage)
			}
		}
	} else {
		filteredOutages = outages
	}

	config.Outages = make([]OutageDataModel, len(filteredOutages))
	for i, outage := range filteredOutages {
		d.mapOutageToDataModel(&outage, &config.Outages[i], &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// shouldIncludeOutage determines if an outage matches the filter criteria.
func (d *OutagesDataSource) shouldIncludeOutage(outage *client.Outage, filter *OutageFilterModel, diags *diag.Diagnostics) bool {
	return ApplyAllFilters(
		// Name regex filter (matches monitor name)
		func() bool {
			match, err := MatchesNameRegex(outage.Monitor.Name, filter.NameRegex)
			if err != nil {
				diags.AddError(
					"Invalid filter regex",
					fmt.Sprintf("Failed to compile name_regex pattern: %s", err),
				)
				return false
			}
			return match
		},
		// Monitor UUID filter
		func() bool {
			return MatchesExact(outage.Monitor.UUID, filter.MonitorUUID)
		},
	)
}

// mapOutageToDataModel maps a client.Outage to the list data model
// using the shared MapOutageNestedObjects helper for nested monitor/acknowledged_by.
func (d *OutagesDataSource) mapOutageToDataModel(outage *client.Outage, model *OutageDataModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(outage.UUID)
	model.MonitorUUID = types.StringValue(outage.Monitor.UUID)
	model.StartDate = types.StringValue(outage.StartDate)
	model.StatusCode = types.Int64Value(int64(outage.StatusCode))
	model.Description = types.StringValue(outage.Description)
	model.OutageType = types.StringValue(outage.OutageType)
	model.IsResolved = types.BoolValue(outage.IsResolved)
	model.DurationMs = types.Int64Value(int64(outage.DurationMs))
	model.DetectedLocation = types.StringValue(outage.DetectedLocation)

	if outage.EndDate != nil {
		model.EndDate = types.StringValue(*outage.EndDate)
	} else {
		model.EndDate = types.StringNull()
	}

	model.Monitor, model.AcknowledgedBy = MapOutageNestedObjects(outage, diags)
}
