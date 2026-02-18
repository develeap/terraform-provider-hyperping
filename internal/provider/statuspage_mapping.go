// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// =============================================================================
// Status Page Common Fields (shared by resource + data sources)
// =============================================================================

// StatusPageCommonFields contains fields shared between resource and data sources.
type StatusPageCommonFields struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Hostname        types.String `tfsdk:"hostname"`
	HostedSubdomain types.String `tfsdk:"hosted_subdomain"`
	URL             types.String `tfsdk:"url"`
	Settings        types.Object `tfsdk:"settings"`
	Sections        types.List   `tfsdk:"sections"`
}

// HyperpingSubdomainSuffix is the suffix appended to hosted subdomains by Hyperping API.
const HyperpingSubdomainSuffix = ".hyperping.app"

// normalizeSubdomain strips the .hyperping.app suffix from a subdomain if present.
// This ensures the Terraform state matches the user's configuration.
// Example: "mycompany.hyperping.app" -> "mycompany"
func normalizeSubdomain(subdomain string) string {
	if strings.HasSuffix(subdomain, HyperpingSubdomainSuffix) {
		return strings.TrimSuffix(subdomain, HyperpingSubdomainSuffix)
	}
	return subdomain
}

// MapStatusPageCommonFields maps common status page fields from API response to Terraform types.
// This is shared between StatusPageResource and StatusPage data sources to avoid duplication.
// Use MapStatusPageCommonFieldsWithFilter for resources that need localized field filtering.
func MapStatusPageCommonFields(sp *client.StatusPage, diags *diag.Diagnostics) StatusPageCommonFields {
	return MapStatusPageCommonFieldsWithFilter(sp, nil, diags)
}

// MapStatusPageCommonFieldsWithFilter maps common status page fields with optional language filtering.
// When configuredLangs is provided, localized fields (description, section/service names) are filtered
// to only include the configured languages, preventing drift from API auto-population.
func MapStatusPageCommonFieldsWithFilter(sp *client.StatusPage, configuredLangs []string, diags *diag.Diagnostics) StatusPageCommonFields {
	if sp == nil {
		return StatusPageCommonFields{
			ID:              types.StringNull(),
			Name:            types.StringNull(),
			Hostname:        types.StringNull(),
			HostedSubdomain: types.StringNull(),
			URL:             types.StringNull(),
			Settings:        types.ObjectNull(StatusPageSettingsAttrTypes()),
			Sections:        types.ListNull(types.ObjectType{AttrTypes: SectionAttrTypes()}),
		}
	}

	// Normalize subdomain by stripping .hyperping.app suffix
	// This ensures state matches the user's configuration
	normalizedSubdomain := normalizeSubdomain(sp.HostedSubdomain)

	result := StatusPageCommonFields{
		ID:              types.StringValue(sp.UUID),
		Name:            types.StringValue(sp.Name),
		HostedSubdomain: types.StringValue(normalizedSubdomain),
		URL:             types.StringValue(sp.URL),
	}

	// Handle optional hostname
	if sp.Hostname != nil && *sp.Hostname != "" {
		result.Hostname = types.StringValue(*sp.Hostname)
	} else {
		result.Hostname = types.StringNull()
	}

	// Map nested settings with optional language filtering
	result.Settings = mapSettingsToTFWithFilter(sp.Settings, configuredLangs, diags)

	// Map sections list with optional language filtering
	result.Sections = mapSectionsToTFWithFilter(sp.Sections, configuredLangs, diags)

	return result
}

// =============================================================================
// Settings Mapping (Nested Object)
// =============================================================================

// mapSettingsToTF converts API settings to Terraform Object type.
// For data sources that don't need language filtering.
func mapSettingsToTF(settings client.StatusPageSettings, diags *diag.Diagnostics) types.Object {
	return mapSettingsToTFWithFilter(settings, nil, diags)
}

