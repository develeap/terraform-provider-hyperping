// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &StatusPageDataSource{}

func NewStatusPageDataSource() datasource.DataSource {
	return &StatusPageDataSource{}
}

// StatusPageDataSource defines the data source implementation.
type StatusPageDataSource struct {
	client client.HyperpingAPI
}

// StatusPageDataSourceModel describes the data source data model.
type StatusPageDataSourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Hostname        types.String `tfsdk:"hostname"`
	HostedSubdomain types.String `tfsdk:"hosted_subdomain"`
	URL             types.String `tfsdk:"url"`
	Settings        types.Object `tfsdk:"settings"`
	Sections        types.List   `tfsdk:"sections"`
}

func (d *StatusPageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_statuspage"
}

func (d *StatusPageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches a single Hyperping status page by UUID.\n\n" +
			"Use this data source to retrieve details about an existing status page, " +
			"including its settings, sections, and services.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Status page UUID to fetch",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name of the status page",
				Computed:            true,
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Custom domain for the status page",
				Computed:            true,
			},
			"hosted_subdomain": schema.StringAttribute{
				MarkdownDescription: "Hyperping-hosted subdomain",
				Computed:            true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Public URL of the status page",
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
					"description": schema.MapAttribute{
						MarkdownDescription: "Localized descriptions",
						ElementType:         types.StringType,
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
							"enabled": schema.BoolAttribute{
								MarkdownDescription: "Subscriptions enabled",
								Computed:            true,
							},
							"email": schema.BoolAttribute{
								MarkdownDescription: "Email subscriptions allowed",
								Computed:            true,
							},
							"sms": schema.BoolAttribute{
								MarkdownDescription: "SMS subscriptions allowed",
								Computed:            true,
							},
							"slack": schema.BoolAttribute{
								MarkdownDescription: "Slack subscriptions allowed",
								Computed:            true,
							},
							"teams": schema.BoolAttribute{
								MarkdownDescription: "Teams subscriptions allowed",
								Computed:            true,
							},
						},
					},
					"authentication": schema.SingleNestedAttribute{
						MarkdownDescription: "Access control settings",
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"password_protection": schema.BoolAttribute{
								MarkdownDescription: "Password protection enabled",
								Computed:            true,
							},
							"google_sso": schema.BoolAttribute{
								MarkdownDescription: "Google SSO enabled",
								Computed:            true,
							},
							"saml_sso": schema.BoolAttribute{
								MarkdownDescription: "SAML SSO enabled",
								Computed:            true,
							},
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
							MarkdownDescription: "Services split into rows",
							Computed:            true,
						},
						"services": schema.ListNestedAttribute{
							MarkdownDescription: "Services in section",
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
										MarkdownDescription: "Service is a group",
										Computed:            true,
									},
									"show_uptime": schema.BoolAttribute{
										MarkdownDescription: "Show uptime",
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
	}
}

func (d *StatusPageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *StatusPageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config StatusPageDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate UUID format
	if err := client.ValidateResourceID(config.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Status Page ID",
			fmt.Sprintf("Status page ID must be a valid UUID: %s", err.Error()),
		)
		return
	}

	// Fetch status page from API
	statusPage, err := d.client.GetStatusPage(ctx, config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.Append(newReadError("Status Page", config.ID.ValueString(), err))
		return
	}

	// Map API response to data source model
	d.mapStatusPageToModel(statusPage, &config, resp)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

// mapStatusPageToModel maps API response to data source model.
func (d *StatusPageDataSource) mapStatusPageToModel(sp *client.StatusPage, model *StatusPageDataSourceModel, resp *datasource.ReadResponse) {
	commonFields := MapStatusPageCommonFields(sp, &resp.Diagnostics)
	model.ID = commonFields.ID
	model.Name = commonFields.Name
	model.Hostname = commonFields.Hostname
	model.HostedSubdomain = commonFields.HostedSubdomain
	model.URL = commonFields.URL
	model.Settings = commonFields.Settings
	model.Sections = commonFields.Sections
}
