// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &HealthcheckResource{}
	_ resource.ResourceWithImportState = &HealthcheckResource{}
)

// NewHealthcheckResource creates a new healthcheck resource.
func NewHealthcheckResource() resource.Resource {
	return &HealthcheckResource{}
}

// HealthcheckResource defines the resource implementation.
type HealthcheckResource struct {
	client client.HealthcheckAPI
}

// HealthcheckResourceModel describes the resource data model.
type HealthcheckResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	PingURL          types.String `tfsdk:"ping_url"`
	Cron             types.String `tfsdk:"cron"`
	Timezone         types.String `tfsdk:"timezone"`
	PeriodValue      types.Int64  `tfsdk:"period_value"`
	PeriodType       types.String `tfsdk:"period_type"`
	GracePeriodValue types.Int64  `tfsdk:"grace_period_value"`
	GracePeriodType  types.String `tfsdk:"grace_period_type"`
	EscalationPolicy types.String `tfsdk:"escalation_policy"`
	IsPaused         types.Bool   `tfsdk:"is_paused"`
	IsDown           types.Bool   `tfsdk:"is_down"`
	Period           types.Int64  `tfsdk:"period"`
	GracePeriod      types.Int64  `tfsdk:"grace_period"`
	LastPing         types.String `tfsdk:"last_ping"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

// Metadata returns the resource type name.
func (r *HealthcheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_healthcheck"
}

// Schema defines the schema for the resource.
func (r *HealthcheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hyperping healthcheck for cron job monitoring (dead man's switch).",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the healthcheck.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the healthcheck.",
				Required:            true,
				Validators: []validator.String{
					StringLength(1, 255),
				},
			},
			"ping_url": schema.StringAttribute{
				MarkdownDescription: "The auto-generated ping URL. Your cron job pings this URL to prove it ran.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cron": schema.StringAttribute{
				MarkdownDescription: "Cron expression defining the schedule (e.g., `0 0 * * *`). Mutually exclusive with `period_value`/`period_type`.",
				Optional:            true,
				Validators: []validator.String{
					CronExpression(),
				},
			},
			"timezone": schema.StringAttribute{
				MarkdownDescription: "Timezone for the cron expression (e.g., `America/New_York`). Required when `cron` is set.",
				Optional:            true,
				Validators: []validator.String{
					Timezone(),
				},
			},
			"period_value": schema.Int64Attribute{
				MarkdownDescription: "Numeric value for the expected interval. Mutually exclusive with `cron`/`tz`.",
				Optional:            true,
			},
			"period_type": schema.StringAttribute{
				MarkdownDescription: "Unit for `period_value`. Valid values: `seconds`, `minutes`, `hours`, `days`.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(client.AllowedPeriodTypes...),
				},
			},
			"grace_period_value": schema.Int64Attribute{
				MarkdownDescription: "Numeric value for the grace period buffer before alerting.",
				Required:            true,
			},
			"grace_period_type": schema.StringAttribute{
				MarkdownDescription: "Unit for `grace_period_value`. Valid values: `seconds`, `minutes`, `hours`, `days`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(client.AllowedPeriodTypes...),
				},
			},
			"escalation_policy": schema.StringAttribute{
				MarkdownDescription: "UUID of the escalation policy to link to this healthcheck.",
				Optional:            true,
				Validators: []validator.String{
					UUIDFormat(),
				},
			},
			"is_paused": schema.BoolAttribute{
				MarkdownDescription: "Whether the healthcheck is paused. Defaults to `false`.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"is_down": schema.BoolAttribute{
				MarkdownDescription: "Whether the healthcheck is currently in a failure state (read-only).",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"period": schema.Int64Attribute{
				MarkdownDescription: "Calculated period in seconds (read-only).",
				Computed:            true,
			},
			"grace_period": schema.Int64Attribute{
				MarkdownDescription: "Calculated grace period in seconds (read-only).",
				Computed:            true,
			},
			"last_ping": schema.StringAttribute{
				MarkdownDescription: "Timestamp of the last ping received in ISO 8601 format (read-only).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp in ISO 8601 format (read-only).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *HealthcheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// validateCronPeriodExclusivity checks that cron/tz and period_value/period_type are mutually exclusive.
func validateCronPeriodExclusivity(plan *HealthcheckResourceModel) error {
	hasCron := !isNullOrUnknown(plan.Cron) && plan.Cron.ValueString() != ""
	hasTz := !isNullOrUnknown(plan.Timezone) && plan.Timezone.ValueString() != ""
	hasPeriodValue := !isNullOrUnknown(plan.PeriodValue)
	hasPeriodType := !isNullOrUnknown(plan.PeriodType)

	if (hasCron || hasTz) && (hasPeriodValue || hasPeriodType) {
		return fmt.Errorf("specify either (cron + tz) or (period_value + period_type), not both")
	}

	if hasCron && !hasTz {
		return fmt.Errorf("tz is required when cron is set")
	}

	if hasTz && !hasCron {
		return fmt.Errorf("cron is required when tz is set")
	}

	if hasPeriodValue && !hasPeriodType {
		return fmt.Errorf("period_type is required when period_value is set")
	}

	if hasPeriodType && !hasPeriodValue {
		return fmt.Errorf("period_value is required when period_type is set")
	}

	if !hasCron && !hasPeriodValue {
		return fmt.Errorf("either cron or period_value must be specified")
	}

	return nil
}

// Create creates the resource and sets the initial Terraform state.
func (r *HealthcheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan HealthcheckResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateCronPeriodExclusivity(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}

	// Capture user's desired pause state BEFORE mapHealthcheckToModel overwrites it
	// with the API's current value (which is always false on create).
	wantPaused := !plan.IsPaused.IsNull() && plan.IsPaused.ValueBool()

	createReq := buildCreateHealthcheckRequest(&plan)

	created, err := r.client.CreateHealthcheck(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating healthcheck",
			fmt.Sprintf("Could not create healthcheck: %s", err),
		)
		return
	}

	// Write the ID to state immediately to prevent orphaned resources if subsequent operations fail.
	// If GetHealthcheck fails, the user can retry and Terraform will attempt an update (not recreate).
	plan.ID = types.StringValue(created.UUID)

	// Read back full state
	healthcheck, err := r.client.GetHealthcheck(ctx, created.UUID)
	if err != nil {
		// Write partial state with just the ID so the resource isn't orphaned.
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		resp.Diagnostics.AddError(
			"Error reading created healthcheck",
			fmt.Sprintf("Healthcheck created (ID: %s) but failed to read full state: %s. "+
				"The resource has been saved to state with its ID. "+
				"Run 'terraform apply' again to retry reading the full state.", created.UUID, err),
		)
		return
	}

	r.mapHealthcheckToModel(healthcheck, &plan)

	// If user requested paused, call the pause API. On failure, still write state
	// (the resource exists remotely) but surface a warning so the user knows
	// the pause did not take effect.
	if wantPaused {
		_, err := r.client.PauseHealthcheck(ctx, healthcheck.UUID)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Healthcheck created but not paused",
				fmt.Sprintf("Healthcheck %s was created successfully but the pause request failed: %s. "+
					"The resource is active. Set is_paused = true again on next apply to retry.", healthcheck.UUID, err),
			)
		} else {
			plan.IsPaused = types.BoolValue(true)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// buildCreateHealthcheckRequest converts the plan model into an API request.
func buildCreateHealthcheckRequest(plan *HealthcheckResourceModel) client.CreateHealthcheckRequest {
	req := client.CreateHealthcheckRequest{
		Name:             plan.Name.ValueString(),
		GracePeriodValue: int(plan.GracePeriodValue.ValueInt64()),
		GracePeriodType:  plan.GracePeriodType.ValueString(),
	}

	if !plan.Cron.IsNull() {
		cron := plan.Cron.ValueString()
		req.Cron = &cron
	}
	if !plan.Timezone.IsNull() {
		tz := plan.Timezone.ValueString()
		req.Timezone = &tz
	}
	if !plan.PeriodValue.IsNull() {
		pv := int(plan.PeriodValue.ValueInt64())
		req.PeriodValue = &pv
	}
	if !plan.PeriodType.IsNull() {
		pt := plan.PeriodType.ValueString()
		req.PeriodType = &pt
	}
	if !plan.EscalationPolicy.IsNull() {
		ep := plan.EscalationPolicy.ValueString()
		req.EscalationPolicy = &ep
	}

	return req
}

// Read refreshes the Terraform state with the latest data.
func (r *HealthcheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state HealthcheckResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	healthcheck, err := r.client.GetHealthcheck(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading healthcheck",
			fmt.Sprintf("Could not read healthcheck %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	r.mapHealthcheckToModel(healthcheck, &state)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *HealthcheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan HealthcheckResourceModel
	var state HealthcheckResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := validateCronPeriodExclusivity(&plan); err != nil {
		resp.Diagnostics.AddError("Invalid Configuration", err.Error())
		return
	}

	// Capture desired pause state before mapping overwrites it.
	// Guard against Unknown values — ValueBool() on Unknown returns false,
	// which would silently trigger an unintended resume.
	pauseKnown := !isNullOrUnknown(plan.IsPaused)
	wantPaused := pauseKnown && plan.IsPaused.ValueBool()
	pauseChanged := pauseKnown && !plan.IsPaused.Equal(state.IsPaused)

	r.applyFieldChanges(ctx, &plan, &state, resp)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle pause/resume — surface warnings on failure so state is still written
	if pauseChanged {
		r.applyPauseState(ctx, state.ID.ValueString(), wantPaused, resp)
	}

	// Read back final state
	healthcheck, err := r.client.GetHealthcheck(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading updated healthcheck",
			fmt.Sprintf("Could not read healthcheck %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	r.mapHealthcheckToModel(healthcheck, &plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// applyFieldChanges builds and sends an update request for changed non-pause fields.
func (r *HealthcheckResource) applyFieldChanges(ctx context.Context, plan, state *HealthcheckResourceModel, resp *resource.UpdateResponse) {
	updateReq := client.UpdateHealthcheckRequest{}
	hasChanges := false

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
		hasChanges = true
	}
	if !plan.Cron.Equal(state.Cron) {
		if plan.Cron.IsNull() {
			empty := ""
			updateReq.Cron = &empty
		} else {
			cron := plan.Cron.ValueString()
			updateReq.Cron = &cron
		}
		hasChanges = true
	}
	if !plan.Timezone.Equal(state.Timezone) {
		if plan.Timezone.IsNull() {
			empty := ""
			updateReq.Timezone = &empty
		} else {
			tz := plan.Timezone.ValueString()
			updateReq.Timezone = &tz
		}
		hasChanges = true
	}
	if !plan.PeriodValue.Equal(state.PeriodValue) {
		if !plan.PeriodValue.IsNull() {
			pv := int(plan.PeriodValue.ValueInt64())
			updateReq.PeriodValue = &pv
			hasChanges = true
		}
		// When clearing period_value (switching to cron mode), omit the field entirely.
		// The API will interpret PeriodType="" as the signal to clear period-based scheduling.
		// Sending 0 is ambiguous and may be rejected as invalid.
	}
	if !plan.PeriodType.Equal(state.PeriodType) {
		if plan.PeriodType.IsNull() {
			empty := ""
			updateReq.PeriodType = &empty
		} else {
			pt := plan.PeriodType.ValueString()
			updateReq.PeriodType = &pt
		}
		hasChanges = true
	}
	if !plan.GracePeriodValue.Equal(state.GracePeriodValue) {
		gpv := int(plan.GracePeriodValue.ValueInt64())
		updateReq.GracePeriodValue = &gpv
		hasChanges = true
	}
	if !plan.GracePeriodType.Equal(state.GracePeriodType) {
		gpt := plan.GracePeriodType.ValueString()
		updateReq.GracePeriodType = &gpt
		hasChanges = true
	}
	if !plan.EscalationPolicy.Equal(state.EscalationPolicy) {
		if plan.EscalationPolicy.IsNull() {
			empty := ""
			updateReq.EscalationPolicy = &empty
		} else {
			ep := plan.EscalationPolicy.ValueString()
			updateReq.EscalationPolicy = &ep
		}
		hasChanges = true
	}

	if !hasChanges {
		return
	}

	_, err := r.client.UpdateHealthcheck(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating healthcheck",
			fmt.Sprintf("Could not update healthcheck %s: %s", state.ID.ValueString(), err),
		)
	}
}

// applyPauseState calls pause or resume and surfaces a warning (not error) on failure.
func (r *HealthcheckResource) applyPauseState(ctx context.Context, id string, wantPaused bool, resp *resource.UpdateResponse) {
	if wantPaused {
		_, err := r.client.PauseHealthcheck(ctx, id)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Failed to pause healthcheck",
				fmt.Sprintf("Healthcheck %s updates applied but pause failed: %s. Retry on next apply.", id, err),
			)
		}
	} else {
		_, err := r.client.ResumeHealthcheck(ctx, id)
		if err != nil {
			resp.Diagnostics.AddWarning(
				"Failed to resume healthcheck",
				fmt.Sprintf("Healthcheck %s updates applied but resume failed: %s. Retry on next apply.", id, err),
			)
		}
	}
}

// Delete deletes the resource.
func (r *HealthcheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state HealthcheckResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteHealthcheck(ctx, state.ID.ValueString())
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError(
				"Error deleting healthcheck",
				fmt.Sprintf("Could not delete healthcheck %s: %s", state.ID.ValueString(), err),
			)
			return
		}
	}

	resp.Diagnostics.AddWarning(
		"Healthcheck deleted — ping URL is now inactive",
		fmt.Sprintf("The healthcheck %s and its ping URL (%s) have been destroyed. "+
			"Any cron jobs still pinging this URL will receive errors. "+
			"Update or disable those jobs to avoid false alerts.",
			state.ID.ValueString(), state.PingURL.ValueString()),
	)
}

// ImportState imports an existing resource into Terraform.
func (r *HealthcheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if err := client.ValidateResourceID(req.ID); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import healthcheck: %s", err))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapHealthcheckToModel maps a client.Healthcheck to the Terraform resource model
// using the shared HealthcheckCommonFields mapping.
func (r *HealthcheckResource) mapHealthcheckToModel(hc *client.Healthcheck, model *HealthcheckResourceModel) {
	f := MapHealthcheckCommonFields(hc)
	model.ID = f.ID
	model.Name = f.Name
	model.PingURL = f.PingURL
	model.Cron = f.Cron
	model.Timezone = f.Timezone
	model.PeriodValue = f.PeriodValue
	model.PeriodType = f.PeriodType
	model.GracePeriodValue = f.GracePeriodValue
	model.GracePeriodType = f.GracePeriodType
	model.EscalationPolicy = f.EscalationPolicy
	model.IsPaused = f.IsPaused
	model.IsDown = f.IsDown
	model.Period = f.Period
	model.GracePeriod = f.GracePeriod
	model.LastPing = f.LastPing
	model.CreatedAt = f.CreatedAt
}