// mapSettingsToTFWithFilter converts API settings to Terraform Object type with optional language filtering.
// When configuredLangs is provided, the description map is filtered to only include configured languages.
func mapSettingsToTFWithFilter(settings client.StatusPageSettings, configuredLangs []string, diags *diag.Diagnostics) types.Object {
	// Map subscribe settings
	subscribeObj, subDiags := types.ObjectValue(SubscribeSettingsAttrTypes(), map[string]attr.Value{
		"enabled": types.BoolValue(settings.Subscribe.Enabled),
		"email":   types.BoolValue(settings.Subscribe.Email),
		"slack":   types.BoolValue(settings.Subscribe.Slack),
		"teams":   types.BoolValue(settings.Subscribe.Teams),
		"sms":     types.BoolValue(settings.Subscribe.SMS),
	})
	diags.Append(subDiags...)

	// Map authentication settings
	authObj, authDiags := types.ObjectValue(AuthenticationSettingsAttrTypes(), map[string]attr.Value{
		"password_protection": types.BoolValue(settings.Authentication.PasswordProtection),
		"google_sso":          types.BoolValue(settings.Authentication.GoogleSSO),
		"saml_sso":            types.BoolValue(settings.Authentication.SAMLSSO),
		"allowed_domains":     mapStringSliceToList(settings.Authentication.AllowedDomains, diags),
	})
	diags.Append(authDiags...)

	// Map description (map[string]string) with optional language filtering
	// Filter to only configured languages to prevent drift from API auto-population
	filteredDescription := filterLocalizedMap(settings.Description, configuredLangs)
	descriptionMap := mapStringMapToTF(filteredDescription)

	// Map languages ([]string)
	languagesList := mapStringSliceToList(settings.Languages, diags)

	// Handle optional fields
	var logoValue, faviconValue, googleAnalyticsValue types.String
	if settings.Logo != nil && *settings.Logo != "" {
		logoValue = types.StringValue(*settings.Logo)
	} else {
		logoValue = types.StringNull()
	}

	if settings.Favicon != nil && *settings.Favicon != "" {
		faviconValue = types.StringValue(*settings.Favicon)
	} else {
		faviconValue = types.StringNull()
	}

	if settings.GoogleAnalytics != nil && *settings.GoogleAnalytics != "" {
		googleAnalyticsValue = types.StringValue(*settings.GoogleAnalytics)
	} else {
		googleAnalyticsValue = types.StringNull()
	}

	settingsObj, settingsDiags := types.ObjectValue(StatusPageSettingsAttrTypes(), map[string]attr.Value{
		"name":                     types.StringValue(settings.Name),
		"website":                  types.StringValue(settings.Website),
		"description":              descriptionMap,
		"languages":                languagesList,
		"default_language":         types.StringValue(settings.DefaultLanguage),
		"theme":                    types.StringValue(settings.Theme),
		"font":                     types.StringValue(settings.Font),
		"accent_color":             types.StringValue(settings.AccentColor),
		"auto_refresh":             types.BoolValue(settings.AutoRefresh),
		"banner_header":            types.BoolValue(settings.BannerHeader),
		"logo":                     logoValue,
		"logo_height":              types.StringValue(settings.LogoHeight),
		"favicon":                  faviconValue,
		"hide_powered_by":          types.BoolValue(settings.HidePoweredBy),
		"hide_from_search_engines": types.BoolValue(settings.HideFromSearchEngines),
		"google_analytics":         googleAnalyticsValue,
		"subscribe":                subscribeObj,
		"authentication":           authObj,
	})
	diags.Append(settingsDiags...)

	return settingsObj
}

// mapTFToSettings converts Terraform Object to API settings structures.
// Returns subscribe and authentication settings for create/update requests.
func mapTFToSettings(ctx context.Context, obj types.Object, diags *diag.Diagnostics) (*client.CreateStatusPageSubscribeSettings, *client.CreateStatusPageAuthenticationSettings) {
	if obj.IsNull() || obj.IsUnknown() {
		return nil, nil
	}

	attrs := obj.Attributes()

	subscribeObj, ok1 := attrs["subscribe"].(types.Object)
	if !ok1 {
		subscribeObj = types.ObjectNull(SubscribeSettingsAttrTypes())
	}
	authObj, ok2 := attrs["authentication"].(types.Object)
	if !ok2 {
		authObj = types.ObjectNull(AuthenticationSettingsAttrTypes())
	}

	return extractSubscribeSettings(subscribeObj, diags), extractAuthSettings(ctx, authObj, diags)
}

