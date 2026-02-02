// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &StatusPagesDataSource{}

func NewStatusPagesDataSource() datasource.DataSource {
	return &StatusPagesDataSource{}
}

// StatusPagesDataSource defines the data source implementation.
type StatusPagesDataSource struct {
	client client.HyperpingAPI
}

// StatusPagesDataSourceModel describes the data source data model.
type StatusPagesDataSourceModel struct {
	StatusPages types.List   `tfsdk:"statuspages"`
	Page        types.Int64  `tfsdk:"page"`
	Search      types.String `tfsdk:"search"`
	HasNextPage types.Bool   `tfsdk:"has_next_page"`
	Total       types.Int64  `tfsdk:"total"`
}

func (d *StatusPagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_statuspages"
}

func (d *StatusPagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a list of Hyperping status pages with optional filtering.\n\n" +
			"This data source supports pagination and search filtering by name, hostname, or subdomain.",

		Attributes: map[string]schema.Attribute{
			"page": schema.Int64Attribute{
				MarkdownDescription: "Page number (0-indexed) for pagination. Defaults to 0.",
				Optional:            true,
			},
			"search": schema.StringAttribute{
				MarkdownDescription: "Search filter for name, hostname, or subdomain",
				Optional:            true,
			},
			"has_next_page": schema.BoolAttribute{
				MarkdownDescription: "Whether there are more pages available",
				Computed:            true,
			},
			"total": schema.Int64Attribute{
				MarkdownDescription: "Total number of status pages matching filters",
				Computed:            true,
			},
			"statuspages": schema.ListNestedAttribute{
				MarkdownDescription: "List of status pages",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Status page UUID",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Display name",
							Computed:            true,
						},
						"hostname": schema.StringAttribute{
							MarkdownDescription: "Custom domain",
							Computed:            true,
						},
						"hosted_subdomain": schema.StringAttribute{
							MarkdownDescription: "Hyperping subdomain",
							Computed:            true,
						},
						"url": schema.StringAttribute{
							MarkdownDescription: "Public URL",
							Computed:            true,
						},
						"settings": schema.SingleNestedAttribute{
							MarkdownDescription: "Settings",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									Computed: true,
								},
								"website": schema.StringAttribute{
									Computed: true,
								},
								"description": schema.MapAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
								"languages": schema.ListAttribute{
									ElementType: types.StringType,
									Computed:    true,
								},
								"default_language": schema.StringAttribute{
									Computed: true,
								},
								"theme": schema.StringAttribute{
									Computed: true,
								},
								"font": schema.StringAttribute{
									Computed: true,
								},
								"accent_color": schema.StringAttribute{
									Computed: true,
								},
								"auto_refresh": schema.BoolAttribute{
									Computed: true,
								},
								"banner_header": schema.BoolAttribute{
									Computed: true,
								},
								"logo": schema.StringAttribute{
									Computed: true,
								},
								"logo_height": schema.StringAttribute{
									Computed: true,
								},
								"favicon": schema.StringAttribute{
									Computed: true,
								},
								"hide_powered_by": schema.BoolAttribute{
									Computed: true,
								},
								"hide_from_search_engines": schema.BoolAttribute{
									Computed: true,
								},
								"google_analytics": schema.StringAttribute{
									Computed: true,
								},
								"subscribe": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"enabled": schema.BoolAttribute{Computed: true},
										"email":   schema.BoolAttribute{Computed: true},
										"sms":     schema.BoolAttribute{Computed: true},
										"slack":   schema.BoolAttribute{Computed: true},
										"teams":   schema.BoolAttribute{Computed: true},
									},
								},
								"authentication": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"password_protection": schema.BoolAttribute{Computed: true},
										"google_sso":          schema.BoolAttribute{Computed: true},
										"saml_sso":            schema.BoolAttribute{Computed: true},
										"allowed_domains": schema.ListAttribute{
											ElementType: types.StringType,
											Computed:    true,
										},
									},
								},
							},
						},
						"sections": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.MapAttribute{
										ElementType: types.StringType,
										Computed:    true,
									},
									"is_split": schema.BoolAttribute{
										Computed: true,
									},
									"services": schema.ListNestedAttribute{
										Computed: true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"id": schema.StringAttribute{
													Computed: true,
												},
												"uuid": schema.StringAttribute{
													Computed: true,
												},
												"name": schema.MapAttribute{
													ElementType: types.StringType,
													Computed:    true,
												},
												"is_group": schema.BoolAttribute{
													Computed: true,
												},
												"show_uptime": schema.BoolAttribute{
													Computed: true,
												},
												"show_response_times": schema.BoolAttribute{
													Computed: true,
												},
											},
										},
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

func (d *StatusPagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StatusPagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StatusPagesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Extract pagination and search parameters
	var page *int
	if !isNullOrUnknown(config.Page) {
		p := int(config.Page.ValueInt64())
		page = &p
	}

	var search *string
	if !isNullOrUnknown(config.Search) {
		s := config.Search.ValueString()
		search = &s
	}

	// Fetch status pages from API
	paginatedResp, err := d.client.ListStatusPages(ctx, page, search)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading status pages",
			fmt.Sprintf("Could not list status pages: %s", err.Error()),
		)
		return
	}

	// Map pagination metadata
	config.HasNextPage = types.BoolValue(paginatedResp.HasNextPage)
	config.Total = types.Int64Value(int64(paginatedResp.Total))

	// Map status pages to list
	statusPages := make([]StatusPageCommonFields, len(paginatedResp.StatusPages))
	for i, sp := range paginatedResp.StatusPages {
		statusPages[i] = MapStatusPageCommonFields(&sp, &resp.Diagnostics)
	}

	// Convert to Terraform list
	listValue, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: statusPageCommonFieldsAttrTypes()}, statusPages)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	config.StatusPages = listValue

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// statusPageCommonFieldsAttrTypes returns the attribute types for StatusPageCommonFields.
func statusPageCommonFieldsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":               types.StringType,
		"name":             types.StringType,
		"hostname":         types.StringType,
		"hosted_subdomain": types.StringType,
		"url":              types.StringType,
		"settings":         types.ObjectType{AttrTypes: StatusPageSettingsAttrTypes()},
		"sections":         types.ListType{ElemType: types.ObjectType{AttrTypes: SectionAttrTypes()}},
	}
}
