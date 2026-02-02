// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &IncidentsDataSource{}
	_ datasource.DataSourceWithConfigure = &IncidentsDataSource{}
)

// NewIncidentsDataSource creates a new incidents data source.
func NewIncidentsDataSource() datasource.DataSource {
	return &IncidentsDataSource{}
}

// IncidentsDataSource defines the data source implementation.
type IncidentsDataSource struct {
	client client.IncidentAPI
}

// IncidentsDataSourceModel describes the data source data model.
type IncidentsDataSourceModel struct {
	Incidents []IncidentDataModel `tfsdk:"incidents"`
}

// IncidentDataModel describes a single incident in the data source.
type IncidentDataModel struct {
	ID                 types.String `tfsdk:"id"`
	Title              types.String `tfsdk:"title"`
	Text               types.String `tfsdk:"text"`
	Type               types.String `tfsdk:"type"`
	Date               types.String `tfsdk:"date"`
	AffectedComponents types.List   `tfsdk:"affected_components"`
	StatusPages        types.List   `tfsdk:"status_pages"`
	Updates            types.List   `tfsdk:"updates"`
}

// Metadata returns the data source type name.
func (d *IncidentsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incidents"
}

// IncidentUpdateAttrTypes returns the attribute types for incident updates.
func IncidentUpdateAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":   types.StringType,
		"date": types.StringType,
		"text": types.StringType,
		"type": types.StringType,
	}
}

// Schema defines the schema for the data source.
func (d *IncidentsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of all Hyperping incidents.",

		Attributes: map[string]schema.Attribute{
			"incidents": schema.ListNestedAttribute{
				MarkdownDescription: "List of incidents.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier (UUID) of the incident.",
							Computed:            true,
						},
						"title": schema.StringAttribute{
							MarkdownDescription: "The title of the incident (English).",
							Computed:            true,
						},
						"text": schema.StringAttribute{
							MarkdownDescription: "The description text of the incident (English).",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of incident (`outage` or `incident`).",
							Computed:            true,
						},
						"date": schema.StringAttribute{
							MarkdownDescription: "The date of the incident in ISO 8601 format.",
							Computed:            true,
						},
						"affected_components": schema.ListAttribute{
							MarkdownDescription: "List of component UUIDs affected by this incident.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"status_pages": schema.ListAttribute{
							MarkdownDescription: "List of status page UUIDs this incident is displayed on.",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"updates": schema.ListNestedAttribute{
							MarkdownDescription: "List of updates for this incident.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "The unique identifier of the update.",
										Computed:            true,
									},
									"date": schema.StringAttribute{
										MarkdownDescription: "The date of the update.",
										Computed:            true,
									},
									"text": schema.StringAttribute{
										MarkdownDescription: "The text content of the update (English).",
										Computed:            true,
									},
									"type": schema.StringAttribute{
										MarkdownDescription: "The type of update.",
										Computed:            true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *IncidentsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *IncidentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IncidentsDataSourceModel

	// Fetch all incidents from API
	incidents, err := d.client.ListIncidents(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading incidents",
			fmt.Sprintf("Could not list incidents: %s", err),
		)
		return
	}

	// Map response to model
	state.Incidents = make([]IncidentDataModel, len(incidents))
	for i, incident := range incidents {
		d.mapIncidentToDataModel(&incident, &state.Incidents[i], &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// mapIncidentToDataModel maps a client.Incident to the Terraform data model.
func (d *IncidentsDataSource) mapIncidentToDataModel(incident *client.Incident, model *IncidentDataModel, diags *diag.Diagnostics) {
	model.ID = types.StringValue(incident.UUID)
	model.Title = types.StringValue(incident.Title.En)
	model.Text = types.StringValue(incident.Text.En)
	model.Type = types.StringValue(incident.Type)

	if incident.Date != "" {
		model.Date = types.StringValue(incident.Date)
	} else {
		model.Date = types.StringNull()
	}

	model.AffectedComponents = mapStringSliceToList(incident.AffectedComponents, diags)
	model.StatusPages = mapStringSliceToList(incident.StatusPages, diags)

	// Map updates
	if len(incident.Updates) == 0 {
		model.Updates = types.ListNull(types.ObjectType{AttrTypes: IncidentUpdateAttrTypes()})
	} else {
		values := make([]attr.Value, len(incident.Updates))
		for i, u := range incident.Updates {
			obj, objDiags := types.ObjectValue(IncidentUpdateAttrTypes(), map[string]attr.Value{
				"id":   types.StringValue(u.UUID),
				"date": types.StringValue(u.Date),
				"text": types.StringValue(u.Text.En),
				"type": types.StringValue(u.Type),
			})
			diags.Append(objDiags...)
			values[i] = obj
		}
		list, listDiags := types.ListValue(types.ObjectType{AttrTypes: IncidentUpdateAttrTypes()}, values)
		diags.Append(listDiags...)
		model.Updates = list
	}
}