// extractSubscribeSettings converts a subscribe settings Terraform Object to the API struct.
// Returns nil when the object is null or unknown.
// Handles: enabled, email, slack, teams, sms.
func extractSubscribeSettings(obj types.Object, diags *diag.Diagnostics) *client.CreateStatusPageSubscribeSettings {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	_ = diags // reserved for future diagnostics
	attrs := obj.Attributes()
	subscribe := &client.CreateStatusPageSubscribeSettings{}

	if enabled, ok := attrs["enabled"].(types.Bool); ok && !enabled.IsNull() {
		val := enabled.ValueBool()
		subscribe.Enabled = &val
	}
	if email, ok := attrs["email"].(types.Bool); ok && !email.IsNull() {
		val := email.ValueBool()
		subscribe.Email = &val
	}
	if slack, ok := attrs["slack"].(types.Bool); ok && !slack.IsNull() {
		val := slack.ValueBool()
		subscribe.Slack = &val
	}
	if teams, ok := attrs["teams"].(types.Bool); ok && !teams.IsNull() {
		val := teams.ValueBool()
		subscribe.Teams = &val
	}
	if sms, ok := attrs["sms"].(types.Bool); ok && !sms.IsNull() {
		val := sms.ValueBool()
		subscribe.SMS = &val
	}

	return subscribe
}

// extractAuthSettings converts an authentication settings Terraform Object to the API struct.
// Returns nil when the object is null or unknown.
// Handles: password_protection, google_sso, saml_sso, allowed_domains.
func extractAuthSettings(ctx context.Context, obj types.Object, diags *diag.Diagnostics) *client.CreateStatusPageAuthenticationSettings {
	if obj.IsNull() || obj.IsUnknown() {
		return nil
	}

	attrs := obj.Attributes()
	authentication := &client.CreateStatusPageAuthenticationSettings{}

	if passwordProtection, ok := attrs["password_protection"].(types.Bool); ok && !passwordProtection.IsNull() {
		val := passwordProtection.ValueBool()
		authentication.PasswordProtection = &val
	}
	if googleSSO, ok := attrs["google_sso"].(types.Bool); ok && !googleSSO.IsNull() {
		val := googleSSO.ValueBool()
		authentication.GoogleSSO = &val
	}
	if samlSSO, ok := attrs["saml_sso"].(types.Bool); ok && !samlSSO.IsNull() {
		val := samlSSO.ValueBool()
		authentication.SAMLSSO = &val
	}
	if allowedDomains, ok := attrs["allowed_domains"].(types.List); ok && !isNullOrUnknown(allowedDomains) {
		var domains []string
		diags.Append(allowedDomains.ElementsAs(ctx, &domains, false)...)
		authentication.AllowedDomains = domains
	}

	return authentication
}

// =============================================================================
// Sections Mapping (List of Nested Objects with Recursive Services)
// =============================================================================

// mapSectionsToTF converts API sections array to Terraform List type.
// For data sources that don't need language filtering.
func mapSectionsToTF(sections []client.StatusPageSection, diags *diag.Diagnostics) types.List {
	return mapSectionsToTFWithFilter(sections, nil, diags)
}

// mapSectionsToTFWithFilter converts API sections array to Terraform List type with optional language filtering.
// When configuredLangs is provided, section and service names are filtered to only include configured languages.
func mapSectionsToTFWithFilter(sections []client.StatusPageSection, configuredLangs []string, diags *diag.Diagnostics) types.List {
	if len(sections) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: SectionAttrTypes()})
	}

	values := make([]attr.Value, len(sections))
	for i, section := range sections {
		// Map section name (map[string]string) with optional language filtering
		filteredName := filterLocalizedMap(section.Name, configuredLangs)
		nameMap := mapStringMapToTF(filteredName)

		// Map services list (recursive) with optional language filtering
		servicesList := mapServicesToTFWithFilter(section.Services, configuredLangs, diags)

		sectionObj, sectionDiags := types.ObjectValue(SectionAttrTypes(), map[string]attr.Value{
			"name":     nameMap,
			"is_split": types.BoolValue(section.IsSplit),
			"services": servicesList,
		})
		diags.Append(sectionDiags...)
		values[i] = sectionObj
	}

	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: SectionAttrTypes()}, values)
	diags.Append(listDiags...)
	return list
}

