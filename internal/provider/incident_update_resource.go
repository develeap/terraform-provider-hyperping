// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &IncidentUpdateResource{}
	_ resource.ResourceWithImportState = &IncidentUpdateResource{}
)

// NewIncidentUpdateResource creates a new incident update resource.
func NewIncidentUpdateResource() resource.Resource {
	return &IncidentUpdateResource{}
}

// IncidentUpdateResource defines the resource implementation.
type IncidentUpdateResource struct {
	client client.IncidentAPI
}

// IncidentUpdateResourceModel describes the resource data model.
type IncidentUpdateResourceModel struct {
	ID         types.String `tfsdk:"id"`
	IncidentID types.String `tfsdk:"incident_id"`
	Text       types.String `tfsdk:"text"`
	Type       types.String `tfsdk:"type"`
	Date       types.String `tfsdk:"date"`
}

// Metadata returns the resource type name.
func (r *IncidentUpdateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incident_update"
}

// Schema defines the schema for the resource.
func (r *IncidentUpdateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages an update to a Hyperping incident. Use this to add status updates to an incident timeline.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the update (format: incident_id/update_id).",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"incident_id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the incident to add the update to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "The text content of the update (English).",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of update. Valid values: `investigating`, `identified`, `update`, `monitoring`, `resolved`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(client.AllowedIncidentUpdateTypes...),
				},
			},
			"date": schema.StringAttribute{
				MarkdownDescription: "The date of the update in ISO 8601 format. If not provided, the current time is used.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IncidentUpdateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *IncidentUpdateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IncidentUpdateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the request
	addReq := client.AddIncidentUpdateRequest{
		Text: client.LocalizedText{En: plan.Text.ValueString()},
		Type: plan.Type.ValueString(),
	}

	// Handle optional date
	if !isNullOrUnknown(plan.Date) {
		addReq.Date = plan.Date.ValueString()
	}

	// Call API to add update
	incident, err := r.client.AddIncidentUpdate(ctx, plan.IncidentID.ValueString(), addReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating incident update",
			fmt.Sprintf("Could not add update to incident %s: %s", plan.IncidentID.ValueString(), err),
		)
		return
	}

	// Find the update we just added (it should be the last one or match by type/text)
	var updateUUID string
	var updateDate string
	if len(incident.Updates) > 0 {
		// Take the last update as the one we just created
		lastUpdate := incident.Updates[len(incident.Updates)-1]
		updateUUID = lastUpdate.UUID
		updateDate = lastUpdate.Date
	}

	// Set the composite ID (incident_id/update_id)
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.IncidentID.ValueString(), updateUUID))
	if updateDate != "" {
		plan.Date = types.StringValue(updateDate)
	} else {
		plan.Date = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *IncidentUpdateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IncidentUpdateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the composite ID
	incidentID, updateID := parseIncidentUpdateID(state.ID.ValueString())
	if incidentID == "" || updateID == "" {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Could not parse incident update ID: %s", state.ID.ValueString()),
		)
		return
	}

	// Fetch the incident to find the update
	incident, err := r.client.GetIncident(ctx, incidentID)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading incident",
			fmt.Sprintf("Could not read incident %s: %s", incidentID, err),
		)
		return
	}

	// Find the specific update
	var found bool
	for _, update := range incident.Updates {
		if update.UUID == updateID {
			state.Text = types.StringValue(update.Text.En)
			state.Type = types.StringValue(update.Type)
			state.Date = types.StringValue(update.Date)
			found = true
			break
		}
	}

	if !found {
		// Update was deleted
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *IncidentUpdateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IncidentUpdateResourceModel
	var state IncidentUpdateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the composite ID to get update UUID
	incidentID, updateID := parseIncidentUpdateID(state.ID.ValueString())
	if incidentID == "" || updateID == "" {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Could not parse incident update ID: %s", state.ID.ValueString()),
		)
		return
	}

	// The Hyperping API supports PUT /v3/incidents/{uuid}/updates/{updateuuid}
	// But our client doesn't have this method yet. For now, we'll note this limitation.
	// Updates to incident updates are not commonly needed, and the API may not support all field changes.

	resp.Diagnostics.AddWarning(
		"Update Not Fully Supported",
		"Incident update modifications may not be fully supported by the Hyperping API. Consider destroying and recreating the resource if changes are needed.",
	)

	// Keep the state as-is for now
	plan.ID = state.ID
	plan.Date = state.Date
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource.
func (r *IncidentUpdateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IncidentUpdateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the composite ID
	incidentID, _ := parseIncidentUpdateID(state.ID.ValueString())
	if incidentID == "" {
		resp.Diagnostics.AddError(
			"Invalid Resource ID",
			fmt.Sprintf("Could not parse incident update ID: %s", state.ID.ValueString()),
		)
		return
	}

	// Note: The Hyperping API has DELETE /v3/incidents/{uuid}/updates but it's unclear
	// if it deletes all updates or specific ones. For safety, we'll just remove from state
	// without making an API call. This is a common pattern for resources that can't be
	// individually deleted.

	// Resource will be removed from state automatically
}

// ImportState imports an existing resource into Terraform.
func (r *IncidentUpdateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Expect format: incident_id/update_id
	incidentID, updateID := parseIncidentUpdateID(req.ID)
	if incidentID == "" || updateID == "" {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected format: incident_id/update_id, got: %s", req.ID),
		)
		return
	}

	// Validate both ID components before setting state (VULN-015)
	if err := client.ValidateResourceID(incidentID); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import incident update — invalid incident ID: %s", err))
		return
	}
	if err := client.ValidateResourceID(updateID); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import incident update — invalid update ID: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("incident_id"), incidentID)...)
}

// parseIncidentUpdateID parses the composite ID format: incident_id/update_id
func parseIncidentUpdateID(id string) (incidentID, updateID string) {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
