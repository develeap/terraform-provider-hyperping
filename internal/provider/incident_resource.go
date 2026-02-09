// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &IncidentResource{}
	_ resource.ResourceWithImportState = &IncidentResource{}
)

// NewIncidentResource creates a new incident resource.
func NewIncidentResource() resource.Resource {
	return &IncidentResource{}
}

// IncidentResource defines the resource implementation.
type IncidentResource struct {
	client client.IncidentAPI
}

// IncidentResourceModel describes the resource data model.
type IncidentResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Title              types.String `tfsdk:"title"`
	Text               types.String `tfsdk:"text"`
	Type               types.String `tfsdk:"type"`
	AffectedComponents types.List   `tfsdk:"affected_components"`
	StatusPages        types.List   `tfsdk:"status_pages"`
	Date               types.String `tfsdk:"date"`
}

// Metadata returns the resource type name.
func (r *IncidentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incident"
}

// Schema defines the schema for the resource.
func (r *IncidentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hyperping incident for status page updates.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the incident.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "The title of the incident (English).",
				Required:            true,
			},
			"text": schema.StringAttribute{
				MarkdownDescription: "The description text of the incident (English).",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of incident. Valid values: `outage`, `incident`. Defaults to `incident`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("incident"),
				Validators: []validator.String{
					stringvalidator.OneOf(client.AllowedIncidentTypes...),
				},
			},
			"affected_components": schema.ListAttribute{
				MarkdownDescription: "List of component UUIDs affected by this incident.",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"status_pages": schema.ListAttribute{
				MarkdownDescription: "List of status page UUIDs to display this incident on. Required.",
				Required:            true,
				ElementType:         types.StringType,
			},
			"date": schema.StringAttribute{
				MarkdownDescription: "The date of the incident in ISO 8601 format (read-only).",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IncidentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *IncidentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IncidentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request with localized text
	createReq := client.CreateIncidentRequest{
		Title: client.LocalizedText{En: plan.Title.ValueString()},
		Text:  client.LocalizedText{En: plan.Text.ValueString()},
		Type:  plan.Type.ValueString(),
	}

	// Handle status_pages (required)
	var statusPages []string
	resp.Diagnostics.Append(plan.StatusPages.ElementsAs(ctx, &statusPages, false)...)
	if resp.Diagnostics.HasError() {
		return
	}
	createReq.StatusPages = statusPages

	// Handle optional affected_components
	if !isNullOrUnknown(plan.AffectedComponents) {
		var components []string
		resp.Diagnostics.Append(plan.AffectedComponents.ElementsAs(ctx, &components, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		createReq.AffectedComponents = components
	}

	// Call API to create incident
	createResp, err := r.client.CreateIncident(ctx, createReq)
	if err != nil {
		resp.Diagnostics.Append(newCreateError("Incident", err))
		return
	}

	// Read full incident details (create response only contains UUID)
	incident, err := r.client.GetIncident(ctx, createResp.UUID)
	if err != nil {
		resp.Diagnostics.Append(newReadAfterCreateError("Incident", createResp.UUID, err))
		return
	}

	// Map complete API response to Terraform state
	r.mapIncidentToModel(incident, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *IncidentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IncidentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	incident, err := r.client.GetIncident(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading incident",
			fmt.Sprintf("Could not read incident %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	r.mapIncidentToModel(incident, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state.
func (r *IncidentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan IncidentResourceModel
	var state IncidentResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request
	updateReq := client.UpdateIncidentRequest{}

	// Only include changed fields
	if !plan.Title.Equal(state.Title) {
		title := client.LocalizedText{En: plan.Title.ValueString()}
		updateReq.Title = &title
	}

	if !plan.Type.Equal(state.Type) {
		incidentType := plan.Type.ValueString()
		updateReq.Type = &incidentType
	}

	// Handle affected_components
	if !plan.AffectedComponents.Equal(state.AffectedComponents) {
		if plan.AffectedComponents.IsNull() {
			emptyComponents := []string{}
			updateReq.AffectedComponents = &emptyComponents
		} else {
			var components []string
			resp.Diagnostics.Append(plan.AffectedComponents.ElementsAs(ctx, &components, false)...)
			if resp.Diagnostics.HasError() {
				return
			}
			updateReq.AffectedComponents = &components
		}
	}

	// Handle status_pages
	if !plan.StatusPages.Equal(state.StatusPages) {
		var statusPages []string
		resp.Diagnostics.Append(plan.StatusPages.ElementsAs(ctx, &statusPages, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		updateReq.StatusPages = &statusPages
	}

	// Call API to update incident
	_, err := r.client.UpdateIncident(ctx, state.ID.ValueString(), updateReq)
	if err != nil {
		resp.Diagnostics.Append(newUpdateError("Incident", state.ID.ValueString(), err))
		return
	}

	// Read full incident details (update response doesn't contain complete data)
	incident, err := r.client.GetIncident(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(newReadAfterUpdateError("Incident", state.ID.ValueString(), err))
		return
	}

	// Map complete API response to Terraform state
	r.mapIncidentToModel(incident, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource.
func (r *IncidentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IncidentResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteIncident(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			// Already deleted, no error
			return
		}
		resp.Diagnostics.AddError(
			"Error deleting incident",
			fmt.Sprintf("Could not delete incident %s: %s", state.ID.ValueString(), err),
		)
		return
	}
}

// ImportState imports an existing resource into Terraform.
func (r *IncidentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Validate the import ID before setting state (VULN-015)
	if err := client.ValidateResourceID(req.ID); err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", fmt.Sprintf("Cannot import incident: %s", err))
		return
	}
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapIncidentToModel maps a client.Incident to the Terraform model.
func (r *IncidentResource) mapIncidentToModel(incident *client.Incident, model *IncidentResourceModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(incident.UUID)
	model.Title = types.StringValue(incident.Title.En)

	// NOTE: Text field behavior - Hyperping API quirk
	// The API accepts text during CREATE/UPDATE but may not return it in GET responses
	// If API returns it (non-empty), use that value; otherwise preserve plan value
	if incident.Text.En != "" {
		model.Text = types.StringValue(incident.Text.En)
	}
	// If empty and model.Text is already set (from plan), keep the existing value
	// This prevents state drift when API doesn't return the field

	model.Type = types.StringValue(incident.Type)

	// Handle date
	if incident.Date != "" {
		model.Date = types.StringValue(incident.Date)
	} else {
		model.Date = types.StringNull()
	}

	// Handle affected_components using shared helper
	model.AffectedComponents = mapStringSliceToList(incident.AffectedComponents, diags)

	// Handle status_pages using shared helper
	model.StatusPages = mapStringSliceToList(incident.StatusPages, diags)
}
