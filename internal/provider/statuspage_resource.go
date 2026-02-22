// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
						Validators: []validator.String{
							URLFormat(),
						},
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
										MarkdownDescription: "Monitor UUID to display. Omit for group header entries (is_group=true).",
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
										MarkdownDescription: "Nested monitor services within this group (only used when is_group=true)",
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

	// Preserve write-only fields not returned by the API
	priorPassword := state.Password

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

	// Restore password: API never returns this field, so preserve prior state value
	// to prevent perpetual drift on every plan/apply cycle.
	if !priorPassword.IsNull() {
		state.Password = priorPassword
	}

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
// It extracts configured languages from the model's settings to filter API response localized fields,
// preventing drift from API auto-population of all supported languages.
func (r *StatusPageResource) mapStatusPageToModel(sp *client.StatusPage, model *StatusPageResourceModel, diags *diag.Diagnostics) {
	// Extract configured languages from the model's settings
	// This is used to filter localized fields in the API response
	configuredLangs := r.extractConfiguredLanguages(model.Settings, diags)

	// Preserve plan values BEFORE they get overwritten by API response
	// 1. settings.name - API returns resource.name in settings.name field
	var planSettingsName types.String
	if !model.Settings.IsNull() && !model.Settings.IsUnknown() {
		planAttrs := model.Settings.Attributes()
		if name, ok := planAttrs["name"].(types.String); ok && !name.IsNull() {
			planSettingsName = name
		}
	}

	// 2. sections - needed to preserve show_response_times boolean values
	planSections := model.Sections

	// Map with language filtering to prevent drift
	commonFields := MapStatusPageCommonFieldsWithFilter(sp, configuredLangs, diags)
	model.ID = commonFields.ID
	model.Name = commonFields.Name
	model.Hostname = commonFields.Hostname
	model.HostedSubdomain = commonFields.HostedSubdomain
	model.URL = commonFields.URL
	model.Settings = commonFields.Settings

	// Restore settings.name from plan to prevent API override drift
	if !planSettingsName.IsNull() && !planSettingsName.IsUnknown() {
		model.Settings = r.replaceSettingsName(model.Settings, planSettingsName, diags)
	}

	// Map sections from API
	model.Sections = commonFields.Sections

	// Preserve show_response_times and show_uptime boolean values from plan
	// API may return false even when true was set (API bug or default value issue)
	if !planSections.IsNull() && !planSections.IsUnknown() {
		model.Sections = r.preserveServiceBooleans(planSections, model.Sections, diags)
	}
}

// extractConfiguredLanguages extracts the list of configured languages from the settings object.
func (r *StatusPageResource) extractConfiguredLanguages(settings types.Object, diags *diag.Diagnostics) []string {
	if settings.IsNull() || settings.IsUnknown() {
		return nil
	}

	attrs := settings.Attributes()
	langsAttr, ok := attrs["languages"].(types.List)
	if !ok || langsAttr.IsNull() || langsAttr.IsUnknown() {
		return nil
	}

	return mapListToStringSlice(langsAttr, diags)
}

// replaceSettingsName replaces only the name field in a settings object while preserving all other fields.
// This is needed because the Hyperping API returns resource.name in settings.name, overriding the user's value.
func (r *StatusPageResource) replaceSettingsName(settings types.Object, name types.String, diags *diag.Diagnostics) types.Object {
	if settings.IsNull() || settings.IsUnknown() {
		return settings
	}

	// Get all current attributes
	attrs := settings.Attributes()

	// Create new map with all attributes, replacing only name
	newAttrs := make(map[string]attr.Value, len(attrs))
	for k, v := range attrs {
		if k == "name" {
			newAttrs[k] = name // Use plan value instead of API value
		} else {
			newAttrs[k] = v // Keep API value
		}
	}

	// Build new settings object with replaced name
	newSettings, newDiags := types.ObjectValue(StatusPageSettingsAttrTypes(), newAttrs)
	diags.Append(newDiags...)

	return newSettings
}

