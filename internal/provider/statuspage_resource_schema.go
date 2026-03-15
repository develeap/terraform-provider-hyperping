// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func (r *StatusPageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Hyperping status page.\n\n" +
			"Status pages provide a public or private view of your service health, " +
			"allowing you to communicate incidents and maintenance to your users.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Status page UUID (computed)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Display name for the status page",
				Required:            true,
				Validators: []validator.String{
					StringLength(1, 255),
				},
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Custom domain for the status page (optional). If not provided, uses hosted subdomain.",
				Optional:            true,
				Computed:            true,
			},
			"hosted_subdomain": schema.StringAttribute{
				MarkdownDescription: "Hyperping-hosted subdomain (e.g., 'status' for status.hyperping.app). " +
					"Optional when a custom `hostname` is set.",
				Optional: true,
				Computed: true,
			},
			"url": schema.StringAttribute{
				MarkdownDescription: "Public URL of the status page (computed)",
				Computed:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Password for password-protected status pages. Set this along with " +
					"`settings.authentication.password_protection = true` to require visitors to enter a password.",
				Optional:  true,
				Sensitive: true,
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Status page appearance and behavior settings",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Internal name for settings",
						Required:            true,
					},
					"website": schema.StringAttribute{
						MarkdownDescription: "Link to your main website",
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							URLFormat(),
						},
					},
					"description": schema.StringAttribute{
						MarkdownDescription: "Status page description. The API accepts a plain string on write; the read response wraps it in a localized map, from which the 'en' value (or default language) is used.",
						Optional:            true,
						Computed:            true,
					},
					"languages": schema.ListAttribute{
						MarkdownDescription: "Supported language codes (e.g., ['en', 'fr', 'de'])",
						ElementType:         types.StringType,
						Required:            true,
						Validators: []validator.List{
							listvalidator.ValueStringsAre(stringvalidator.OneOf(client.AllowedLanguages...)),
						},
					},
					"default_language": schema.StringAttribute{
						MarkdownDescription: "Default language code",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("en"),
						Validators: []validator.String{
							stringvalidator.OneOf(client.AllowedLanguages...),
						},
					},
					"theme": schema.StringAttribute{
						MarkdownDescription: "Color theme: light, dark, or system (default: system)",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("system"),
						Validators: []validator.String{
							stringvalidator.OneOf("light", "dark", "system"),
						},
					},
					"font": schema.StringAttribute{
						MarkdownDescription: "Font family (default: Inter)",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("Inter"),
						Validators: []validator.String{
							stringvalidator.OneOf(
								"system-ui", "Lato", "Manrope", "Inter", "Open Sans",
								"Montserrat", "Poppins", "Roboto", "Raleway", "Nunito",
								"Merriweather", "DM Sans", "Work Sans",
							),
						},
					},
					"accent_color": schema.StringAttribute{
						MarkdownDescription: "Accent color in hex format (default: #36b27e)",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("#36b27e"),
						Validators: []validator.String{
							HexColor(),
						},
					},
					"auto_refresh": schema.BoolAttribute{
						MarkdownDescription: "Enable auto-refresh of status page",
						Optional:            true,
						Computed:            true,
					},
					"banner_header": schema.BoolAttribute{
						MarkdownDescription: "Show banner header",
						Optional:            true,
						Computed:            true,
					},
					"logo": schema.StringAttribute{
						MarkdownDescription: "Logo URL",
						Optional:            true,
						Computed:            true,
					},
					"logo_height": schema.StringAttribute{
						MarkdownDescription: "Logo height (CSS value)",
						Optional:            true,
						Computed:            true,
					},
					"favicon": schema.StringAttribute{
						MarkdownDescription: "Favicon URL",
						Optional:            true,
						Computed:            true,
					},
					"hide_powered_by": schema.BoolAttribute{
						MarkdownDescription: "Hide 'Powered by Hyperping' footer",
						Optional:            true,
						Computed:            true,
					},
					"hide_from_search_engines": schema.BoolAttribute{
						MarkdownDescription: "Hide from search engines (noindex)",
						Optional:            true,
						Computed:            true,
					},
					"google_analytics": schema.StringAttribute{
						MarkdownDescription: "Google Analytics tracking ID",
						Optional:            true,
						Computed:            true,
					},
					"subscribe": schema.SingleNestedAttribute{
						MarkdownDescription: "Subscription settings",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable subscriptions",
								Optional:            true,
								Computed:            true,
							},
							"email": schema.BoolAttribute{
								MarkdownDescription: "Allow email subscriptions",
								Optional:            true,
								Computed:            true,
							},
							"sms": schema.BoolAttribute{
								MarkdownDescription: "Allow SMS subscriptions",
								Optional:            true,
								Computed:            true,
							},
							"slack": schema.BoolAttribute{
								MarkdownDescription: "Allow Slack subscriptions",
								Optional:            true,
								Computed:            true,
							},
							"teams": schema.BoolAttribute{
								MarkdownDescription: "Allow Microsoft Teams subscriptions",
								Optional:            true,
								Computed:            true,
							},
						},
					},
					"authentication": schema.SingleNestedAttribute{
						MarkdownDescription: "Access control settings",
						Optional:            true,
						Computed:            true,
						Attributes: map[string]schema.Attribute{
							"password_protection": schema.BoolAttribute{
								MarkdownDescription: "Enable password protection",
								Optional:            true,
								Computed:            true,
							},
							"google_sso": schema.BoolAttribute{
								MarkdownDescription: "Enable Google SSO",
								Optional:            true,
								Computed:            true,
							},
							"saml_sso": schema.BoolAttribute{
								MarkdownDescription: "Enable SAML SSO",
								Optional:            true,
								Computed:            true,
							},
							"allowed_domains": schema.ListAttribute{
								MarkdownDescription: "Allowed email domains for SSO",
								ElementType:         types.StringType,
								Optional:            true,
								Computed:            true,
							},
						},
					},
				},
			},
			"sections": schema.ListNestedAttribute{
				MarkdownDescription: "Status page sections containing monitors/services",
				Optional:            true,
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.MapAttribute{
							MarkdownDescription: "Localized section name (language code -> text)",
							ElementType:         types.StringType,
							Required:            true,
						},
						"is_split": schema.BoolAttribute{
							MarkdownDescription: "Split services in this section into separate rows",
							Optional:            true,
							Computed:            true,
						},
						"services": schema.ListNestedAttribute{
							MarkdownDescription: "Services/monitors in this section",
							Optional:            true,
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"id": schema.StringAttribute{
										MarkdownDescription: "Service ID (computed)",
										Computed:            true,
									},
									"uuid": schema.StringAttribute{
										MarkdownDescription: "Monitor UUID to display. Required for non-group services (is_group=false). Omit for group header entries (is_group=true).",
										Optional:            true,
										Computed:            true,
									},
									"name": schema.MapAttribute{
										MarkdownDescription: "Localized service name (language code -> text)",
										ElementType:         types.StringType,
										Optional:            true,
										Computed:            true,
									},
									"is_group": schema.BoolAttribute{
										MarkdownDescription: "Whether this service is a group containing nested services",
										Optional:            true,
										Computed:            true,
									},
									"show_uptime": schema.BoolAttribute{
										MarkdownDescription: "Show uptime percentage",
										Optional:            true,
										Computed:            true,
									},
									"show_response_times": schema.BoolAttribute{
										MarkdownDescription: "Show response times",
										Optional:            true,
										Computed:            true,
									},
									"services": schema.ListNestedAttribute{
										MarkdownDescription: "Nested monitor services within this group. Required when is_group=true; must contain at least one entry. Ignored when is_group=false.",
										Optional:            true,
										Computed:            true,
										NestedObject: schema.NestedAttributeObject{
											Attributes: map[string]schema.Attribute{
												"id": schema.StringAttribute{
													MarkdownDescription: "Service ID (computed)",
													Computed:            true,
												},
												"uuid": schema.StringAttribute{
													MarkdownDescription: "Monitor UUID to display",
													Optional:            true,
													Computed:            true,
												},
												"name": schema.MapAttribute{
													MarkdownDescription: "Localized service name (language code -> text)",
													ElementType:         types.StringType,
													Optional:            true,
													Computed:            true,
												},
												"is_group": schema.BoolAttribute{
													MarkdownDescription: "Whether this nested service is a group",
													Optional:            true,
													Computed:            true,
												},
												"show_uptime": schema.BoolAttribute{
													MarkdownDescription: "Show uptime percentage",
													Optional:            true,
													Computed:            true,
												},
												"show_response_times": schema.BoolAttribute{
													MarkdownDescription: "Show response times",
													Optional:            true,
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
	}
}
