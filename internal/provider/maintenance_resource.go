// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &MaintenanceResource{}
	_ resource.ResourceWithImportState = &MaintenanceResource{}
)

// NewMaintenanceResource creates a new maintenance resource.
func NewMaintenanceResource() resource.Resource {
	return &MaintenanceResource{}
}

// MaintenanceResource defines the resource implementation.
type MaintenanceResource struct {
	client client.MaintenanceAPI
}

// MaintenanceResourceModel describes the resource data model.
type MaintenanceResourceModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Title               types.String `tfsdk:"title"`
	Text                types.String `tfsdk:"text"`
	StartDate           types.String `tfsdk:"start_date"`
	EndDate             types.String `tfsdk:"end_date"`
	Monitors            types.List   `tfsdk:"monitors"`
	StatusPages         types.List   `tfsdk:"status_pages"`
	NotificationOption  types.String `tfsdk:"notification_option"`
	NotificationMinutes types.Int64  `tfsdk:"notification_minutes"`
}

// Metadata returns the resource type name.
func (r *MaintenanceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_maintenance"
}

// Schema defines the schema for the resource.
func (r *MaintenanceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hyperping maintenance window for scheduled downtime.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the maintenance window.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The internal name of the maintenance window.",
				Required:            true,
				Validators: []validator.String{
					StringLength(1, 255),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The public title of the maintenance window (English).",
				Optional:            true,
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "The description text of the maintenance (English).",
				Optional:            true,
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The scheduled start time in ISO 8601 format (e.g., `2026-01-20T02:00:00.000Z`).",
				Required:            true,
			},
			"end_date": schema.StringAttribute{
				MarkdownDescription: "The scheduled end time in ISO 8601 format (e.g., `2026-01-20T04:00:00.000Z`).",
				Required:            true,
			},
			"monitors": schema.ListAttribute{
				MarkdownDescription: "List of monitor UUIDs affected by this maintenance window.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"status_pages": schema.ListAttribute{
				MarkdownDescription: "List of status page UUIDs to display this maintenance on.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"notification_option": schema.StringAttribute{
				MarkdownDescription: "When to notify subscribers. Valid values: `scheduled`, `immediate`. Defaults to `scheduled`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("scheduled"),
			},
			"notification_minutes": schema.Int64Attribute{
				MarkdownDescription: "Number of minutes before the maintenance to notify subscribers. Defaults to `60`. Only used when notification_option is `scheduled`. Must be at least 1.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(client.DefaultNotifyBeforeMinutes),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *MaintenanceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = c
}

// Create creates the resource and sets the initial Terraform state.
func (r *MaintenanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MaintenanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate date format, order, and add warnings
	resp.Diagnostics.Append(validateMaintenanceDates(&plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	startDate := plan.StartDate.ValueString()
	endDate := plan.EndDate.ValueString()

	// Build create request
	createReq := client.CreateMaintenanceRequest{
		Name:      plan.Name.ValueString(),
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Handle optional title
	if !plan.Title.IsNull() {
		createReq.Title = client.LocalizedText{En: plan.Title.ValueString()}
	}

	// Handle optional text
	if !plan.Text.IsNull() {
		createReq.Text = client.LocalizedText{En: plan.Text.ValueString()}
	}

	// Handle monitors (required)
	var monitors []string
	resp.Diagnostics.Append(plan.Monitors.ElementsAs(ctx, &monitors, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createReq.Monitors = monitors

	// Handle optional status_pages
	if !isNullOrUnknown(plan.StatusPages) {
		var statusPages []string
		resp.Diagnostics.Append(plan.StatusPages.ElementsAs(ctx, &statusPages, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.StatusPages = statusPages
	}

	// Handle notification options
	if !plan.NotificationOption.IsNull() {
		createReq.NotificationOption = plan.NotificationOption.ValueString()
	}
	if !plan.NotificationMinutes.IsNull() {
		notifyBefore := int(plan.NotificationMinutes.ValueInt64())
		createReq.NotificationMinutes = &notifyBefore
	}

	// Call API to create maintenance window
	createResp, err := r.client.CreateMaintenance(ctx, createReq)
	if err != nil {
		resp.Diagnostics.Append(newCreateError("Maintenance Window", err))
		return
	}

	// Read full maintenance details (create response only contains UUID)
	maintenance, err := r.client.GetMaintenance(ctx, createResp.UUID)
	if err != nil {
		resp.Diagnostics.Append(newReadAfterCreateError("Maintenance Window", createResp.UUID, err))
		return
	}

	// Map complete API response to Terraform state
	r.mapMaintenanceToModel(maintenance, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *MaintenanceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MaintenanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	maintenance, err := r.client.GetMaintenance(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(newReadError("Maintenance Window", state.ID.ValueString(), err))
		return
	}

	r.mapMaintenanceToModel(maintenance, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *MaintenanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MaintenanceResourceModel
	var state MaintenanceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate dates if they changed
	if !plan.StartDate.Equal(state.StartDate) || !plan.EndDate.Equal(state.EndDate) {
		resp.Diagnostics.Append(validateMaintenanceDates(&plan)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Build update request
	updateReq := client.UpdateMaintenanceRequest{}

	// Only include changed fields
	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Title.Equal(state.Title) && !plan.Title.IsNull() {
		title := client.LocalizedText{En: plan.Title.ValueString()}
		updateReq.Title = &title
	}

	if !plan.Text.Equal(state.Text) && !plan.Text.IsNull() {
		text := client.LocalizedText{En: plan.Text.ValueString()}
		updateReq.Text = &text
	}

	if !plan.StartDate.Equal(state.StartDate) {
		startDate := plan.StartDate.ValueString()
		updateReq.StartDate = &startDate
	}

	if !plan.EndDate.Equal(state.EndDate) {
		endDate := plan.EndDate.ValueString()
		updateReq.EndDate = &endDate
	}

	// Handle monitors
	if !plan.Monitors.Equal(state.Monitors) {
		var monitors []string
		resp.Diagnostics.Append(plan.Monitors.ElementsAs(ctx, &monitors, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.Monitors = &monitors
	}

	// Call API to update maintenance window
	_, err := r.client.UpdateMaintenance(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.Append(newUpdateError("Maintenance Window", state.ID.ValueString(), err))
		return
	}

	// Read full maintenance details (update response doesn't contain complete data)
	maintenance, err := r.client.GetMaintenance(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(newReadAfterUpdateError("Maintenance Window", state.ID.ValueString(), err))
		return
	}

	// Map complete API response to Terraform state
	r.mapMaintenanceToModel(maintenance, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource.
func (r *MaintenanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MaintenanceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMaintenance(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			// Already deleted, no error
			return
		}
		resp.Diagnostics.Append(newDeleteError("Maintenance Window", state.ID.ValueString(), err))
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *MaintenanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Validate the import ID before setting state (VULN-015)
	if err := client.ValidateResourceID(req.ID); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import maintenance window: %s", err))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// validateMaintenanceDates validates the maintenance date range and adds diagnostics.
// It checks:
// - Both dates can be parsed as ISO 8601 (RFC3339)
// - end_date is after start_date (error)
// - start_date is in the future (warning)
// - duration is reasonable (< 7 days warning)
func validateMaintenanceDates(plan *MaintenanceResourceModel) diag.Diagnostics {
	var diags diag.Diagnostics

	if plan.StartDate.IsNull() || plan.EndDate.IsNull() {
		return diags
	}

	startDateStr := plan.StartDate.ValueString()
	endDateStr := plan.EndDate.ValueString()

	// Parse start date
	startTime, err := time.Parse(time.RFC3339, startDateStr)
	if err != nil {
		diags.AddAttributeError(
			path.Root("start_date"),
			"Invalid Start Date",
			fmt.Sprintf("Could not parse start_date as ISO 8601 time: %s", err),
		)
		return diags
	}

	// Parse end date
	endTime, err := time.Parse(time.RFC3339, endDateStr)
	if err != nil {
		diags.AddAttributeError(
			path.Root("end_date"),
			"Invalid End Date",
			fmt.Sprintf("Could not parse end_date as ISO 8601 time: %s", err),
		)
		return diags
	}

	// Validate end is after start (error)
	if !endTime.After(startTime) {
		diags.AddAttributeError(
			path.Root("end_date"),
			"Invalid Date Range",
			"end_date must be after start_date",
		)
		return diags
	}

	// Check if start date is in the past (warning)
	now := time.Now()
	if startTime.Before(now) {
		diags.AddAttributeWarning(
			path.Root("start_date"),
			"Past Start Date",
			fmt.Sprintf("start_date is in the past (%s). This maintenance window may not trigger as expected. "+
				"Consider using a future date for scheduled maintenance.", startTime.Format(time.RFC3339)),
		)
	}

	// Check duration (warning if > 7 days)
	duration := endTime.Sub(startTime)
	if duration > 7*24*time.Hour {
		diags.AddAttributeWarning(
			path.Root("end_date"),
			"Long Maintenance Window",
			fmt.Sprintf("Maintenance window duration is %v (%.1f days). "+
				"Consider breaking long maintenance into multiple shorter windows for better visibility.",
				duration.Round(time.Hour), duration.Hours()/24),
		)
	}

	return diags
}

// mapMaintenanceToModel maps a client.Maintenance to the Terraform model.
func (r *MaintenanceResource) mapMaintenanceToModel(maintenance *client.Maintenance, model *MaintenanceResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(maintenance.UUID)
	model.Name = types.StringValue(maintenance.Name)

	// Handle title
	if maintenance.Title.En != "" {
		model.Title = types.StringValue(maintenance.Title.En)
	} else {
		model.Title = types.StringNull()
	}

	// NOTE: Text field behavior - Hyperping API quirk
	// The API accepts text during CREATE/UPDATE but may not return it in GET responses
	// If API returns it (non-empty), use that value; otherwise preserve plan value
	if maintenance.Text.En != "" {
		model.Text = types.StringValue(maintenance.Text.En)
	}
	// If empty and model.Text is already set (from plan), keep the existing value
	// This prevents state drift when API doesn't return the field

	// Handle dates
	if maintenance.StartDate != nil {
		model.StartDate = types.StringValue(*maintenance.StartDate)
	} else {
		model.StartDate = types.StringNull()
	}

	if maintenance.EndDate != nil {
		model.EndDate = types.StringValue(*maintenance.EndDate)
	} else {
		model.EndDate = types.StringNull()
	}

	// Handle monitors using shared helper
	model.Monitors = mapStringSliceToList(maintenance.Monitors, diags)

	// Handle status_pages using shared helper
	model.StatusPages = mapStringSliceToList(maintenance.StatusPages, diags)

	// Handle notification options
	if maintenance.NotificationOption != "" {
		model.NotificationOption = types.StringValue(maintenance.NotificationOption)
	}
	if maintenance.NotificationMinutes != nil {
		model.NotificationMinutes = types.Int64Value(int64(*maintenance.NotificationMinutes))
	}
}
