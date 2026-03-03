// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// writeOnlyBooleans holds the plan values for booleans the API doesn't persist.
type writeOnlyBooleans struct {
	showUptime        *bool
	showResponseTimes *bool
}

// extractWriteOnlyBooleans walks the plan sections tree and extracts write-only
// boolean values keyed by service UUID. Also extracts is_split per section index.
// Returns empty maps for null/unknown plans (e.g. import scenarios).
func extractWriteOnlyBooleans(planSections types.List) (serviceMap map[string]writeOnlyBooleans, sectionIsSplit map[int]bool) {
	serviceMap = make(map[string]writeOnlyBooleans)
	sectionIsSplit = make(map[int]bool)

	if planSections.IsNull() || planSections.IsUnknown() {
		return serviceMap, sectionIsSplit
	}

	for secIdx, secElem := range planSections.Elements() {
		secObj, ok := secElem.(types.Object)
		if !ok {
			continue
		}
		secAttrs := secObj.Attributes()

		// Extract is_split for this section
		if isSplit, isSplitOK := secAttrs["is_split"].(types.Bool); isSplitOK && !isSplit.IsNull() && isSplit.ValueBool() {
			sectionIsSplit[secIdx] = true
		}

		// Extract service booleans
		servicesList, svcOK := secAttrs["services"].(types.List)
		if !svcOK || servicesList.IsNull() || servicesList.IsUnknown() {
			continue
		}

		extractServicesWriteOnlyBooleans(servicesList, serviceMap)
	}

	return serviceMap, sectionIsSplit
}

// extractServicesWriteOnlyBooleans extracts write-only booleans from a services list.
func extractServicesWriteOnlyBooleans(servicesList types.List, serviceMap map[string]writeOnlyBooleans) {
	for _, svcElem := range servicesList.Elements() {
		svcObj, ok := svcElem.(types.Object)
		if !ok {
			continue
		}
		svcAttrs := svcObj.Attributes()

		// Get UUID for this service
		uuid := ""
		if uuidAttr, ok := svcAttrs["uuid"].(types.String); ok && !uuidAttr.IsNull() {
			uuid = uuidAttr.ValueString()
		}

		if uuid != "" {
			wb := writeOnlyBooleans{}
			if su, ok := svcAttrs["show_uptime"].(types.Bool); ok && !su.IsNull() {
				val := su.ValueBool()
				wb.showUptime = &val
			}
			if srt, ok := svcAttrs["show_response_times"].(types.Bool); ok && !srt.IsNull() {
				val := srt.ValueBool()
				wb.showResponseTimes = &val
			}
			serviceMap[uuid] = wb
		}

		// Recurse into nested services (groups)
		if nestedList, ok := svcAttrs["services"].(types.List); ok && !nestedList.IsNull() && !nestedList.IsUnknown() {
			extractNestedServicesWriteOnlyBooleans(nestedList, serviceMap)
		}
	}
}

// extractNestedServicesWriteOnlyBooleans extracts write-only booleans from nested services.
func extractNestedServicesWriteOnlyBooleans(nestedList types.List, serviceMap map[string]writeOnlyBooleans) {
	for _, nestedElem := range nestedList.Elements() {
		nestedObj, ok := nestedElem.(types.Object)
		if !ok {
			continue
		}
		nestedAttrs := nestedObj.Attributes()

		uuid := ""
		if uuidAttr, ok := nestedAttrs["uuid"].(types.String); ok && !uuidAttr.IsNull() {
			uuid = uuidAttr.ValueString()
		}

		if uuid != "" {
			wb := writeOnlyBooleans{}
			if su, ok := nestedAttrs["show_uptime"].(types.Bool); ok && !su.IsNull() {
				val := su.ValueBool()
				wb.showUptime = &val
			}
			if srt, ok := nestedAttrs["show_response_times"].(types.Bool); ok && !srt.IsNull() {
				val := srt.ValueBool()
				wb.showResponseTimes = &val
			}
			serviceMap[uuid] = wb
		}
	}
}

