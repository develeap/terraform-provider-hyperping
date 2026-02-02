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
	_ datasource.DataSource              = &MonitorsDataSource{}
	_ datasource.DataSourceWithConfigure = &MonitorsDataSource{}
)

// NewMonitorsDataSource creates a new monitors data source.
func NewMonitorsDataSource() datasource.DataSource {
	return &MonitorsDataSource{}
}

// MonitorsDataSource defines the data source implementation.
type MonitorsDataSource struct {
	client client.MonitorAPI
}

// MonitorsDataSourceModel describes the data source data model.
type MonitorsDataSourceModel struct {
	Monitors []MonitorDataModel `tfsdk:"monitors"`
}

// MonitorDataModel describes a single monitor in the data source.
type MonitorDataModel struct {
	ID                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	URL                types.String `tfsdk:"url"`
	Protocol           types.String `tfsdk:"protocol"`
	HTTPMethod         types.String `tfsdk:"http_method"`
	CheckFrequency     types.Int64  `tfsdk:"check_frequency"`
	Regions            types.List   `tfsdk:"regions"`
	RequestHeaders     types.List   `tfsdk:"request_headers"`
	RequestBody        types.String `tfsdk:"request_body"`
	ExpectedStatusCode types.String `tfsdk:"expected_status_code"`
	FollowRedirects    types.Bool   `tfsdk:"follow_redirects"`
	Paused             types.Bool   `tfsdk:"paused"`
	Port               types.Int64  `tfsdk:"port"`
	AlertsWait         types.Int64  `tfsdk:"alerts_wait"`
	EscalationPolicy   types.String `tfsdk:"escalation_policy"`
	RequiredKeyword    types.String `tfsdk:"required_keyword"`
}

// Metadata returns the data source type name.
func (d *MonitorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

// Schema defines the schema for the data source.
func (d *MonitorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of all Hyperping monitors.",

		Attributes: map[string]schema.Attribute{
			"monitors": schema.ListNestedAttribute{
				MarkdownDescription: "List of monitors.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier (UUID) of the monitor.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the monitor.",
							Computed:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "The URL being monitored.",
							Computed:            true,
						},
						"protocol": schema.StringAttribute{
							MarkdownDescription: "The protocol used for monitoring (http, icmp, tcp, udp).",
							Computed:            true,
						},
						"http_method": schema.StringAttribute{
							MarkdownDescription: "HTTP method used for the check (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS).",
							Computed:            true,
						},
						"check_frequency": schema.Int64Attribute{
							MarkdownDescription: "Check frequency in seconds.",
							Computed:            true,
						},
						"regions": schema.ListAttribute{
							MarkdownDescription: "List of regions the monitor checks from.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"request_headers": schema.ListNestedAttribute{
							MarkdownDescription: "Custom HTTP headers sent with the request.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "Header name.",
										Computed:            true,
									},
									"value": schema.StringAttribute{
										MarkdownDescription: "Header value.",
										Computed:            true,
									},
								},
							},
						},
						"request_body": schema.StringAttribute{
							MarkdownDescription: "Request body for POST/PUT/PATCH requests.",
							Computed:            true,
						},
						"expected_status_code": schema.StringAttribute{
							MarkdownDescription: "Expected HTTP status code or pattern (e.g., `200`, `2xx`).",
							Computed:            true,
						},
						"follow_redirects": schema.BoolAttribute{
							MarkdownDescription: "Whether to follow HTTP redirects.",
							Computed:            true,
						},
						"paused": schema.BoolAttribute{
							MarkdownDescription: "Whether the monitor is paused.",
							Computed:            true,
						},
						"port": schema.Int64Attribute{
							MarkdownDescription: "Port number for port protocol monitors.",
							Computed:            true,
						},
						"alerts_wait": schema.Int64Attribute{
							MarkdownDescription: "Seconds to wait before sending alerts after an outage is detected.",
							Computed:            true,
						},
						"escalation_policy": schema.StringAttribute{
							MarkdownDescription: "UUID of the escalation policy linked to this monitor.",
							Computed:            true,
						},
						"required_keyword": schema.StringAttribute{
							MarkdownDescription: "Keyword that must appear in the response body.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *MonitorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *MonitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state MonitorsDataSourceModel

	// Fetch all monitors from API
	monitors, err := d.client.ListMonitors(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading monitors",
			fmt.Sprintf("Could not list monitors: %s", err),
		)
		return
	}

	// Map response to model
	state.Monitors = make([]MonitorDataModel, len(monitors))
	for i, monitor := range monitors {
		d.mapMonitorToDataModel(&monitor, &state.Monitors[i], &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// mapMonitorToDataModel maps a client.Monitor to the Terraform data model.
func (d *MonitorsDataSource) mapMonitorToDataModel(monitor *client.Monitor, model *MonitorDataModel, diags *diag.Diagnostics) {
	fields := MapMonitorCommonFields(monitor, diags)

	model.ID = fields.ID
	model.Name = fields.Name
	model.URL = fields.URL
	model.Protocol = fields.Protocol
	model.HTTPMethod = fields.HTTPMethod
	model.CheckFrequency = fields.CheckFrequency
	model.ExpectedStatusCode = fields.ExpectedStatusCode
	model.FollowRedirects = fields.FollowRedirects
	model.Paused = fields.Paused
	model.Regions = fields.Regions
	model.RequestHeaders = fields.RequestHeaders
	model.RequestBody = fields.RequestBody
	model.Port = fields.Port
	model.AlertsWait = fields.AlertsWait
	model.EscalationPolicy = fields.EscalationPolicy
	model.RequiredKeyword = fields.RequiredKeyword
}
