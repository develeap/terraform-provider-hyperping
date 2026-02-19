// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &MonitorReportsDataSource{}
	_ datasource.DataSourceWithConfigure = &MonitorReportsDataSource{}
)

// NewMonitorReportsDataSource creates a new monitor reports data source.
func NewMonitorReportsDataSource() datasource.DataSource {
	return &MonitorReportsDataSource{}
}

// MonitorReportsDataSource defines the data source implementation.
type MonitorReportsDataSource struct {
	client client.ReportsAPI
}

// MonitorReportsDataSourceModel describes the data source data model.
type MonitorReportsDataSourceModel struct {
	From     types.String `tfsdk:"from"`
	To       types.String `tfsdk:"to"`
	Monitors types.List   `tfsdk:"monitors"`
}

// MonitorReportListItemModel describes a single monitor in the reports list.
type MonitorReportListItemModel struct {
	ID                     types.String  `tfsdk:"id"`
	Name                   types.String  `tfsdk:"name"`
	Protocol               types.String  `tfsdk:"protocol"`
	SLA                    types.Float64 `tfsdk:"sla"`
	MTTR                   types.Int64   `tfsdk:"mttr"`
	MTTRFormatted          types.String  `tfsdk:"mttr_formatted"`
	OutageCount            types.Int64   `tfsdk:"outage_count"`
	TotalDowntime          types.Int64   `tfsdk:"total_downtime"`
	TotalDowntimeFormatted types.String  `tfsdk:"total_downtime_formatted"`
}

// monitorReportListItemAttrTypes returns the attribute types for a MonitorReportListItemModel.
func monitorReportListItemAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                       types.StringType,
		"name":                     types.StringType,
		"protocol":                 types.StringType,
		"sla":                      types.Float64Type,
		"mttr":                     types.Int64Type,
		"mttr_formatted":           types.StringType,
		"outage_count":             types.Int64Type,
		"total_downtime":           types.Int64Type,
		"total_downtime_formatted": types.StringType,
	}
}

// Metadata returns the data source type name.
func (d *MonitorReportsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitor_reports"
}

// Schema defines the schema for the data source.
func (d *MonitorReportsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches SLA and outage reports for all Hyperping monitors.",

		Attributes: map[string]schema.Attribute{
			"from": schema.StringAttribute{
				MarkdownDescription: "Start of reporting period (ISO 8601). Defaults to 7 days ago.",
				Optional:            true,
			},
			"to": schema.StringAttribute{
				MarkdownDescription: "End of reporting period (ISO 8601). Defaults to now.",
				Optional:            true,
			},
			"monitors": schema.ListNestedAttribute{
				MarkdownDescription: "List of monitor reports.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The UUID of the monitor.",
							Computed:            true,
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
							MarkdownDescription: "Uptime percentage for the reporting period (e.g. 99.95).",
							Computed:            true,
						},
						"mttr": schema.Int64Attribute{
							MarkdownDescription: "Mean time to recovery in seconds.",
							Computed:            true,
						},
						"mttr_formatted": schema.StringAttribute{
							MarkdownDescription: "Human-readable mean time to recovery (e.g. \"35min 16s\").",
							Computed:            true,
						},
						"outage_count": schema.Int64Attribute{
							MarkdownDescription: "Total number of outages in the reporting period.",
							Computed:            true,
						},
						"total_downtime": schema.Int64Attribute{
							MarkdownDescription: "Total downtime in seconds.",
							Computed:            true,
						},
						"total_downtime_formatted": schema.StringAttribute{
							MarkdownDescription: "Human-readable total downtime (e.g. \"2min 0s\").",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *MonitorReportsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *MonitorReportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MonitorReportsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	from := config.From.ValueString()
	to := config.To.ValueString()

	reports, err := d.client.ListMonitorReports(ctx, from, to)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading monitor reports",
			fmt.Sprintf("Could not list monitor reports: %s", err),
		)
		return
	}

	items := mapReportsToListItems(reports)
	monitorsListValue, listDiags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: monitorReportListItemAttrTypes()}, items)
	resp.Diagnostics.Append(listDiags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Monitors = monitorsListValue

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// mapReportsToListItems maps a slice of client.MonitorReport to MonitorReportListItemModel slice.
func mapReportsToListItems(reports []client.MonitorReport) []MonitorReportListItemModel {
	items := make([]MonitorReportListItemModel, len(reports))
	for i, r := range reports {
		items[i] = MonitorReportListItemModel{
			ID:                     types.StringValue(r.UUID),
			Name:                   types.StringValue(r.Name),
			Protocol:               types.StringValue(r.Protocol),
			SLA:                    types.Float64Value(r.SLA),
			MTTR:                   types.Int64Value(int64(r.MTTR)),
			MTTRFormatted:          types.StringValue(r.MTTRFormatted),
			OutageCount:            types.Int64Value(int64(r.Outages.Count)),
			TotalDowntime:          types.Int64Value(int64(r.Outages.TotalDowntime)),
			TotalDowntimeFormatted: types.StringValue(r.Outages.TotalDowntimeFormatted),
		}
	}
	return items
}
