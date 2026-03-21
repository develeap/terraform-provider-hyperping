// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func (r *StatusPageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	apiClient, ok := req.ProviderData.(client.HyperpingAPI)
	if !ok {
		resp.Diagnostics.Append(newUnexpectedConfigTypeError("client.HyperpingAPI", req.ProviderData))
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

	// Translate mon_xxx -> numeric IDs for the uptime renderer
	maps, err := buildMonitorIDMaps(ctx, r.client.ListMonitors)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch monitors for ID translation", err.Error())
		return
	}
	if createReq.Sections != nil {
		translateSectionsUUIDsToNumericIDs(createReq.Sections, maps.uuidToNumericID, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Create status page via API
	statusPage, err := r.client.CreateStatusPage(ctx, *createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating status page", err.Error())
		return
	}

	// Translate numeric IDs back to mon_xxx for state
	translateResponseNumericIDsToUUIDs(statusPage, maps.numericIDToUUID, &resp.Diagnostics)

	// Save plan sections before mapping (contains write-only nested service fields)
	planSections := plan.Sections

	// Map response to state
	r.mapStatusPageToModel(ctx, statusPage, &plan, &resp.Diagnostics)

	// Restore write-only fields on nested services that the API doesn't return.
	// The API accepts description and show_response_times on write but may not
	// return them (or returns defaults) on read for deeply nested services.
	plan.Sections = preserveNestedServiceWriteOnlyFields(planSections, plan.Sections)

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
	priorSections := state.Sections

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

	// Translate numeric IDs to mon_xxx for state
	maps, mapErr := buildMonitorIDMaps(ctx, r.client.ListMonitors)
	if mapErr != nil {
		resp.Diagnostics.AddError("Failed to fetch monitors for ID translation", mapErr.Error())
		return
	}
	translateResponseNumericIDsToUUIDs(statusPage, maps.numericIDToUUID, &resp.Diagnostics)

	// Map response to state
	r.mapStatusPageToModel(ctx, statusPage, &state, &resp.Diagnostics)

	// Restore password: API never returns this field, so preserve prior state value
	if !priorPassword.IsNull() {
		state.Password = priorPassword
	}

	// Restore write-only fields on nested services
	state.Sections = preserveNestedServiceWriteOnlyFields(priorSections, state.Sections)

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

	// Translate mon_xxx -> numeric IDs for the uptime renderer
	maps, err := buildMonitorIDMaps(ctx, r.client.ListMonitors)
	if err != nil {
		resp.Diagnostics.AddError("Failed to fetch monitors for ID translation", err.Error())
		return
	}
	if updateReq.Sections != nil {
		translateSectionsUUIDsToNumericIDs(updateReq.Sections, maps.uuidToNumericID, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	// Update status page via API
	statusPage, err := r.client.UpdateStatusPage(ctx, state.ID.ValueString(), *updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating status page", err.Error())
		return
	}

	// Translate numeric IDs back to mon_xxx for state
	translateResponseNumericIDsToUUIDs(statusPage, maps.numericIDToUUID, &resp.Diagnostics)

	// Preserve write-only fields from plan before mapping
	planPassword := plan.Password
	planSections := plan.Sections

	// Map response to state
	r.mapStatusPageToModel(ctx, statusPage, &plan, &resp.Diagnostics)

	// Restore password: API never returns this field, so preserve plan value
	if !planPassword.IsNull() {
		plan.Password = planPassword
	}

	// Restore write-only fields on nested services
	plan.Sections = preserveNestedServiceWriteOnlyFields(planSections, plan.Sections)

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

// preserveNestedServiceWriteOnlyFields restores write-only fields on nested
// services (inside groups) from the plan/state. The Hyperping API accepts
// `description` and `show_response_times` on write for nested services but
// may not return them (or returns defaults) on read, causing Terraform's
// "inconsistent result after apply" error.
//
// This walks the sections -> services -> nested services tree and copies
// `description` and `show_response_times` from the configured values when
// the API response has null/default values.
func preserveNestedServiceWriteOnlyFields(configured, fromAPI types.List) types.List {
	if configured.IsNull() || configured.IsUnknown() || fromAPI.IsNull() || fromAPI.IsUnknown() {
		return fromAPI
	}

	configuredElems := configured.Elements()
	apiElems := fromAPI.Elements()

	if len(configuredElems) != len(apiElems) {
		return fromAPI
	}

	modified := false
	newElems := make([]attr.Value, len(apiElems))
	copy(newElems, apiElems)

	for i := range apiElems {
		configSection, ok1 := configuredElems[i].(types.Object)
		apiSection, ok2 := apiElems[i].(types.Object)
		if !ok1 || !ok2 {
			continue
		}

		configServices, ok1 := configSection.Attributes()["services"].(types.List)
		apiServices, ok2 := apiSection.Attributes()["services"].(types.List)
		if !ok1 || !ok2 || configServices.IsNull() || apiServices.IsNull() {
			continue
		}

		configSvcElems := configServices.Elements()
		apiSvcElems := apiServices.Elements()
		if len(configSvcElems) != len(apiSvcElems) {
			continue
		}

		svcModified := false
		newSvcElems := make([]attr.Value, len(apiSvcElems))
		copy(newSvcElems, apiSvcElems)

		for j := range apiSvcElems {
			configSvc, ok1 := configSvcElems[j].(types.Object)
			apiSvc, ok2 := apiSvcElems[j].(types.Object)
			if !ok1 || !ok2 {
				continue
			}

			configSvcAttrs := configSvc.Attributes()
			apiSvcAttrs := apiSvc.Attributes()

			// Preserve top-level service description if API doesn't return it
			configSvcDesc, hasConfigSvcDesc := configSvcAttrs["description"].(types.Map)
			apiSvcDesc, hasAPISvcDesc := apiSvcAttrs["description"].(types.Map)
			if hasConfigSvcDesc && !configSvcDesc.IsNull() && (!hasAPISvcDesc || apiSvcDesc.IsNull()) {
				newTopAttrs := make(map[string]attr.Value, len(apiSvcAttrs))
				for key, val := range apiSvcAttrs {
					newTopAttrs[key] = val
				}
				newTopAttrs["description"] = configSvcDesc
				newSvcObj, _ := types.ObjectValue(ServiceAttrTypes(), newTopAttrs)
				newSvcElems[j] = newSvcObj
				// Re-read apiSvc from the updated element for nested processing
				apiSvc = newSvcObj
				apiSvcAttrs = apiSvc.Attributes()
				svcModified = true
			}

			configNested, ok1 := configSvcAttrs["services"].(types.List)
			apiNested, ok2 := apiSvcAttrs["services"].(types.List)
			if !ok1 || !ok2 || configNested.IsNull() || apiNested.IsNull() {
				continue
			}

			configNestedElems := configNested.Elements()
			apiNestedElems := apiNested.Elements()
			if len(configNestedElems) != len(apiNestedElems) {
				continue
			}

			nestedModified := false
			newNestedElems := make([]attr.Value, len(apiNestedElems))
			copy(newNestedElems, apiNestedElems)

			for k := range apiNestedElems {
				configChild, ok1 := configNestedElems[k].(types.Object)
				apiChild, ok2 := apiNestedElems[k].(types.Object)
				if !ok1 || !ok2 {
					continue
				}

				configAttrs := configChild.Attributes()
				apiAttrs := apiChild.Attributes()
				newAttrs := make(map[string]attr.Value, len(apiAttrs))
				for key, val := range apiAttrs {
					newAttrs[key] = val
				}

				childModified := false

				// Preserve description: API may not return it for nested services
				configDesc, hasConfigDesc := configAttrs["description"].(types.Map)
				apiDesc, hasAPIDesc := apiAttrs["description"].(types.Map)
				if hasConfigDesc && !configDesc.IsNull() && (!hasAPIDesc || apiDesc.IsNull()) {
					newAttrs["description"] = configDesc
					childModified = true
				}

				// Preserve show_response_times: API may return wrong default for nested services
				configSRT, hasConfigSRT := configAttrs["show_response_times"].(types.Bool)
				apiSRT, hasAPISRT := apiAttrs["show_response_times"].(types.Bool)
				if hasConfigSRT && !configSRT.IsNull() && !configSRT.IsUnknown() &&
					hasAPISRT && configSRT.ValueBool() != apiSRT.ValueBool() {
					newAttrs["show_response_times"] = configSRT
					childModified = true
				}

				if childModified {
					newObj, _ := types.ObjectValue(NestedServiceAttrTypes(), newAttrs)
					newNestedElems[k] = newObj
					nestedModified = true
				}
			}

			if nestedModified {
				newNestedList, _ := types.ListValue(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}, newNestedElems)
				// Rebuild the parent service with updated nested services
				svcAttrs := make(map[string]attr.Value, len(apiSvc.Attributes()))
				for key, val := range apiSvc.Attributes() {
					svcAttrs[key] = val
				}
				svcAttrs["services"] = newNestedList
				newSvcObj, _ := types.ObjectValue(ServiceAttrTypes(), svcAttrs)
				newSvcElems[j] = newSvcObj
				svcModified = true
			}
		}

		if svcModified {
			newSvcList, _ := types.ListValue(types.ObjectType{AttrTypes: ServiceAttrTypes()}, newSvcElems)
			// Rebuild the section with updated services
			secAttrs := make(map[string]attr.Value, len(apiSection.Attributes()))
			for key, val := range apiSection.Attributes() {
				secAttrs[key] = val
			}
			secAttrs["services"] = newSvcList
			newSecObj, _ := types.ObjectValue(SectionAttrTypes(), secAttrs)
			newElems[i] = newSecObj
			modified = true
		}
	}

	if !modified {
		return fromAPI
	}

	newList, _ := types.ListValue(types.ObjectType{AttrTypes: SectionAttrTypes()}, newElems)
	return newList
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
func (r *StatusPageResource) mapStatusPageToModel(_ context.Context, sp *client.StatusPage, model *StatusPageResourceModel, diags *diag.Diagnostics) {
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

	// Map sections from API (is_split now returned correctly by the API)
	model.Sections = commonFields.Sections
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

// buildCreateRequest builds a CreateStatusPageRequest from the Terraform plan.
func (r *StatusPageResource) buildCreateRequest(ctx context.Context, plan *StatusPageResourceModel, diags *diag.Diagnostics) *client.CreateStatusPageRequest {
	req := &client.CreateStatusPageRequest{
		Name: plan.Name.ValueString(),
	}

	req.Subdomain = tfStringToPtr(plan.HostedSubdomain)

	req.Hostname = tfStringToPtr(plan.Hostname)
	req.Password = tfStringToPtr(plan.Password)

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

	req.Name = tfStringToPtr(plan.Name)
	req.Subdomain = tfStringToPtr(plan.HostedSubdomain)
	req.Hostname = tfStringToPtr(plan.Hostname)
	req.Password = tfStringToPtr(plan.Password)

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
	Description           **string
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
			*f.dest = tfStringToPtr(v)
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
			*f.dest = tfBoolToPtr(v)
		}
	}
}

// populateCollectionSettings populates the description string and languages list into target.
// Handles: description (plain string for API write), languages ([]string).
// API asymmetry: description is written as a plain string but read back as a localized map.
func populateCollectionSettings(attrs map[string]attr.Value, target *statusPageSettingsTarget, diags *diag.Diagnostics) {
	if descAttr, ok := attrs["description"].(types.String); ok && !descAttr.IsNull() && !descAttr.IsUnknown() {
		if v := descAttr.ValueString(); v != "" {
			*target.Description = &v
		}
	}

	if langsAttr, ok := attrs["languages"].(types.List); ok && !langsAttr.IsNull() {
		langs := mapListToStringSlice(langsAttr, diags)
		if len(langs) > 0 {
			*target.Languages = langs
		}
	}
}