// preserveServiceBooleans preserves boolean field values from plan when API returns false.
// This handles API bugs where show_response_times and show_uptime may not persist correctly.
func (r *StatusPageResource) preserveServiceBooleans(planSections, apiSections types.List, diags *diag.Diagnostics) types.List {
	if planSections.IsNull() || planSections.IsUnknown() || apiSections.IsNull() || apiSections.IsUnknown() {
		return apiSections
	}

	planElements := planSections.Elements()
	apiElements := apiSections.Elements()

	// If lengths don't match, something is very wrong - just return API sections
	if len(planElements) != len(apiElements) {
		return apiSections
	}

	// Build new sections list with preserved boolean values
	newSections := make([]attr.Value, len(apiElements))

	for i := range apiElements {
		planSection, planOk := planElements[i].(types.Object)
		apiSection, apiOk := apiElements[i].(types.Object)

		if !planOk || !apiOk {
			newSections[i] = apiElements[i]
			continue
		}

		// Preserve booleans in services list
		newSection := r.preserveSectionServiceBooleans(planSection, apiSection, diags)
		newSections[i] = newSection
	}

	newList, newDiags := types.ListValue(apiSections.ElementType(context.Background()), newSections)
	diags.Append(newDiags...)

	return newList
}

// preserveSectionServiceBooleans preserves boolean values in a single section's services.
func (r *StatusPageResource) preserveSectionServiceBooleans(planSection, apiSection types.Object, diags *diag.Diagnostics) types.Object {
	planAttrs := planSection.Attributes()
	apiAttrs := apiSection.Attributes()

	planServices, planOk := planAttrs["services"].(types.List)
	apiServices, apiOk := apiAttrs["services"].(types.List)

	if !planOk || !apiOk || planServices.IsNull() || apiServices.IsNull() {
		return apiSection
	}

	// Preserve booleans in each service
	newServices := r.preserveServicesListBooleans(planServices, apiServices, diags)

	// Build new section with preserved services
	newAttrs := make(map[string]attr.Value, len(apiAttrs))
	for k, v := range apiAttrs {
		if k == "services" {
			newAttrs[k] = newServices
		} else {
			newAttrs[k] = v
		}
	}

	newSection, newDiags := types.ObjectValue(SectionAttrTypes(), newAttrs)
	diags.Append(newDiags...)

	return newSection
}

// preserveServicesListBooleans preserves boolean values in a services list.
func (r *StatusPageResource) preserveServicesListBooleans(planServices, apiServices types.List, diags *diag.Diagnostics) types.List {
	planElements := planServices.Elements()
	apiElements := apiServices.Elements()

	if len(planElements) != len(apiElements) {
		return apiServices
	}

	newServices := make([]attr.Value, len(apiElements))

	for i := range apiElements {
		planService, planOk := planElements[i].(types.Object)
		apiService, apiOk := apiElements[i].(types.Object)

		if !planOk || !apiOk {
			newServices[i] = apiElements[i]
			continue
		}

		planAttrs := planService.Attributes()
		apiAttrs := apiService.Attributes()

		// Check show_response_times: if plan=true and api=false, keep plan value
		needsPreservation := false
		newAttrs := make(map[string]attr.Value, len(apiAttrs))
		for k, v := range apiAttrs {
			if k == "show_response_times" || k == "show_uptime" {
				planVal, planHas := planAttrs[k].(types.Bool)
				apiVal, apiHas := v.(types.Bool)

				if planHas && apiHas && !planVal.IsNull() && !apiVal.IsNull() {
					// If plan had true but API returned false, preserve plan value
					if planVal.ValueBool() && !apiVal.ValueBool() {
						newAttrs[k] = planVal
						needsPreservation = true
						continue
					}
				}
			}
			newAttrs[k] = v
		}

		if needsPreservation {
			newService, newDiags := types.ObjectValue(ServiceAttrTypes(), newAttrs)
			diags.Append(newDiags...)
			newServices[i] = newService
		} else {
			newServices[i] = apiElements[i]
		}
	}

	newList, newDiags := types.ListValue(apiServices.ElementType(context.Background()), newServices)
	diags.Append(newDiags...)

	return newList
}

