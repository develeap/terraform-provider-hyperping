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
	_ datasource.DataSource              = &IncidentDataSource{}
	_ datasource.DataSourceWithConfigure = &IncidentDataSource{}
)

// NewIncidentDataSource creates a new single incident data source.
func NewIncidentDataSource() datasource.DataSource {
	return &IncidentDataSource{}
}

// IncidentDataSource defines the data source implementation for a single incident.
type IncidentDataSource struct {
	client client.IncidentAPI
}

// IncidentDataSourceModel describes the data source data model.
type IncidentDataSourceModel struct {
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
func (d *IncidentDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_incident"
}

// Schema defines the schema for the data source.
func (d *IncidentDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this data source to retrieve information about a specific Hyperping incident by its ID.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the incident to look up.",
				Required:            true,
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *IncidentDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *IncidentDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IncidentDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := client.ValidateResourceID(config.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Invalid Incident ID", fmt.Sprintf("Cannot look up incident: %s", err))
		return
	}

	incident, err := d.client.GetIncident(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(newReadError("Incident", config.ID.ValueString(), err))
		return
	}

	d.mapIncidentToDataSourceModel(incident, &config, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// mapIncidentToDataSourceModel maps a client.Incident to the data source model.
func (d *IncidentDataSource) mapIncidentToDataSourceModel(incident *client.Incident, model *IncidentDataSourceModel, diags *diag.Diagnostics) {
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
