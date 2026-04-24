// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hyperping "github.com/develeap/hyperping-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &OnCallScheduleDataSource{}
	_ datasource.DataSourceWithConfigure = &OnCallScheduleDataSource{}
)

// NewOnCallScheduleDataSource creates a new single on-call schedule data source.
func NewOnCallScheduleDataSource() datasource.DataSource {
	return &OnCallScheduleDataSource{}
}

// OnCallScheduleDataSource defines the data source implementation.
type OnCallScheduleDataSource struct {
	client *hyperping.MCPClient
}

// OnCallScheduleDataSourceModel describes the data source data model.
type OnCallScheduleDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Team          types.String `tfsdk:"team"`
	CurrentOnCall types.String `tfsdk:"current_oncall"`
	NextOnCall    types.String `tfsdk:"next_oncall"`
	RotationStart types.String `tfsdk:"rotation_start"`
	RotationEnd   types.String `tfsdk:"rotation_end"`
}

// Metadata returns the data source type name.
func (d *OnCallScheduleDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_schedule"
}

// Schema defines the schema for the data source.
func (d *OnCallScheduleDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single Hyperping on-call schedule by ID or name via MCP.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the schedule.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the schedule to look up.",
				Optional:            true,
				Computed:            true,
			},
			"team": schema.StringAttribute{
				MarkdownDescription: "The team name this schedule belongs to.",
				Computed:            true,
			},
			"current_oncall": schema.StringAttribute{
				MarkdownDescription: "The name of the person currently on-call.",
				Computed:            true,
			},
			"next_oncall": schema.StringAttribute{
				MarkdownDescription: "The name of the person next on-call.",
				Computed:            true,
			},
			"rotation_start": schema.StringAttribute{
				MarkdownDescription: "Start time of the current rotation.",
				Computed:            true,
			},
			"rotation_end": schema.StringAttribute{
				MarkdownDescription: "End time of the current rotation.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *OnCallScheduleDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*hyperpingClients)
	if !ok {
		resp.Diagnostics.Append(newUnexpectedConfigTypeError("*hyperpingClients", req.ProviderData))
		return
	}

	d.client = clients.MCP
}

// Read refreshes the Terraform state with the latest data.
func (d *OnCallScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data OnCallScheduleDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedules, err := d.client.ListOnCallSchedules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching on-call schedules", err.Error())
		return
	}

	var found *hyperping.OnCallSchedule
	if !data.ID.IsNull() {
		id := data.ID.ValueString()
		for _, s := range schedules {
			if s.UUID == id {
				found = &s
				break
			}
		}
	} else if !data.Name.IsNull() {
		name := data.Name.ValueString()
		for _, s := range schedules {
			if s.Name == name {
				found = &s
				break
			}
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"On-Call Schedule Not Found",
			"Could not find an on-call schedule matching the provided criteria.",
		)
		return
	}

	data.ID = types.StringValue(found.UUID)
	data.Name = types.StringValue(found.Name)
	data.Team = types.StringValue(found.Team)
	data.CurrentOnCall = types.StringValue(found.CurrentOncall)
	data.NextOnCall = types.StringValue(found.NextOncall)
	data.RotationStart = types.StringValue(found.RotationStart)
	data.RotationEnd = types.StringValue(found.RotationEnd)

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
