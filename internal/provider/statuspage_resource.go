// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StatusPageResource{}
var _ resource.ResourceWithImportState = &StatusPageResource{}

func NewStatusPageResource() resource.Resource {
	return &StatusPageResource{}
}

// StatusPageResource defines the resource implementation.
type StatusPageResource struct {
	client client.HyperpingAPI
}

// StatusPageResourceModel describes the resource data model.
type StatusPageResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Hostname        types.String `tfsdk:"hostname"`
	HostedSubdomain types.String `tfsdk:"hosted_subdomain"`
	URL             types.String `tfsdk:"url"`
	Password        types.String `tfsdk:"password"`
	Settings        types.Object `tfsdk:"settings"`
	Sections        types.List   `tfsdk:"sections"`
}

func (r *StatusPageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_statuspage"
}

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
			},
			"hostname": schema.StringAttribute{
				MarkdownDescription: "Custom domain for the status page (optional). If not provided, uses hosted subdomain.",
				Optional:            true,
				Computed:            true,
			},
			"hosted_subdomain": schema.StringAttribute{
				MarkdownDescription: "Hyperping-hosted subdomain (e.g., 'status' for status.hyperping.app)",
				Required:            true,
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
					},
					"description": schema.MapAttribute{
						MarkdownDescription: "Localized descriptions (language code -> text)",
						ElementType:         types.StringType,
						Optional:            true,
						Computed:            true,
					},
					"languages": schema.ListAttribute{
						MarkdownDescription: "Supported language codes (e.g., ['en', 'fr', 'de'])",
						ElementType:         types.StringType,
						Required:            true,
					},
					"default_language": schema.StringAttribute{
						MarkdownDescription: "Default language code",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("en"),
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
										MarkdownDescription: "Monitor UUID to display",
										Required:            true,
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
								},
							},
						},
					},
				},
			},
		},
	}
}

func (r *StatusPageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(client.HyperpingAPI)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected client.HyperpingAPI, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = apiClient
}

