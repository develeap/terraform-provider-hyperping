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

	hyperping "github.com/develeap/hyperping-go"
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
	client hyperping.IncidentAPI
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
				Validators: []validator.String{
					StringLength(1, 10000),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of update. Valid values: `investigating`, `identified`, `update`, `monitoring`, `resolved`.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(hyperping.AllowedIncidentUpdateTypes...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"date": schema.StringAttribute{
				MarkdownDescription: "The date of the update in ISO 8601 format. If not provided, the current time is used.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IncidentUpdateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *IncidentUpdateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IncidentUpdateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build the request
	addReq := hyperping.AddIncidentUpdateRequest{
		Text: hyperping.LocalizedText{En: plan.Text.ValueString()},
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

	// Find the update we just added by matching type and text content.
	// Iterating in reverse since the newest update is most likely at the end.
	var updateUUID string
	var updateDate string
	for i := len(incident.Updates) - 1; i >= 0; i-- {
		u := incident.Updates[i]
		if u.Type == addReq.Type && u.Text.En == addReq.Text.En {
			updateUUID = u.UUID
			updateDate = u.Date
			break
		}
	}
	// Fallback to last update if no match found (e.g., API normalizes text)
	if updateUUID == "" && len(incident.Updates) > 0 {
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
		if hyperping.IsNotFound(err) {
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

// Update handles in-place updates. Since text, type, and incident_id all have
// RequiresReplace plan modifiers, the only field that can trigger an in-place
// update is date. Preserve computed fields from state and write the plan.
func (r *IncidentUpdateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IncidentUpdateResourceModel
	var state IncidentUpdateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve computed fields from current state.
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

	// The Hyperping API does not support deleting individual incident updates.
	// The resource is removed from Terraform state only — the update will persist
	// in Hyperping's incident timeline as a permanent audit record.
	resp.Diagnostics.AddWarning(
		"Incident Update Not Deleted From Hyperping",
		fmt.Sprintf("The incident update %s has been removed from Terraform state, "+
			"but it still exists in Hyperping's incident timeline for incident %s. "+
			"Incident updates are permanent audit records and cannot be deleted via the API.",
			state.ID.ValueString(), incidentID),
	)
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
	if err := hyperping.ValidateResourceID(incidentID); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import incident update — invalid incident ID: %s", err))
		return
	}
	if err := hyperping.ValidateResourceID(updateID); err != nil {
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
