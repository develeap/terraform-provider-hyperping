// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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

	hyperping "github.com/develeap/hyperping-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                   = &MaintenanceResource{}
	_ resource.ResourceWithImportState    = &MaintenanceResource{}
	_ resource.ResourceWithValidateConfig = &MaintenanceResource{}
)

// NewMaintenanceResource creates a new maintenance resource.
func NewMaintenanceResource() resource.Resource {
	return &MaintenanceResource{}
}

// MaintenanceResource defines the resource implementation.
type MaintenanceResource struct {
	client hyperping.MaintenanceAPI
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
				Validators: []validator.String{
					StringLength(1, 255),
				},
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "The description text of the maintenance (English).",
				Optional:            true,
				Validators: []validator.String{
					StringLength(1, 10000),
				},
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The scheduled start time in ISO 8601 format (e.g., `2026-01-20T02:00:00.000Z`).",
				Required:            true,
				Validators: []validator.String{
					ISO8601(),
				},
			},
			"end_date": schema.StringAttribute{
				MarkdownDescription: "The scheduled end time in ISO 8601 format (e.g., `2026-01-20T04:00:00.000Z`).",
				Required:            true,
				Validators: []validator.String{
					ISO8601(),
				},
			},
			"monitors": schema.ListAttribute{
				MarkdownDescription: "List of monitor UUIDs affected by this maintenance window.",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.List{
					listvalidator.ValueStringsAre(UUIDFormat()),
				},
			},
			"status_pages": schema.ListAttribute{
				MarkdownDescription: "List of status page UUIDs to display this maintenance on.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"notification_option": schema.StringAttribute{
				MarkdownDescription: "When to notify subscribers. Valid values: `none`, `scheduled`, `immediate`. Defaults to `none` (no notification).",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.OneOf(hyperping.AllowedNotificationOptions...),
				},
			},
			"notification_minutes": schema.Int64Attribute{
				MarkdownDescription: "Number of minutes before the maintenance to notify subscribers. Defaults to `60`. Only used when notification_option is `scheduled`. Must be at least 1.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(hyperping.DefaultNotifyBeforeMinutes),
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

	clients, ok := req.ProviderData.(*hyperpingClients)
	if !ok {
		resp.Diagnostics.Append(newUnexpectedConfigTypeError("*hyperpingClients", req.ProviderData))
		return
	}

	r.client = clients.REST
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
	createReq := hyperping.CreateMaintenanceRequest{
		Name:      plan.Name.ValueString(),
		StartDate: startDate,
		EndDate:   endDate,
	}

	// Handle optional title
	if !plan.Title.IsNull() {
		createReq.Title = hyperping.LocalizedText{En: plan.Title.ValueString()}
	}

	// Handle optional text
	if !plan.Text.IsNull() {
		createReq.Text = hyperping.LocalizedText{En: plan.Text.ValueString()}
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
		resp.Diagnostics.Append(NewCreateErrorWithContext("Maintenance Window", err))
		return
	}

	plan.ID = types.StringValue(createResp.UUID)

	// Map API response to Terraform state (POST now returns complete object)
	r.mapMaintenanceToModel(createResp, &plan, &resp.Diagnostics)
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
		if hyperping.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.Append(NewReadErrorWithContext("Maintenance Window", state.ID.ValueString(), err))
		return
	}

	r.mapMaintenanceToModel(maintenance, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// buildMaintenanceUpdateRequest constructs an UpdateMaintenanceRequest with only changed fields.
// Compares plan vs state and populates request with differences.
func buildMaintenanceUpdateRequest(ctx context.Context, plan *MaintenanceResourceModel, state *MaintenanceResourceModel, diags *diag.Diagnostics) hyperping.UpdateMaintenanceRequest {
	updateReq := hyperping.UpdateMaintenanceRequest{}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Title.Equal(state.Title) {
		title := hyperping.LocalizedText{En: plan.Title.ValueString()}
		updateReq.Title = &title
	}

	if !plan.Text.Equal(state.Text) {
		// When text is cleared (set to null), ValueString() returns "" and the API
		// receives {"en":""}. This is the correct behavior for this API: sending an
		// empty string clears the text field on the Hyperping side.
		text := hyperping.LocalizedText{En: plan.Text.ValueString()}
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

	if !plan.Monitors.Equal(state.Monitors) {
		var monitors []string
		diags.Append(plan.Monitors.ElementsAs(ctx, &monitors, false)...)
		if !diags.HasError() {
			updateReq.Monitors = &monitors
		}
	}

	if !plan.StatusPages.Equal(state.StatusPages) {
		var statusPages []string
		if !isNullOrUnknown(plan.StatusPages) {
			diags.Append(plan.StatusPages.ElementsAs(ctx, &statusPages, false)...)
		}
		if !diags.HasError() {
			updateReq.StatusPages = &statusPages
		}
	}

	if !plan.NotificationOption.Equal(state.NotificationOption) {
		notifOption := plan.NotificationOption.ValueString()
		updateReq.NotificationOption = &notifOption
	}

	if !plan.NotificationMinutes.Equal(state.NotificationMinutes) {
		notifMinutes := int(plan.NotificationMinutes.ValueInt64())
		updateReq.NotificationMinutes = &notifMinutes
	}

	return updateReq
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

	// Build update request with only changed fields
	updateReq := buildMaintenanceUpdateRequest(ctx, &plan, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to update maintenance window
	updateResp, err := r.client.UpdateMaintenance(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.Append(NewUpdateErrorWithContext("Maintenance Window", state.ID.ValueString(), err))
		return
	}

	// Map API response to Terraform state (PUT now returns complete object)
	r.mapMaintenanceToModel(updateResp, &plan, &resp.Diagnostics)
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
		if hyperping.IsNotFound(err) {
			// Already deleted, no error
			return
		}
		resp.Diagnostics.Append(NewDeleteErrorWithContext("Maintenance Window", state.ID.ValueString(), err))
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *MaintenanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Validate the import ID before setting state (VULN-015)
	if err := hyperping.ValidateResourceID(req.ID); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import maintenance window: %s", err))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// ValidateConfig implements resource.ResourceWithValidateConfig for cross-field
// validation at plan time, before any API call.
//
// Design: This is the first validation layer (plan-time). It only checks
// end_date > start_date. The second layer, validateMaintenanceDates, runs at
// apply-time and adds warnings (past start_date, long duration) that are
// inappropriate at plan-time where values may change before apply.
// Unparseable dates are silently skipped here; the ISO8601 schema validators
// catch format issues independently.
func (r *MaintenanceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var startDate types.String
	var endDate types.String
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("start_date"), &startDate)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("end_date"), &endDate)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Skip validation when dates are unknown (module composition support)
	if startDate.IsUnknown() || startDate.IsNull() || endDate.IsUnknown() || endDate.IsNull() {
		return
	}

	startTime, errStart := time.Parse(time.RFC3339, startDate.ValueString())
	endTime, errEnd := time.Parse(time.RFC3339, endDate.ValueString())
	if errStart != nil || errEnd != nil {
		return // ISO8601 validators will catch format issues
	}

	if !endTime.After(startTime) {
		resp.Diagnostics.AddAttributeError(
			path.Root("end_date"),
			"Invalid Date Range",
			"end_date must be after start_date",
		)
	}
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

// mapMaintenanceToModel maps a hyperping.Maintenance to the Terraform model.
func (r *MaintenanceResource) mapMaintenanceToModel(maintenance *hyperping.Maintenance, model *MaintenanceResourceModel, diags *diag.Diagnostics) {
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
