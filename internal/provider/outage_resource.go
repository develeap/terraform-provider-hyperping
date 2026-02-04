// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// OutageTypeManual is the outage type value for resources created by this provider.
const OutageTypeManual = "manual"

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &OutageResource{}
	_ resource.ResourceWithImportState = &OutageResource{}
)

// NewOutageResource creates a new outage resource.
func NewOutageResource() resource.Resource {
	return &OutageResource{}
}

// OutageResource defines the resource implementation.
type OutageResource struct {
	client client.OutageAPI
}

// OutageResourceModel describes the resource data model.
type OutageResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	MonitorUUID          types.String `tfsdk:"monitor_uuid"`
	StartDate            types.String `tfsdk:"start_date"`
	EndDate              types.String `tfsdk:"end_date"`
	StatusCode           types.Int64  `tfsdk:"status_code"`
	Description          types.String `tfsdk:"description"`
	EscalationPolicyUUID types.String `tfsdk:"escalation_policy_uuid"`
	OutageType           types.String `tfsdk:"outage_type"`
	IsResolved           types.Bool   `tfsdk:"is_resolved"`
	DurationMs           types.Int64  `tfsdk:"duration_ms"`
	DetectedLocation     types.String `tfsdk:"detected_location"`
	Monitor              types.Object `tfsdk:"monitor"`
	AcknowledgedBy       types.Object `tfsdk:"acknowledged_by"`
}

// Metadata returns the resource type name.
func (r *OutageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_outage"
}

// Schema defines the schema for the resource.
func (r *OutageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a manual Hyperping outage. All user-settable fields are ForceNew since outages cannot be updated via the API.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the outage.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"monitor_uuid": schema.StringAttribute{
				MarkdownDescription: "The UUID of the monitor this outage is associated with.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					UUIDFormat(),
				},
			},
			"start_date": schema.StringAttribute{
				MarkdownDescription: "The start date of the outage in ISO 8601 format.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					ISO8601(),
				},
			},
			"end_date": schema.StringAttribute{
				MarkdownDescription: "The end date of the outage in ISO 8601 format. Null if ongoing.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					ISO8601(),
				},
			},
			"status_code": schema.Int64Attribute{
				MarkdownDescription: "The HTTP status code that caused the outage. Must be a valid HTTP status (100-599).",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.Between(100, 599),
				},
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of the outage (error message or reason).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"escalation_policy_uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the escalation policy to link to this outage. If provided, the policy will be triggered according to its step timing.",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					UUIDFormat(),
				},
			},
			"outage_type": schema.StringAttribute{
				MarkdownDescription: "The type of outage. Always `manual` for created outages (read-only).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_resolved": schema.BoolAttribute{
				MarkdownDescription: "Whether the outage is resolved (read-only).",
				Computed:            true,
			},
			"duration_ms": schema.Int64Attribute{
				MarkdownDescription: "Duration of the outage in milliseconds (read-only).",
				Computed:            true,
			},
			"detected_location": schema.StringAttribute{
				MarkdownDescription: "The location that detected the outage (read-only).",
				Computed:            true,
			},
			"monitor": schema.SingleNestedAttribute{
				MarkdownDescription: "The monitor associated with this outage (read-only).",
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
				MarkdownDescription: "The user who acknowledged this outage (read-only, null if not acknowledged).",
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
	}
}

// Configure adds the provider configured client to the resource.
func (r *OutageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *OutageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OutageResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateOutageRequest{
		MonitorUUID: plan.MonitorUUID.ValueString(),
		StartDate:   plan.StartDate.ValueString(),
		StatusCode:  int(plan.StatusCode.ValueInt64()),
		Description: plan.Description.ValueString(),
		OutageType:  OutageTypeManual,
	}

	if !plan.EndDate.IsNull() {
		endDate := plan.EndDate.ValueString()
		createReq.EndDate = &endDate
	}

	if !plan.EscalationPolicyUUID.IsNull() {
		escPolicyUUID := plan.EscalationPolicyUUID.ValueString()
		createReq.EscalationPolicyUUID = &escPolicyUUID
	}

	created, err := r.client.CreateOutage(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating outage",
			fmt.Sprintf("Could not create outage: %s", err),
		)
		return
	}

	// Read back full state
	outage, err := r.client.GetOutage(ctx, created.UUID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading created outage",
			fmt.Sprintf("Outage created (ID: %s) but failed to read: %s", created.UUID, err),
		)
		return
	}

	r.mapOutageToModel(outage, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *OutageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OutageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	outage, err := r.client.GetOutage(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading outage",
			fmt.Sprintf("Could not read outage %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	r.mapOutageToModel(outage, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update is a no-op since all fields are ForceNew.
// If this method is ever called, it indicates a bug where a field was made non-ForceNew.
func (r *OutageResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All user-settable fields are ForceNew, so Terraform will destroy and recreate.
	// This should never be called in practice. If it is, it indicates a schema misconfiguration.
	resp.Diagnostics.AddError(
		"Unexpected Update Call",
		"BUG: OutageResource.Update was called, but all attributes are ForceNew. "+
			"This indicates a schema misconfiguration. Please report this issue to the provider developers.",
	)
}

// Delete removes the outage from Terraform state without calling the API.
// Outages are forensic records — destroying them would erase incident history.
// Terraform state removal is sufficient; the outage remains in Hyperping for audit.
func (r *OutageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OutageResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"Outage removed from state only",
		fmt.Sprintf("Outage %s has been removed from Terraform state but NOT deleted from Hyperping. "+
			"The outage record is preserved for audit and incident history purposes. "+
			"To permanently delete the outage, use the Hyperping dashboard or API directly: "+
			"DELETE /v1/outages/%s", state.ID.ValueString(), state.ID.ValueString()),
	)
}

// ImportState imports an existing resource into Terraform.
func (r *OutageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	if err := client.ValidateResourceID(req.ID); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import outage: %s", err))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapOutageToModel maps a client.Outage to the Terraform resource model
// using the shared MapOutageNestedObjects helper for nested monitor/acknowledged_by.
func (r *OutageResource) mapOutageToModel(outage *client.Outage, model *OutageResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(outage.UUID)
	model.MonitorUUID = types.StringValue(outage.Monitor.UUID)
	model.StartDate = types.StringValue(outage.StartDate)
	model.StatusCode = types.Int64Value(int64(outage.StatusCode))
	model.Description = types.StringValue(outage.Description)
	model.IsResolved = types.BoolValue(outage.IsResolved)
	model.DurationMs = types.Int64Value(int64(outage.DurationMs))
	model.DetectedLocation = types.StringValue(outage.DetectedLocation)
	// Map outage_type directly from API response. Resources created by this provider
	// send OutageTypeManual in the request, so the API should echo it back.
	// If the API returns empty string (shouldn't happen), map it truthfully to maintain
	// consistency with data sources.
	model.OutageType = types.StringValue(outage.OutageType)

	if outage.EndDate != nil {
		model.EndDate = types.StringValue(*outage.EndDate)
	} else {
		model.EndDate = types.StringNull()
	}

	if outage.EscalationPolicy != nil {
		model.EscalationPolicyUUID = types.StringValue(outage.EscalationPolicy.UUID)
	} else {
		model.EscalationPolicyUUID = types.StringNull()
	}

	model.Monitor, model.AcknowledgedBy = MapOutageNestedObjects(outage, diags)
}