// buildCreateRequest builds a CreateStatusPageRequest from the Terraform plan.
func (r *StatusPageResource) buildCreateRequest(ctx context.Context, plan *StatusPageResourceModel, diags *diag.Diagnostics) *client.CreateStatusPageRequest {
	req := &client.CreateStatusPageRequest{
		Name: plan.Name.ValueString(),
	}

	subdomain := plan.HostedSubdomain.ValueString()
	req.Subdomain = &subdomain

	req.Hostname = extractOptionalStringPtr(plan.Hostname)
	req.Password = extractOptionalStringPtr(plan.Password)

	populateSettingsFields(ctx, plan.Settings, &statusPageSettingsTarget{
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
		DefaultLanguage:       &req.DefaultLanguage,
		Subscribe:             &req.Subscribe,
		Authentication:        &req.Authentication,
	}, diags)

	if !isNullOrUnknown(plan.Sections) {
		req.Sections = mapTFToSections(plan.Sections, diags)
	}

	return req
}

// buildUpdateRequest builds an UpdateStatusPageRequest from the Terraform plan.
func (r *StatusPageResource) buildUpdateRequest(ctx context.Context, plan *StatusPageResourceModel, diags *diag.Diagnostics) *client.UpdateStatusPageRequest {
	req := &client.UpdateStatusPageRequest{}

	req.Name = extractOptionalStringPtr(plan.Name)
	req.Subdomain = extractOptionalStringPtr(plan.HostedSubdomain)
	req.Hostname = extractOptionalStringPtr(plan.Hostname)
	req.Password = extractOptionalStringPtr(plan.Password)

	populateSettingsFields(ctx, plan.Settings, &statusPageSettingsTarget{
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
		DefaultLanguage:       &req.DefaultLanguage,
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
	DefaultLanguage       **string
	Subscribe             **client.CreateStatusPageSubscribeSettings
	Authentication        **client.CreateStatusPageAuthenticationSettings
}

// populateSettingsFields extracts all settings fields and populates the target request.
func populateSettingsFields(ctx context.Context, settings types.Object, target *statusPageSettingsTarget, diags *diag.Diagnostics) {
	if settings.IsNull() || settings.IsUnknown() {
		return
	}

	attrs := settings.Attributes()
	populateStringSettings(attrs, target)
	populateBoolSettings(attrs, target)
	populateCollectionSettings(attrs, target, diags)

	subscribe, auth := mapTFToSettings(ctx, settings, diags)
	*target.Subscribe = subscribe
	*target.Authentication = auth
}

// populateStringSettings populates all string settings fields from the attrs map into target.
// Handles: website, theme, font, accent_color, logo, logo_height, favicon, google_analytics, default_language.
func populateStringSettings(attrs map[string]attr.Value, target *statusPageSettingsTarget) {
	stringFields := []struct {
		key  string
		dest **string
	}{
		{"website", target.Website},
		{"theme", target.Theme},
		{"font", target.Font},
		{"accent_color", target.AccentColor},
		{"logo", target.Logo},
		{"logo_height", target.LogoHeight},
		{"favicon", target.Favicon},
		{"google_analytics", target.GoogleAnalytics},
		{"default_language", target.DefaultLanguage},
	}

	for _, f := range stringFields {
		if v, ok := attrs[f.key].(types.String); ok && !v.IsNull() {
			*f.dest = extractOptionalStringPtr(v)
		}
	}
}

// populateBoolSettings populates all bool settings fields from the attrs map into target.
// Handles: auto_refresh, banner_header, hide_powered_by, hide_from_search_engines.
func populateBoolSettings(attrs map[string]attr.Value, target *statusPageSettingsTarget) {
	boolFields := []struct {
		key  string
		dest **bool
	}{
		{"auto_refresh", target.AutoRefresh},
		{"banner_header", target.BannerHeader},
		{"hide_powered_by", target.HidePoweredBy},
		{"hide_from_search_engines", target.HideFromSearchEngines},
	}

	for _, f := range boolFields {
		if v, ok := attrs[f.key].(types.Bool); ok && !v.IsNull() {
			*f.dest = extractOptionalBoolPtr(v)
		}
	}
}

// populateCollectionSettings populates the description map and languages list into target.
// Handles: description (map[string]string), languages ([]string).
func populateCollectionSettings(attrs map[string]attr.Value, target *statusPageSettingsTarget, diags *diag.Diagnostics) {
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
}
