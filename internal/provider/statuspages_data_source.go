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
	StatusPages types.List             `tfsdk:"statuspages"`
	Page        types.Int64            `tfsdk:"page"`
	Search      types.String           `tfsdk:"search"`
	Filter      *StatusPageFilterModel `tfsdk:"filter"`
	HasNextPage types.Bool             `tfsdk:"has_next_page"`
	Total       types.Int64            `tfsdk:"total"`
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
				MarkdownDescription: "Search filter for name, hostname, or subdomain (server-side)",
				Optional:            true,
			},
			"filter": StatusPageFilterSchema(),
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
							MarkdownDescription: "Status page appearance and behavior settings",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Internal settings name",
									Computed:            true,
								},
								"website": schema.StringAttribute{
									MarkdownDescription: "Link to main website",
									Computed:            true,
								},
								"description": schema.StringAttribute{
									MarkdownDescription: "Status page description",
									Computed:            true,
								},
								"languages": schema.ListAttribute{
									MarkdownDescription: "Supported language codes",
									ElementType:         types.StringType,
									Computed:            true,
								},
								"default_language": schema.StringAttribute{
									MarkdownDescription: "Default language code",
									Computed:            true,
								},
								"theme": schema.StringAttribute{
									MarkdownDescription: "Color theme",
									Computed:            true,
								},
								"font": schema.StringAttribute{
									MarkdownDescription: "Font family",
									Computed:            true,
								},
								"accent_color": schema.StringAttribute{
									MarkdownDescription: "Accent color (hex)",
									Computed:            true,
								},
								"auto_refresh": schema.BoolAttribute{
									MarkdownDescription: "Auto-refresh enabled",
									Computed:            true,
								},
								"banner_header": schema.BoolAttribute{
									MarkdownDescription: "Banner header shown",
									Computed:            true,
								},
								"logo": schema.StringAttribute{
									MarkdownDescription: "Logo URL",
									Computed:            true,
								},
								"logo_height": schema.StringAttribute{
									MarkdownDescription: "Logo height",
									Computed:            true,
								},
								"favicon": schema.StringAttribute{
									MarkdownDescription: "Favicon URL",
									Computed:            true,
								},
								"hide_powered_by": schema.BoolAttribute{
									MarkdownDescription: "Hide powered by footer",
									Computed:            true,
								},
								"hide_from_search_engines": schema.BoolAttribute{
									MarkdownDescription: "Hide from search engines",
									Computed:            true,
								},
								"google_analytics": schema.StringAttribute{
									MarkdownDescription: "Google Analytics ID",
									Computed:            true,
								},
								"subscribe": schema.SingleNestedAttribute{
									MarkdownDescription: "Subscription settings",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"enabled": schema.BoolAttribute{MarkdownDescription: "Subscriptions enabled", Computed: true},
										"email":   schema.BoolAttribute{MarkdownDescription: "Email subscriptions allowed", Computed: true},
										"sms":     schema.BoolAttribute{MarkdownDescription: "SMS subscriptions allowed", Computed: true},
										"slack":   schema.BoolAttribute{MarkdownDescription: "Slack subscriptions allowed", Computed: true},
										"teams":   schema.BoolAttribute{MarkdownDescription: "Teams subscriptions allowed", Computed: true},
									},
								},
								"authentication": schema.SingleNestedAttribute{
									MarkdownDescription: "Access control settings",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"password_protection": schema.BoolAttribute{MarkdownDescription: "Password protection enabled", Computed: true},
										"google_sso":          schema.BoolAttribute{MarkdownDescription: "Google SSO enabled", Computed: true},
										"saml_sso":            schema.BoolAttribute{MarkdownDescription: "SAML SSO enabled", Computed: true},
										"allowed_domains": schema.ListAttribute{
											MarkdownDescription: "Allowed domains for SSO",
											ElementType:         types.StringType,
											Computed:            true,
										},
									},
								},
							},
						},
						"sections": schema.ListNestedAttribute{
							MarkdownDescription: "Status page sections",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.MapAttribute{
										MarkdownDescription: "Localized section name",
										ElementType:         types.StringType,
										Computed:            true,
									},
									"is_split": schema.BoolAttribute{
										MarkdownDescription: "Split services into separate rows",
										Computed:            true,
									},
									"services": schema.ListNestedAttribute{
										MarkdownDescription: "Services/monitors in this section",
										Computed:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"id": schema.StringAttribute{
													MarkdownDescription: "Service ID",
													Computed:            true,
												},
												"uuid": schema.StringAttribute{
													MarkdownDescription: "Monitor UUID",
													Computed:            true,
												},
												"name": schema.MapAttribute{
													MarkdownDescription: "Localized service name",
													ElementType:         types.StringType,
													Computed:            true,
												},
												"is_group": schema.BoolAttribute{
													MarkdownDescription: "Whether this is a group",
													Computed:            true,
												},
												"show_uptime": schema.BoolAttribute{
													MarkdownDescription: "Show uptime percentage",
													Computed:            true,
												},
												"show_response_times": schema.BoolAttribute{
													MarkdownDescription: "Show response times",
													Computed:            true,
												},
												"services": schema.ListNestedAttribute{
													MarkdownDescription: "Nested services within group",
													Computed:            true,
													NestedObject: schema.NestedAttributeObject{
														Attributes: map[string]schema.Attribute{
															"id": schema.StringAttribute{
																MarkdownDescription: "Service ID",
																Computed:            true,
															},
															"uuid": schema.StringAttribute{
																MarkdownDescription: "Monitor UUID",
																Computed:            true,
															},
															"name": schema.MapAttribute{
																MarkdownDescription: "Localized service name",
																ElementType:         types.StringType,
																Computed:            true,
															},
															"is_group": schema.BoolAttribute{
																MarkdownDescription: "Whether this is a group",
																Computed:            true,
															},
															"show_uptime": schema.BoolAttribute{
																MarkdownDescription: "Show uptime percentage",
																Computed:            true,
															},
															"show_response_times": schema.BoolAttribute{
																MarkdownDescription: "Show response times",
																Computed:            true,
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
				},
			},
		},
	}
}

// shouldIncludeStatusPage determines if a status page matches the filter criteria.
func (d *StatusPagesDataSource) shouldIncludeStatusPage(sp *client.StatusPage, filter *StatusPageFilterModel, diags *diag.Diagnostics) bool {
	return ApplyAllFilters(
		// Name regex filter
		func() bool {
			match, err := MatchesNameRegex(sp.Settings.Name, filter.NameRegex)
			if err != nil {
				diags.AddError(
					"Invalid filter regex",
					fmt.Sprintf("Failed to compile name_regex pattern: %s", err),
				)
				return false
			}
			return match
		},
		// Hostname filter
		func() bool {
			if sp.Hostname == nil {
				return isNullOrUnknown(filter.Hostname)
			}
			return MatchesExact(*sp.Hostname, filter.Hostname)
		},
	)
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

	// Apply client-side filtering if filter provided
	var filteredStatusPages []client.StatusPage
	if config.Filter != nil {
		for _, sp := range paginatedResp.StatusPages {
			if d.shouldIncludeStatusPage(&sp, config.Filter, &resp.Diagnostics) {
				if resp.Diagnostics.HasError() {
					return
				}
				filteredStatusPages = append(filteredStatusPages, sp)
			}
		}
	} else {
		filteredStatusPages = paginatedResp.StatusPages
	}

	// Map pagination metadata (update total if client-side filtering applied)
	config.HasNextPage = types.BoolValue(paginatedResp.HasNextPage)
	if config.Filter != nil {
		// If client-side filtering is applied, update total to reflect filtered count
		config.Total = types.Int64Value(int64(len(filteredStatusPages)))
	} else {
		config.Total = types.Int64Value(int64(paginatedResp.Total))
	}

	// Map status pages to list
	statusPages := make([]StatusPageCommonFields, len(filteredStatusPages))
	for i, sp := range filteredStatusPages {
		warnUnresolvedNumericUUIDs(&sp, &resp.Diagnostics)
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
