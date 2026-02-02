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
	_ datasource.DataSource              = &MaintenanceWindowDataSource{}
	_ datasource.DataSourceWithConfigure = &MaintenanceWindowDataSource{}
)

// NewMaintenanceWindowDataSource creates a new single maintenance window data source.
func NewMaintenanceWindowDataSource() datasource.DataSource {
	return &MaintenanceWindowDataSource{}
}

// MaintenanceWindowDataSource defines the data source implementation for a single maintenance window.
type MaintenanceWindowDataSource struct {
	client client.MaintenanceAPI
}

// MaintenanceWindowDataSourceModel describes the data source data model.
type MaintenanceWindowDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Title       types.String `tfsdk:"title"`
	Text        types.String `tfsdk:"text"`
	StartDate   types.String `tfsdk:"start_date"`
	EndDate     types.String `tfsdk:"end_date"`
	Timezone    types.String `tfsdk:"timezone"`
	Monitors    types.List   `tfsdk:"monitors"`
	StatusPages types.List   `tfsdk:"status_pages"`
	CreatedBy   types.String `tfsdk:"created_by"`
}

// Metadata returns the data source type name.
func (d *MaintenanceWindowDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance_window"
}

// Schema defines the schema for the data source.
func (d *MaintenanceWindowDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a specific Hyperping maintenance window by its ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the maintenance window to look up.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The internal name of the maintenance window.",
				Computed:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The public title of the maintenance window (English).",
				Computed:            true,
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "The description text of the maintenance window (English).",
				Computed:            true,
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The start date in ISO 8601 format.",
				Computed:            true,
			},
			"end_date": schema.StringAttribute{
				MarkdownDescription: "The end date in ISO 8601 format.",
				Computed:            true,
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "The timezone of the maintenance window.",
				Computed:            true,
			},
			"monitors": schema.ListAttribute{
				MarkdownDescription: "List of monitor UUIDs affected by this maintenance.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"status_pages": schema.ListAttribute{
				MarkdownDescription: "List of status page UUIDs this maintenance is displayed on.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"created_by": schema.StringAttribute{
				MarkdownDescription: "The email of the user who created this maintenance window.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *MaintenanceWindowDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *MaintenanceWindowDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config MaintenanceWindowDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := client.ValidateResourceID(config.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid Maintenance Window ID", fmt.Sprintf("Cannot look up maintenance window: %s", err))
		return
	}

	maint, err := d.client.GetMaintenance(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading maintenance window",
			fmt.Sprintf("Could not read maintenance window %s: %s", config.ID.ValueString(), err),
		)
		return
	}

	d.mapMaintenanceToDataSourceModel(maint, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// mapMaintenanceToDataSourceModel maps a client.Maintenance to the data source model.
func (d *MaintenanceWindowDataSource) mapMaintenanceToDataSourceModel(maint *client.Maintenance, model *MaintenanceWindowDataSourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(maint.UUID)
	model.Name = types.StringValue(maint.Name)
	model.Title = types.StringValue(maint.Title.En)
	model.Text = types.StringValue(maint.Text.En)
	model.Timezone = types.StringValue(maint.Timezone)
	model.CreatedBy = types.StringValue(maint.CreatedBy)

	if maint.StartDate != nil {
		model.StartDate = types.StringValue(*maint.StartDate)
	} else {
		model.StartDate = types.StringNull()
	}

	if maint.EndDate != nil {
		model.EndDate = types.StringValue(*maint.EndDate)
	} else {
		model.EndDate = types.StringNull()
	}

	model.Monitors = mapStringSliceToList(maint.Monitors, diags)
	model.StatusPages = mapStringSliceToList(maint.StatusPages, diags)
}
