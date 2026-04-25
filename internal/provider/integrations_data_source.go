// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hyperping "github.com/develeap/hyperping-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &IntegrationsDataSource{}
	_ datasource.DataSourceWithConfigure = &IntegrationsDataSource{}
)

// NewIntegrationsDataSource creates a new integrations data source.
func NewIntegrationsDataSource() datasource.DataSource {
	return &IntegrationsDataSource{}
}

// IntegrationsDataSource defines the data source implementation.
type IntegrationsDataSource struct {
	client *hyperping.MCPClient
}

// IntegrationsDataSourceModel describes the data source data model.
type IntegrationsDataSourceModel struct {
	Integrations []IntegrationDataModel `tfsdk:"integrations"`
	IDs          types.List             `tfsdk:"ids"`
}

// IntegrationDataModel describes a single integration.
type IntegrationDataModel struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Type       types.String `tfsdk:"type"`
	Enabled    types.Bool   `tfsdk:"enabled"`
	LastTestAt types.String `tfsdk:"last_test_at"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

// Metadata returns the data source type name.
func (d *IntegrationsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integrations"
}

// Schema defines the schema for the data source.
func (d *IntegrationsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of all Hyperping integrations via MCP.",

		Attributes: map[string]schema.Attribute{
			"ids": schema.ListAttribute{
				MarkdownDescription: "List of integration UUIDs.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"integrations": schema.ListNestedAttribute{
				MarkdownDescription: "List of integrations.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier (UUID) of the integration.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the integration.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the integration (email, slack, webhook, pagerduty).",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the integration is enabled.",
							Computed:            true,
						},
						"last_test_at": schema.StringAttribute{
							MarkdownDescription: "The last time the integration was tested.",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "The time the integration was created.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *IntegrationsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	clients, ok := req.ProviderData.(*hyperpingClients)
	if !ok {
		resp.Diagnostics.Append(newUnexpectedConfigTypeError("*hyperpingClients", req.ProviderData))
		return
	}

	d.client = clients.MCP
}

// Read refreshes the Terraform state with the latest data.
func (d *IntegrationsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state IntegrationsDataSourceModel

	if d.client == nil {
		resp.Diagnostics.AddError("MCP Client Not Configured",
			"The MCP client was not initialized. Ensure the provider is configured with a valid api_key.")
		return
	}

	integrations, err := d.client.ListIntegrations(ctx)
	if err != nil {
		resp.Diagnostics.Append(NewReadErrorWithContext("Integrations", "", err))
		return
	}

	state.Integrations = make([]IntegrationDataModel, 0, len(integrations))
	ids := make([]attr.Value, 0, len(integrations))

	for _, i := range integrations {
		state.Integrations = append(state.Integrations, IntegrationDataModel{
			ID:         types.StringValue(i.UUID),
			Name:       types.StringValue(i.Name),
			Type:       types.StringValue(i.Type),
			Enabled:    types.BoolValue(i.Enabled),
			LastTestAt: types.StringValue(i.LastTestAt),
			CreatedAt:  types.StringValue(i.CreatedAt),
		})
		ids = append(ids, types.StringValue(i.UUID))
	}

	idList, diag := types.ListValue(types.StringType, ids)
	resp.Diagnostics.Append(diag...)
	state.IDs = idList

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
