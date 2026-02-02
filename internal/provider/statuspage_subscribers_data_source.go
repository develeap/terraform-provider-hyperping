// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &StatusPageSubscribersDataSource{}

func NewStatusPageSubscribersDataSource() datasource.DataSource {
	return &StatusPageSubscribersDataSource{}
}

// StatusPageSubscribersDataSource defines the data source implementation.
type StatusPageSubscribersDataSource struct {
	client client.HyperpingAPI
}

// StatusPageSubscribersDataSourceModel describes the data source data model.
type StatusPageSubscribersDataSourceModel struct {
	StatusPageUUID types.String `tfsdk:"statuspage_uuid"`
	Type           types.String `tfsdk:"type"`
	Page           types.Int64  `tfsdk:"page"`
	Subscribers    types.List   `tfsdk:"subscribers"`
	HasNextPage    types.Bool   `tfsdk:"has_next_page"`
	Total          types.Int64  `tfsdk:"total"`
}

func (d *StatusPageSubscribersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_statuspage_subscribers"
}

func (d *StatusPageSubscribersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches subscribers for a Hyperping status page.\n\n" +
			"This data source supports pagination and filtering by subscriber type.",

		Attributes: map[string]schema.Attribute{
			"statuspage_uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the status page",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Filter by subscriber type: all, email, sms, slack, teams",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("all", "email", "sms", "slack", "teams"),
				},
			},
			"page": schema.Int64Attribute{
				MarkdownDescription: "Page number (0-indexed) for pagination",
				Optional:            true,
			},
			"has_next_page": schema.BoolAttribute{
				MarkdownDescription: "Whether there are more pages available",
				Computed:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of subscribers",
				Computed:            true,
			},
			"subscribers": schema.ListNestedAttribute{
				MarkdownDescription: "List of subscribers",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							MarkdownDescription: "Subscriber ID",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Subscriber type (email, sms, slack, teams)",
							Computed:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Subscriber value (email address, phone number, etc.)",
							Computed:            true,
						},
						"email": schema.StringAttribute{
							MarkdownDescription: "Email address (for email subscribers)",
							Computed:            true,
						},
						"phone": schema.StringAttribute{
							MarkdownDescription: "Phone number (for SMS subscribers)",
							Computed:            true,
						},
						"slack_channel": schema.StringAttribute{
							MarkdownDescription: "Slack channel (for Slack subscribers)",
							Computed:            true,
						},
						"created_at": schema.StringAttribute{
							MarkdownDescription: "Creation timestamp",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *StatusPageSubscribersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(client.HyperpingAPI)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected client.HyperpingAPI, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = apiClient
}

func (d *StatusPageSubscribersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StatusPageSubscribersDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate status page UUID
	if err := client.ValidateResourceID(config.StatusPageUUID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Status Page UUID",
			fmt.Sprintf("Status page UUID must be valid: %s", err.Error()),
		)
		return
	}

	// Extract pagination and filter parameters
	var page *int
	if !isNullOrUnknown(config.Page) {
		p := int(config.Page.ValueInt64())
		page = &p
	}

	var subscriberType *string
	if !isNullOrUnknown(config.Type) {
		t := config.Type.ValueString()
		subscriberType = &t
	}

	// Fetch subscribers from API
	paginatedResp, err := d.client.ListSubscribers(ctx, config.StatusPageUUID.ValueString(), page, subscriberType)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading subscribers",
			fmt.Sprintf("Could not list subscribers for status page %s: %s",
				config.StatusPageUUID.ValueString(), err.Error()),
		)
		return
	}

	// Map pagination metadata
	config.HasNextPage = types.BoolValue(paginatedResp.HasNextPage)
	config.Total = types.Int64Value(int64(paginatedResp.Total))

	// Map subscribers to list
	subscribers := make([]SubscriberCommonFields, len(paginatedResp.Subscribers))
	for i, sub := range paginatedResp.Subscribers {
		subscribers[i] = mapSubscriberToTF(&sub, &resp.Diagnostics)
	}

	// Convert to Terraform list
	listValue, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: subscriberCommonFieldsAttrTypes()}, subscribers)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.Subscribers = listValue

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// subscriberCommonFieldsAttrTypes returns the attribute types for SubscriberCommonFields.
func subscriberCommonFieldsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":            types.Int64Type,
		"type":          types.StringType,
		"value":         types.StringType,
		"email":         types.StringType,
		"phone":         types.StringType,
		"slack_channel": types.StringType,
		"created_at":    types.StringType,
	}
}