// applyWriteOnlyBooleans walks the API response sections and overrides write-only
// boolean values using UUID-based matching from the plan. This is resilient to
// reordering and length mismatches.
func (r *StatusPageResource) applyWriteOnlyBooleans(apiSections types.List, serviceMap map[string]writeOnlyBooleans, sectionIsSplit map[int]bool, diags *diag.Diagnostics) types.List {
	if apiSections.IsNull() || apiSections.IsUnknown() {
		return apiSections
	}

	apiElements := apiSections.Elements()
	newSections := make([]attr.Value, len(apiElements))

	for i, secElem := range apiElements {
		secObj, ok := secElem.(types.Object)
		if !ok {
			newSections[i] = secElem
			continue
		}

		secAttrs := secObj.Attributes()
		newAttrs := copyAttrs(secAttrs)

		// Preserve is_split by section index
		if sectionIsSplit[i] {
			if apiIsSplit, ok := secAttrs["is_split"].(types.Bool); ok && !apiIsSplit.IsNull() && !apiIsSplit.ValueBool() {
				newAttrs["is_split"] = types.BoolValue(true)
			}
		}

		// Preserve service-level booleans
		if servicesList, ok := secAttrs["services"].(types.List); ok && !servicesList.IsNull() && !servicesList.IsUnknown() {
			newAttrs["services"] = applyServicesWriteOnlyBooleans(servicesList, serviceMap, ServiceAttrTypes(), diags)
		}

		newSection, newDiags := types.ObjectValue(SectionAttrTypes(), newAttrs)
		diags.Append(newDiags...)
		newSections[i] = newSection
	}

	newList, newDiags := types.ListValue(apiSections.ElementType(context.Background()), newSections)
	diags.Append(newDiags...)

	return newList
}

// applyServicesWriteOnlyBooleans applies write-only boolean preservation to a services list.
func applyServicesWriteOnlyBooleans(servicesList types.List, serviceMap map[string]writeOnlyBooleans, attrTypes map[string]attr.Type, diags *diag.Diagnostics) types.List {
	elements := servicesList.Elements()
	newServices := make([]attr.Value, len(elements))

	for i, svcElem := range elements {
		svcObj, ok := svcElem.(types.Object)
		if !ok {
			newServices[i] = svcElem
			continue
		}

		svcAttrs := svcObj.Attributes()
		uuid := ""
		if uuidAttr, ok := svcAttrs["uuid"].(types.String); ok && !uuidAttr.IsNull() {
			uuid = uuidAttr.ValueString()
		}

		newAttrs := copyAttrs(svcAttrs)
		modified := false

		// Apply boolean overrides if we have plan values for this UUID.
		// The API ignores these booleans on write and always returns its own
		// value, so we trust the plan value whenever it differs from the API.
		if uuid != "" {
			if wb, ok := serviceMap[uuid]; ok {
				if wb.showUptime != nil {
					if apiVal, ok := svcAttrs["show_uptime"].(types.Bool); ok && !apiVal.IsNull() && apiVal.ValueBool() != *wb.showUptime {
						newAttrs["show_uptime"] = types.BoolValue(*wb.showUptime)
						modified = true
					}
				}
				if wb.showResponseTimes != nil {
					if apiVal, ok := svcAttrs["show_response_times"].(types.Bool); ok && !apiVal.IsNull() && apiVal.ValueBool() != *wb.showResponseTimes {
						newAttrs["show_response_times"] = types.BoolValue(*wb.showResponseTimes)
						modified = true
					}
				}
			}
		}

		// Recurse into nested services for groups
		if nestedList, ok := svcAttrs["services"].(types.List); ok && !nestedList.IsNull() && !nestedList.IsUnknown() {
			preservedNested := applyServicesWriteOnlyBooleans(nestedList, serviceMap, NestedServiceAttrTypes(), diags)
			newAttrs["services"] = preservedNested
			modified = true
		}

		if modified {
			newSvc, newDiags := types.ObjectValue(attrTypes, newAttrs)
			diags.Append(newDiags...)
			newServices[i] = newSvc
		} else {
			newServices[i] = svcElem
		}
	}

	newList, newDiags := types.ListValue(servicesList.ElementType(context.Background()), newServices)
	diags.Append(newDiags...)

	return newList
}

// copyAttrs creates a shallow copy of an attribute map.
func copyAttrs(attrs map[string]attr.Value) map[string]attr.Value {
	newAttrs := make(map[string]attr.Value, len(attrs))
	for k, v := range attrs {
		newAttrs[k] = v
	}
	return newAttrs
}
