// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// extractSectionIsSplit captures is_split values from the plan/state sections
// before the API response overwrites them. The Hyperping API accepts is_split
// on write but never returns it on read (always defaults to false).
func extractSectionIsSplit(sections types.List) map[int]bool {
	result := make(map[int]bool)

	if sections.IsNull() || sections.IsUnknown() {
		return result
	}

	for i, elem := range sections.Elements() {
		secObj, ok := elem.(types.Object)
		if !ok {
			continue
		}
		if isSplit, ok := secObj.Attributes()["is_split"].(types.Bool); ok && !isSplit.IsNull() {
			result[i] = isSplit.ValueBool()
		}
	}

	return result
}

// preserveSectionIsSplit restores is_split values from the plan onto the
// API response sections. Without this, Terraform reports "inconsistent result
// after apply" because the API always returns is_split as false.
func preserveSectionIsSplit(apiSections types.List, planIsSplit map[int]bool, diags *diag.Diagnostics) types.List {
	if apiSections.IsNull() || apiSections.IsUnknown() || len(planIsSplit) == 0 {
		return apiSections
	}

	elements := apiSections.Elements()
	newSections := make([]attr.Value, len(elements))

	for i, elem := range elements {
		secObj, ok := elem.(types.Object)
		if !ok {
			newSections[i] = elem
			continue
		}

		planVal, tracked := planIsSplit[i]
		if !tracked {
			newSections[i] = elem
			continue
		}

		apiIsSplit, ok := secObj.Attributes()["is_split"].(types.Bool)
		if !ok || (!apiIsSplit.IsNull() && apiIsSplit.ValueBool() == planVal) {
			newSections[i] = elem
			continue
		}

		// Override with plan value
		newAttrs := make(map[string]attr.Value, len(secObj.Attributes()))
		for k, v := range secObj.Attributes() {
			newAttrs[k] = v
		}
		newAttrs["is_split"] = types.BoolValue(planVal)

		newSection, newDiags := types.ObjectValue(SectionAttrTypes(), newAttrs)
		diags.Append(newDiags...)
		newSections[i] = newSection
	}

	newList, newDiags := types.ListValue(apiSections.ElementType(context.Background()), newSections)
	diags.Append(newDiags...)

	return newList
}