func (r *StatusPageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan StatusPageResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request from plan
	createReq := r.buildCreateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create status page via API
	statusPage, err := r.client.CreateStatusPage(ctx, *createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating status page", err.Error())
		return
	}

	// Map response to state
	r.mapStatusPageToModel(statusPage, &plan, &resp.Diagnostics)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StatusPageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state StatusPageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get status page from API
	statusPage, err := r.client.GetStatusPage(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			// Status page was deleted outside Terraform
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading status page", err.Error())
		return
	}

	// Map response to state
	r.mapStatusPageToModel(statusPage, &state, &resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *StatusPageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan StatusPageResourceModel
	var state StatusPageResourceModel

	// Read Terraform plan and current state
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build update request from plan
	updateReq := r.buildUpdateRequest(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Update status page via API
	statusPage, err := r.client.UpdateStatusPage(ctx, state.ID.ValueString(), *updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating status page", err.Error())
		return
	}

	// Map response to state
	r.mapStatusPageToModel(statusPage, &plan, &resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *StatusPageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state StatusPageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Delete status page via API
	err := r.client.DeleteStatusPage(ctx, state.ID.ValueString())
	if err != nil {
		if !client.IsNotFound(err) {
			resp.Diagnostics.AddError("Error deleting status page", err.Error())
			return
		}
		// Already deleted, continue
	}
}

func (r *StatusPageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Validate UUID format
	if err := client.ValidateResourceID(req.ID); err != nil {
		resp.Diagnostics.AddError(
			"Invalid Status Page ID",
			fmt.Sprintf("Status page ID must be a valid UUID (e.g., sp_abc123): %s", err.Error()),
		)
		return
	}

	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// mapStatusPageToModel maps API response to Terraform model.
func (r *StatusPageResource) mapStatusPageToModel(sp *client.StatusPage, model *StatusPageResourceModel, diags *diag.Diagnostics) {
	commonFields := MapStatusPageCommonFields(sp, diags)
	model.ID = commonFields.ID
	model.Name = commonFields.Name
	model.Hostname = commonFields.Hostname
	model.HostedSubdomain = commonFields.HostedSubdomain
	model.URL = commonFields.URL
	model.Settings = commonFields.Settings
	model.Sections = commonFields.Sections
}

// buildCreateRequest builds a CreateStatusPageRequest from the Terraform plan.
func (r *StatusPageResource) buildCreateRequest(_ context.Context, plan *StatusPageResourceModel, diags *diag.Diagnostics) *client.CreateStatusPageRequest {
	req := &client.CreateStatusPageRequest{
		Name: plan.Name.ValueString(),
	}

	subdomain := plan.HostedSubdomain.ValueString()
	req.Subdomain = &subdomain

	req.Hostname = extractOptionalStringPtr(plan.Hostname)
	req.Password = extractOptionalStringPtr(plan.Password)

	populateSettingsFields(plan.Settings, &statusPageSettingsTarget{
		Website:               &req.Website,
		Theme:                 &req.Theme,
		Font:                  &req.Font,
		AccentColor:           &req.AccentColor,
		Logo:                  &req.Logo,
		LogoHeight:            &req.LogoHeight,
		Favicon:               &req.Favicon,
		GoogleAnalytics:       &req.GoogleAnalytics,
		AutoRefresh:           &req.AutoRefresh,
		BannerHeader:          &req.BannerHeader,
		HidePoweredBy:         &req.HidePoweredBy,
		HideFromSearchEngines: &req.HideFromSearchEngines,
		Description:           &req.Description,
		Languages:             &req.Languages,
		Subscribe:             &req.Subscribe,
		Authentication:        &req.Authentication,
	}, diags)

	if !isNullOrUnknown(plan.Sections) {
		req.Sections = mapTFToSections(plan.Sections, diags)
	}

	return req
}

// buildUpdateRequest builds an UpdateStatusPageRequest from the Terraform plan.
func (r *StatusPageResource) buildUpdateRequest(_ context.Context, plan *StatusPageResourceModel, diags *diag.Diagnostics) *client.UpdateStatusPageRequest {
	req := &client.UpdateStatusPageRequest{}

	req.Name = extractOptionalStringPtr(plan.Name)
	req.Subdomain = extractOptionalStringPtr(plan.HostedSubdomain)
	req.Hostname = extractOptionalStringPtr(plan.Hostname)
	req.Password = extractOptionalStringPtr(plan.Password)

	populateSettingsFields(plan.Settings, &statusPageSettingsTarget{
		Website:               &req.Website,
		Theme:                 &req.Theme,
		Font:                  &req.Font,
		AccentColor:           &req.AccentColor,
		Logo:                  &req.Logo,
		LogoHeight:            &req.LogoHeight,
		Favicon:               &req.Favicon,
		GoogleAnalytics:       &req.GoogleAnalytics,
		AutoRefresh:           &req.AutoRefresh,
		BannerHeader:          &req.BannerHeader,
		HidePoweredBy:         &req.HidePoweredBy,
		HideFromSearchEngines: &req.HideFromSearchEngines,
		Description:           &req.Description,
		Languages:             &req.Languages,
		Subscribe:             &req.Subscribe,
		Authentication:        &req.Authentication,
	}, diags)

	if !isNullOrUnknown(plan.Sections) {
		req.Sections = mapTFToSections(plan.Sections, diags)
	}

	return req
}

// extractOptionalStringPtr extracts a string pointer from a types.String if not null/unknown.
// NOTE: This function is now deprecated in favor of tfStringToPtr from tf_helpers.go.
// It's kept for now to maintain compatibility but should be replaced with tfStringToPtr.
func extractOptionalStringPtr(value types.String) *string {
	return tfStringToPtr(value)
}

// extractOptionalBoolPtr extracts a bool pointer from a types.Bool if not null/unknown.
// NOTE: This function is now deprecated in favor of tfBoolToPtr from tf_helpers.go.
// It's kept for now to maintain compatibility but should be replaced with tfBoolToPtr.
func extractOptionalBoolPtr(value types.Bool) *bool {
	return tfBoolToPtr(value)
}

// statusPageSettingsTarget holds pointers to all settings fields in the request.
type statusPageSettingsTarget struct {
	Website               **string
	Theme                 **string
	Font                  **string
	AccentColor           **string
	Logo                  **string
	LogoHeight            **string
	Favicon               **string
	GoogleAnalytics       **string
	AutoRefresh           **bool
	BannerHeader          **bool
	HidePoweredBy         **bool
	HideFromSearchEngines **bool
	Description           *map[string]string
	Languages             *[]string
	Subscribe             **client.CreateStatusPageSubscribeSettings
	Authentication        **client.CreateStatusPageAuthenticationSettings
}

// populateSettingsFields extracts all settings fields and populates the target request.
func populateSettingsFields(settings types.Object, target *statusPageSettingsTarget, diags *diag.Diagnostics) {
	if settings.IsNull() || settings.IsUnknown() {
		return
	}

	attrs := settings.Attributes()

	if websiteAttr, ok := attrs["website"].(types.String); ok && !websiteAttr.IsNull() {
		*target.Website = extractOptionalStringPtr(websiteAttr)
	}

	if themeAttr, ok := attrs["theme"].(types.String); ok && !themeAttr.IsNull() {
		*target.Theme = extractOptionalStringPtr(themeAttr)
	}

	if fontAttr, ok := attrs["font"].(types.String); ok && !fontAttr.IsNull() {
		*target.Font = extractOptionalStringPtr(fontAttr)
	}

	if accentAttr, ok := attrs["accent_color"].(types.String); ok && !accentAttr.IsNull() {
		*target.AccentColor = extractOptionalStringPtr(accentAttr)
	}

	if logoAttr, ok := attrs["logo"].(types.String); ok && !logoAttr.IsNull() {
		*target.Logo = extractOptionalStringPtr(logoAttr)
	}

	if logoHeightAttr, ok := attrs["logo_height"].(types.String); ok && !logoHeightAttr.IsNull() {
		*target.LogoHeight = extractOptionalStringPtr(logoHeightAttr)
	}

	if faviconAttr, ok := attrs["favicon"].(types.String); ok && !faviconAttr.IsNull() {
		*target.Favicon = extractOptionalStringPtr(faviconAttr)
	}

	if gaAttr, ok := attrs["google_analytics"].(types.String); ok && !gaAttr.IsNull() {
		*target.GoogleAnalytics = extractOptionalStringPtr(gaAttr)
	}

	if autoRefreshAttr, ok := attrs["auto_refresh"].(types.Bool); ok && !autoRefreshAttr.IsNull() {
		*target.AutoRefresh = extractOptionalBoolPtr(autoRefreshAttr)
	}

	if bannerAttr, ok := attrs["banner_header"].(types.Bool); ok && !bannerAttr.IsNull() {
		*target.BannerHeader = extractOptionalBoolPtr(bannerAttr)
	}

	if hidePoweredByAttr, ok := attrs["hide_powered_by"].(types.Bool); ok && !hidePoweredByAttr.IsNull() {
		*target.HidePoweredBy = extractOptionalBoolPtr(hidePoweredByAttr)
	}

	if hideSearchAttr, ok := attrs["hide_from_search_engines"].(types.Bool); ok && !hideSearchAttr.IsNull() {
		*target.HideFromSearchEngines = extractOptionalBoolPtr(hideSearchAttr)
	}

	if descAttr, ok := attrs["description"].(types.Map); ok && !descAttr.IsNull() {
		desc := mapTFToStringMap(descAttr, diags)
		if len(desc) > 0 {
			*target.Description = desc
		}
	}

	if langsAttr, ok := attrs["languages"].(types.List); ok && !langsAttr.IsNull() {
		langs := mapListToStringSlice(langsAttr, diags)
		if len(langs) > 0 {
			*target.Languages = langs
		}
	}

	subscribe, auth := mapTFToSettings(settings, diags)
	*target.Subscribe = subscribe
	*target.Authentication = auth
}