// mapServicesToTFWithFilter converts API services array to Terraform List type with optional language filtering.
// Pass nil for configuredLangs to include all languages (used by data sources).
func mapServicesToTFWithFilter(services []client.StatusPageService, configuredLangs []string, diags *diag.Diagnostics) types.List {
	// Use ServiceAttrTypes for elements since services may contain nested services
	attrs := ServiceAttrTypes()

	if len(services) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: attrs})
	}

	values := make([]attr.Value, len(services))
	for i, service := range services {
		values[i] = mapServiceToTFWithFilter(service, configuredLangs, diags)
	}

	list, listDiags := types.ListValue(types.ObjectType{AttrTypes: attrs}, values)
	diags.Append(listDiags...)
	return list
}

// mapServiceToTFWithFilter converts a single API service to Terraform Object type with optional language filtering.
// Pass nil for configuredLangs to include all languages (used by data sources).
func mapServiceToTFWithFilter(service client.StatusPageService, configuredLangs []string, diags *diag.Diagnostics) types.Object {
	// Map service name (map[string]string) with optional language filtering
	filteredName := filterLocalizedMap(service.Name, configuredLangs)
	nameMap := mapStringMapToTF(filteredName)

	attrs := ServiceAttrTypes()

	// Note: Nested services are not supported in the Terraform schema.
	// If the API returns nested services, they will be ignored.
	// Only the top-level service configuration is mapped.

	serviceObj, serviceDiags := types.ObjectValue(attrs, map[string]attr.Value{
		"id":                  types.StringValue(service.ID),
		"uuid":                types.StringValue(service.UUID),
		"name":                nameMap,
		"is_group":            types.BoolValue(service.IsGroup),
		"show_uptime":         types.BoolValue(service.ShowUptime),
		"show_response_times": types.BoolValue(service.ShowResponseTimes),
	})
	diags.Append(serviceDiags...)

	return serviceObj
}

// mapTFToSections converts Terraform List to API sections array.
func mapTFToSections(list types.List, diags *diag.Diagnostics) []client.CreateStatusPageSection {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	sections := make([]client.CreateStatusPageSection, 0, len(elements))

	for _, elem := range elements {
		obj, ok := elem.(types.Object)
		if !ok {
			diags.AddError("Invalid section element", "Expected object type for section element")
			continue
		}

		attrs := obj.Attributes()

		section := client.CreateStatusPageSection{}

		// Extract name (map[string]string -> string)
		// API expects string on create, but returns map on read
		if nameMap, ok := attrs["name"].(types.Map); ok && !nameMap.IsNull() {
			nameStrMap := mapTFToStringMap(nameMap, diags)
			if len(nameStrMap) > 0 {
				// Prefer "en" if available, otherwise take first value
				if enName, ok := nameStrMap["en"]; ok {
					section.Name = enName
				} else {
					for _, v := range nameStrMap {
						section.Name = v
						break
					}
				}
			}
		}

		// Extract is_split
		if isSplit, ok := attrs["is_split"].(types.Bool); ok && !isSplit.IsNull() {
			val := isSplit.ValueBool()
			section.IsSplit = &val
		}

		// Extract services (recursive)
		if servicesList, ok := attrs["services"].(types.List); ok && !servicesList.IsNull() {
			section.Services = mapTFToServices(servicesList, diags)
		}

		sections = append(sections, section)
	}

	return sections
}

// mapTFToServices converts Terraform List to API services array (recursive).
func mapTFToServices(list types.List, diags *diag.Diagnostics) []client.CreateStatusPageService {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	elements := list.Elements()
	services := make([]client.CreateStatusPageService, 0, len(elements))

	for _, elem := range elements {
		service := mapTFToService(elem, diags)
		services = append(services, service)
	}

	return services
}

