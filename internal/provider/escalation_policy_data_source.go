// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hyperping "github.com/develeap/hyperping-go"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &EscalationPolicyDataSource{}
	_ datasource.DataSourceWithConfigure = &EscalationPolicyDataSource{}
)

// NewEscalationPolicyDataSource creates a new single escalation policy data source.
func NewEscalationPolicyDataSource() datasource.DataSource {
	return &EscalationPolicyDataSource{}
}

// EscalationPolicyDataSource defines the data source implementation.
type EscalationPolicyDataSource struct {
	client *hyperping.MCPClient
}

// EscalationPolicyDataSourceModel describes the data source data model.
type EscalationPolicyDataSourceModel struct {
	ID    types.String `tfsdk:"id"`
	Name  types.String `tfsdk:"name"`
	Team  types.String `tfsdk:"team"`
	Steps types.List   `tfsdk:"steps"`
}

// Metadata returns the data source type name.
func (d *EscalationPolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_escalation_policy"
}

// Schema defines the schema for the data source.
func (d *EscalationPolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single Hyperping escalation policy by ID or name via MCP.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (UUID) of the policy.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("id"), path.MatchRoot("name")),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the policy to look up.",
				Optional:            true,
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *EscalationPolicyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *EscalationPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EscalationPolicyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policies, err := d.client.ListEscalationPolicies(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error fetching escalation policies", err.Error())
		return
	}

	var found *hyperping.EscalationPolicy
	if !data.ID.IsNull() {
		id := data.ID.ValueString()
		for _, p := range policies {
			if p.UUID == id {
				found = &p
				break
			}
		}
	} else if !data.Name.IsNull() {
		name := data.Name.ValueString()
		for _, p := range policies {
			if p.Name == name {
				found = &p
				break
			}
		}
	}

	if found == nil {
		resp.Diagnostics.AddError(
			"Escalation Policy Not Found",
			"Could not find an escalation policy matching the provided criteria.",
		)
		return
	}

	steps := make([]attr.Value, 0, len(found.Steps))
	for _, s := range found.Steps {
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

	data.ID = types.StringValue(found.UUID)
	data.Name = types.StringValue(found.Name)
	data.Team = types.StringValue(found.Team)
	data.Steps = stepList

	// Set state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
