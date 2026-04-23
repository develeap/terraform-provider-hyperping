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
	_ datasource.DataSource              = &EscalationPoliciesDataSource{}
	_ datasource.DataSourceWithConfigure = &EscalationPoliciesDataSource{}
)

// NewEscalationPoliciesDataSource creates a new escalation policies data source.
func NewEscalationPoliciesDataSource() datasource.DataSource {
	return &EscalationPoliciesDataSource{}
}

// EscalationPoliciesDataSource defines the data source implementation.
type EscalationPoliciesDataSource struct {
	client *hyperping.MCPClient
}

// EscalationPoliciesDataSourceModel describes the data source data model.
type EscalationPoliciesDataSourceModel struct {
	Policies []EscalationPolicyDataModel `tfsdk:"policies"`
	IDs      types.List                  `tfsdk:"ids"`
}

// EscalationPolicyDataModel describes a single escalation policy.
type EscalationPolicyDataModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Team  types.String `tfsdk:"team"`
	Steps types.List   `tfsdk:"steps"`
}

// Metadata returns the data source type name.
func (d *EscalationPoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_escalation_policies"
}

// EscalationStepAttrTypes returns the attribute types for escalation steps.
func EscalationStepAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"delay":       types.Int64Type,
		"target_type": types.StringType,
		"target_id":   types.StringType,
	}
}

// Schema defines the schema for the data source.
func (d *EscalationPoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches the list of all Hyperping escalation policies via MCP.",

		Attributes: map[string]schema.Attribute{
			"ids": schema.ListAttribute{
				MarkdownDescription: "List of escalation policy UUIDs.",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"policies": schema.ListNestedAttribute{
				MarkdownDescription: "List of escalation policies.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The unique identifier (UUID) of the policy.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the policy.",
							Computed:            true,
						},
						"team": schema.StringAttribute{
							MarkdownDescription: "The team name this policy belongs to.",
							Computed:            true,
						},
						"steps": schema.ListNestedAttribute{
							MarkdownDescription: "The escalation steps.",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"delay": schema.Int64Attribute{
										MarkdownDescription: "Delay in minutes before the step is triggered.",
										Computed:            true,
									},
									"target_type": schema.StringAttribute{
										MarkdownDescription: "The type of the escalation target (user, schedule, integration).",
										Computed:            true,
									},
									"target_id": schema.StringAttribute{
										MarkdownDescription: "The ID of the escalation target.",
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
func (d *EscalationPoliciesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *EscalationPoliciesDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state EscalationPoliciesDataSourceModel

	policies, err := d.client.ListEscalationPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching escalation policies", err.Error())
		return
	}

	state.Policies = make([]EscalationPolicyDataModel, 0, len(policies))
	ids := make([]attr.Value, 0, len(policies))

	for _, p := range policies {
		steps := make([]attr.Value, 0, len(p.Steps))
		for _, s := range p.Steps {
			stepObj, diag := types.ObjectValue(EscalationStepAttrTypes(), map[string]attr.Value{
				"delay":       types.Int64Value(int64(s.Delay)),
				"target_type": types.StringValue(s.TargetType),
				"target_id":   types.StringValue(s.TargetID),
			})
			resp.Diagnostics.Append(diag...)
			steps = append(steps, stepObj)
		}

		stepList, diag := types.ListValue(types.ObjectType{AttrTypes: EscalationStepAttrTypes()}, steps)
		resp.Diagnostics.Append(diag...)

		state.Policies = append(state.Policies, EscalationPolicyDataModel{
			ID:    types.StringValue(p.UUID),
			Name:  types.StringValue(p.Name),
			Team:  types.StringValue(p.Team),
			Steps: stepList,
		})
		ids = append(ids, types.StringValue(p.UUID))
	}

	idList, diag := types.ListValue(types.StringType, ids)
	resp.Diagnostics.Append(diag...)
	state.IDs = idList

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
