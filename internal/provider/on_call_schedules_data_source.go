// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hyperping "github.com/develeap/hyperping-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &OnCallSchedulesDataSource{}
	_ datasource.DataSourceWithConfigure = &OnCallSchedulesDataSource{}
)

// NewOnCallSchedulesDataSource creates a new on-call schedules data source.
func NewOnCallSchedulesDataSource() datasource.DataSource {
	return &OnCallSchedulesDataSource{}
}

// OnCallSchedulesDataSource defines the data source implementation.
type OnCallSchedulesDataSource struct {
	client *hyperping.MCPClient
}

// OnCallSchedulesDataSourceModel describes the data source data model.
type OnCallSchedulesDataSourceModel struct {
	Schedules []OnCallScheduleDataModel `tfsdk:"schedules"`
	IDs       types.List                `tfsdk:"ids"`
}

// OnCallScheduleDataModel describes a single on-call schedule.
type OnCallScheduleDataModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Team          types.String `tfsdk:"team"`
	CurrentOnCall types.String `tfsdk:"current_oncall"`
	NextOnCall    types.String `tfsdk:"next_oncall"`
	RotationStart types.String `tfsdk:"rotation_start"`
	RotationEnd   types.String `tfsdk:"rotation_end"`
}

// Metadata returns the data source type name.
func (d *OnCallSchedulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_on_call_schedules"
}

// Schema defines the schema for the data source.
func (d *OnCallSchedulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of all Hyperping on-call schedules via MCP.",

		Attributes: map[string]schema.Attribute{
			"ids": schema.ListAttribute{
				MarkdownDescription: "List of on-call schedule UUIDs.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"schedules": schema.ListNestedAttribute{
				MarkdownDescription: "List of on-call schedules.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier (UUID) of the schedule.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the schedule.",
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
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *OnCallSchedulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *OnCallSchedulesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OnCallSchedulesDataSourceModel

	schedules, err := d.client.ListOnCallSchedules(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching on-call schedules", err.Error())
		return
	}

	state.Schedules = make([]OnCallScheduleDataModel, 0, len(schedules))
	ids := make([]attr.Value, 0, len(schedules))

	for _, s := range schedules {
		state.Schedules = append(state.Schedules, OnCallScheduleDataModel{
			ID:            types.StringValue(s.UUID),
			Name:          types.StringValue(s.Name),
			Team:          types.StringValue(s.Team),
			CurrentOnCall: types.StringValue(s.CurrentOncall),
			NextOnCall:    types.StringValue(s.NextOncall),
			RotationStart: types.StringValue(s.RotationStart),
			RotationEnd:   types.StringValue(s.RotationEnd),
		})
		ids = append(ids, types.StringValue(s.UUID))
	}

	idList, diag := types.ListValue(types.StringType, ids)
	resp.Diagnostics.Append(diag...)
	state.IDs = idList

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