// mapTFToService converts a Terraform Object to API service (recursive for nested services).
func mapTFToService(elem attr.Value, diags *diag.Diagnostics) client.CreateStatusPageService {
	obj, ok := elem.(types.Object)
	if !ok {
		diags.AddError("Invalid service element", "Expected object type for service element")
		return client.CreateStatusPageService{}
	}

	attrs := obj.Attributes()
	service := client.CreateStatusPageService{}

	// Extract monitor_uuid
	if monitorUUID, ok := attrs["uuid"].(types.String); ok && !monitorUUID.IsNull() {
		service.MonitorUUID = monitorUUID.ValueString()
	}

	// Extract name_shown (optional)
	if nameMap, ok := attrs["name"].(types.Map); ok && !nameMap.IsNull() {
		// For create requests, we only send name_shown as a string (not localized)
		// Extract the first value from the map as the display name
		nameStrMap := mapTFToStringMap(nameMap, diags)
		if len(nameStrMap) > 0 {
			// Get first value (typically "en")
			for _, v := range nameStrMap {
				nameShown := v
				service.NameShown = &nameShown
				break
			}
		}
	}

	// Extract show_uptime
	if showUptime, ok := attrs["show_uptime"].(types.Bool); ok && !showUptime.IsNull() {
		val := showUptime.ValueBool()
		service.ShowUptime = &val
	}

	// Extract show_response_times
	if showResponseTimes, ok := attrs["show_response_times"].(types.Bool); ok && !showResponseTimes.IsNull() {
		val := showResponseTimes.ValueBool()
		service.ShowResponseTimes = &val
	}

	// Extract is_group
	if isGroup, ok := attrs["is_group"].(types.Bool); ok && !isGroup.IsNull() {
		val := isGroup.ValueBool()
		service.IsGroup = &val
	}

	// Note: Nested services are not supported in the Terraform schema.
	// The services field has been removed from the schema.

	return service
}

// =============================================================================
// Map[string]string Helpers (for multi-language fields)
// =============================================================================

// mapStringMapToTF converts a Go map[string]string to Terraform Map type.
func mapStringMapToTF(m map[string]string) types.Map {
	if len(m) == 0 {
		return types.MapNull(types.StringType)
	}

	values := make(map[string]attr.Value, len(m))
	for k, v := range m {
		values[k] = types.StringValue(v)
	}

	result, _ := types.MapValue(types.StringType, values)
	return result
}

// filterLocalizedMap filters a localized map to only include configured languages.
// This prevents drift when the API auto-populates all languages but TF only configured some.
// If configuredLangs is nil or empty, returns the original map unfiltered.
func filterLocalizedMap(m map[string]string, configuredLangs []string) map[string]string {
	if len(configuredLangs) == 0 || len(m) == 0 {
		return m
	}

	// Build lookup set for configured languages
	langSet := make(map[string]bool, len(configuredLangs))
	for _, lang := range configuredLangs {
		langSet[lang] = true
	}

	// Filter to only configured languages
	filtered := make(map[string]string)
	for k, v := range m {
		if langSet[k] {
			filtered[k] = v
		}
	}

	return filtered
}

// mapTFToStringMap converts Terraform Map to Go map[string]string.
func mapTFToStringMap(tfMap types.Map, diags *diag.Diagnostics) map[string]string {
	if tfMap.IsNull() || tfMap.IsUnknown() {
		return nil
	}

	elements := tfMap.Elements()
	result := make(map[string]string, len(elements))

	for k, v := range elements {
		strVal, ok := v.(types.String)
		if !ok {
			diags.AddError("Invalid map value", "Expected string type in map")
			continue
		}
		if !strVal.IsNull() {
			result[k] = strVal.ValueString()
		}
	}

	return result
}

// mapListToStringSlice converts a types.List to []string.
func mapListToStringSlice(list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return []string{}
	}

	elements := list.Elements()
	result := make([]string, 0, len(elements))

	for i, elem := range elements {
		strVal, ok := elem.(types.String)
		if !ok {
			diags.AddError(
				"Invalid list element type",
				fmt.Sprintf("Expected string value at index %d, got %T", i, elem),
			)
			continue
		}
		result = append(result, strVal.ValueString())
	}

	return result
}

// =============================================================================
// Subscriber Mapping
// =============================================================================

