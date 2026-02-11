// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &MonitorReportDataSource{}
	_ datasource.DataSourceWithConfigure = &MonitorReportDataSource{}
)

// NewMonitorReportDataSource creates a new monitor report data source.
func NewMonitorReportDataSource() datasource.DataSource {
	return &MonitorReportDataSource{}
}

// MonitorReportDataSource defines the data source implementation.
type MonitorReportDataSource struct {
	client client.ReportsAPI
}

// MonitorReportDataSourceModel describes the data source data model.
type MonitorReportDataSourceModel struct {
	ID       types.String  `tfsdk:"id"`
	From     types.String  `tfsdk:"from"`
	To       types.String  `tfsdk:"to"`
	Name     types.String  `tfsdk:"name"`
	Protocol types.String  `tfsdk:"protocol"`
	SLA      types.Float64 `tfsdk:"sla"`
	Outages  types.Object  `tfsdk:"outages"`
}

// OutageAttrTypes returns the attribute types for the outages object.
func OutageAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"count":                    types.Int64Type,
		"total_downtime":           types.Int64Type,
		"total_downtime_formatted": types.StringType,
		"longest_outage":           types.Int64Type,
		"longest_outage_formatted": types.StringType,
	}
}

// Metadata returns the data source type name.
func (d *MonitorReportDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_report"
}

// Schema defines the schema for the data source.
func (d *MonitorReportDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the SLA and outage report for a specific Hyperping monitor.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the monitor to get the report for.",
				Required:            true,
			},
			"from": schema.StringAttribute{
				MarkdownDescription: "Start date for the report period in ISO 8601 format. Optional.",
				Optional:            true,
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "End date for the report period in ISO 8601 format. Optional.",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the monitor.",
				Computed:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The protocol of the monitor.",
				Computed:            true,
			},
			"sla": schema.Float64Attribute{
				MarkdownDescription: "The SLA percentage for the report period (e.g., 99.95).",
				Computed:            true,
			},
			"outages": schema.SingleNestedAttribute{
				MarkdownDescription: "Outage statistics for the report period.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"count": schema.Int64Attribute{
						MarkdownDescription: "Total number of outages.",
						Computed:            true,
					},
					"total_downtime": schema.Int64Attribute{
						MarkdownDescription: "Total downtime in seconds.",
						Computed:            true,
					},
					"total_downtime_formatted": schema.StringAttribute{
						MarkdownDescription: "Human-readable total downtime (e.g., '1hr 10min 31s').",
						Computed:            true,
					},
					"longest_outage": schema.Int64Attribute{
						MarkdownDescription: "Longest single outage in seconds.",
						Computed:            true,
					},
					"longest_outage_formatted": schema.StringAttribute{
						MarkdownDescription: "Human-readable longest outage.",
						Computed:            true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *MonitorReportDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *MonitorReportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MonitorReportDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get optional date range
	from := ""
	to := ""
	if !config.From.IsNull() {
		from = config.From.ValueString()
	}
	if !config.To.IsNull() {
		to = config.To.ValueString()
	}

	if err := client.ValidateResourceID(config.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid Monitor ID", fmt.Sprintf("Cannot look up monitor report: %s", err))
		return
	}

	report, err := d.client.GetMonitorReport(ctx, config.ID.ValueString(), from, to)
	if err != nil {
		resp.Diagnostics.Append(newReadError("Monitor Report", config.ID.ValueString(), err))
		return
	}

	d.mapReportToDataSourceModel(report, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// mapReportToDataSourceModel maps a client.MonitorReport to the data source model.
func (d *MonitorReportDataSource) mapReportToDataSourceModel(report *client.MonitorReport, model *MonitorReportDataSourceModel, diags *diag.Diagnostics) {
	model.Name = types.StringValue(report.Name)
	model.Protocol = types.StringValue(report.Protocol)
	model.SLA = types.Float64Value(report.SLA)

	// Map outages object
	outagesObj, objDiags := types.ObjectValue(OutageAttrTypes(), map[string]attr.Value{
		"count":                    types.Int64Value(int64(report.Outages.Count)),
		"total_downtime":           types.Int64Value(int64(report.Outages.TotalDowntime)),
		"total_downtime_formatted": types.StringValue(report.Outages.TotalDowntimeFormatted),
		"longest_outage":           types.Int64Value(int64(report.Outages.LongestOutage)),
		"longest_outage_formatted": types.StringValue(report.Outages.LongestOutageFormatted),
	})
	diags.Append(objDiags...)
	model.Outages = outagesObj
}