// mapSubscriberToTF converts API subscriber to Terraform model fields.
func mapSubscriberToTF(sub *client.StatusPageSubscriber, _ *diag.Diagnostics) SubscriberCommonFields {
	if sub == nil {
		return SubscriberCommonFields{
			ID:           types.Int64Null(),
			Type:         types.StringNull(),
			Value:        types.StringNull(),
			Email:        types.StringNull(),
			Phone:        types.StringNull(),
			SlackChannel: types.StringNull(),
			CreatedAt:    types.StringNull(),
		}
	}

	result := SubscriberCommonFields{
		ID:        types.Int64Value(int64(sub.ID)),
		Type:      types.StringValue(sub.Type),
		Value:     types.StringValue(sub.Value),
		CreatedAt: types.StringValue(sub.CreatedAt),
	}

	// Handle optional email
	if sub.Email != nil && *sub.Email != "" {
		result.Email = types.StringValue(*sub.Email)
	} else {
		result.Email = types.StringNull()
	}

	// Handle optional phone
	if sub.Phone != nil && *sub.Phone != "" {
		result.Phone = types.StringValue(*sub.Phone)
	} else {
		result.Phone = types.StringNull()
	}

	// Handle optional slack_channel
	if sub.SlackChannel != nil && *sub.SlackChannel != "" {
		result.SlackChannel = types.StringValue(*sub.SlackChannel)
	} else {
		result.SlackChannel = types.StringNull()
	}

	return result
}

// SubscriberCommonFields contains fields shared between subscriber resource and data sources.
type SubscriberCommonFields struct {
	ID           types.Int64  `tfsdk:"id"`
	Type         types.String `tfsdk:"type"`
	Value        types.String `tfsdk:"value"`
	Email        types.String `tfsdk:"email"`
	Phone        types.String `tfsdk:"phone"`
	SlackChannel types.String `tfsdk:"slack_channel"`
	CreatedAt    types.String `tfsdk:"created_at"`
}

// =============================================================================
// AttrTypes Helpers (for schema definitions)
// =============================================================================

// StatusPageSettingsAttrTypes returns the attribute types for settings nested object.
func StatusPageSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                     types.StringType,
		"website":                  types.StringType,
		"description":              types.MapType{ElemType: types.StringType},
		"languages":                types.ListType{ElemType: types.StringType},
		"default_language":         types.StringType,
		"theme":                    types.StringType,
		"font":                     types.StringType,
		"accent_color":             types.StringType,
		"auto_refresh":             types.BoolType,
		"banner_header":            types.BoolType,
		"logo":                     types.StringType,
		"logo_height":              types.StringType,
		"favicon":                  types.StringType,
		"hide_powered_by":          types.BoolType,
		"hide_from_search_engines": types.BoolType,
		"google_analytics":         types.StringType,
		"subscribe":                types.ObjectType{AttrTypes: SubscribeSettingsAttrTypes()},
		"authentication":           types.ObjectType{AttrTypes: AuthenticationSettingsAttrTypes()},
	}
}

// SubscribeSettingsAttrTypes returns the attribute types for subscribe settings nested object.
func SubscribeSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"enabled": types.BoolType,
		"email":   types.BoolType,
		"slack":   types.BoolType,
		"teams":   types.BoolType,
		"sms":     types.BoolType,
	}
}

// AuthenticationSettingsAttrTypes returns the attribute types for authentication settings nested object.
func AuthenticationSettingsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"password_protection": types.BoolType,
		"google_sso":          types.BoolType,
		"saml_sso":            types.BoolType,
		"allowed_domains":     types.ListType{ElemType: types.StringType},
	}
}

// SectionAttrTypes returns the attribute types for section nested object.
func SectionAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":     types.MapType{ElemType: types.StringType},
		"is_split": types.BoolType,
		"services": types.ListType{ElemType: types.ObjectType{AttrTypes: ServiceAttrTypes()}},
	}
}

// ServiceAttrTypes returns the attribute types for service nested object (supports recursion).
func ServiceAttrTypes() map[string]attr.Type {
	// Note: Deeply nested services are not supported due to Terraform Plugin Framework
	// limitations with recursive DynamicType inside collections.
	return map[string]attr.Type{
		"id":                  types.StringType,
		"uuid":                types.StringType,
		"name":                types.MapType{ElemType: types.StringType},
		"is_group":            types.BoolType,
		"show_uptime":         types.BoolType,
		"show_response_times": types.BoolType,
	}
}
