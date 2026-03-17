// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- test helpers ---

func testSection(isSplit bool) types.Object {
	attrs := map[string]attr.Value{
		"name":     types.MapNull(types.StringType),
		"is_split": types.BoolValue(isSplit),
		"services": types.ListNull(types.ObjectType{AttrTypes: ServiceAttrTypes()}),
	}
	obj, _ := types.ObjectValue(SectionAttrTypes(), attrs)
	return obj
}

func testSectionsList(sections ...types.Object) types.List {
	values := make([]attr.Value, len(sections))
	for i, s := range sections {
		values[i] = s
	}
	list, _ := types.ListValue(types.ObjectType{AttrTypes: SectionAttrTypes()}, values)
	return list
}

// --- extractSectionIsSplit tests ---

func TestExtractSectionIsSplit_Basic(t *testing.T) {
	sections := testSectionsList(testSection(true), testSection(false))

	result := extractSectionIsSplit(sections)

	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if !result[0] {
		t.Error("expected index 0 = true")
	}
	if result[1] {
		t.Error("expected index 1 = false")
	}
}

func TestExtractSectionIsSplit_NullSections(t *testing.T) {
	nullSections := types.ListNull(types.ObjectType{AttrTypes: SectionAttrTypes()})

	result := extractSectionIsSplit(nullSections)

	if len(result) != 0 {
		t.Errorf("expected empty map for null sections, got %d entries", len(result))
	}
}

func TestExtractSectionIsSplit_UnknownSections(t *testing.T) {
	unknownSections := types.ListUnknown(types.ObjectType{AttrTypes: SectionAttrTypes()})

	result := extractSectionIsSplit(unknownSections)

	if len(result) != 0 {
		t.Errorf("expected empty map for unknown sections, got %d entries", len(result))
	}
}

// --- preserveSectionIsSplit tests ---

func TestPreserveSectionIsSplit_PlanTrueApiFalse(t *testing.T) {
	var diags diag.Diagnostics
	apiSections := testSectionsList(testSection(false))
	planIsSplit := map[int]bool{0: true}

	result := preserveSectionIsSplit(apiSections, planIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	secObj := result.Elements()[0].(types.Object)
	isSplit := secObj.Attributes()["is_split"].(types.Bool)
	if !isSplit.ValueBool() {
		t.Error("expected is_split=true (plan should override API false)")
	}
}

func TestPreserveSectionIsSplit_PlanFalseApiTrue(t *testing.T) {
	var diags diag.Diagnostics
	apiSections := testSectionsList(testSection(true))
	planIsSplit := map[int]bool{0: false}

	result := preserveSectionIsSplit(apiSections, planIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	secObj := result.Elements()[0].(types.Object)
	isSplit := secObj.Attributes()["is_split"].(types.Bool)
	if isSplit.ValueBool() {
		t.Error("expected is_split=false (plan should override API true)")
	}
}

func TestPreserveSectionIsSplit_NullSectionsPassthrough(t *testing.T) {
	var diags diag.Diagnostics
	nullSections := types.ListNull(types.ObjectType{AttrTypes: SectionAttrTypes()})
	planIsSplit := map[int]bool{0: true}

	result := preserveSectionIsSplit(nullSections, planIsSplit, &diags)

	if !result.IsNull() {
		t.Error("expected null list to pass through unchanged")
	}
}

func TestPreserveSectionIsSplit_EmptyPlanMapPassthrough(t *testing.T) {
	var diags diag.Diagnostics
	apiSections := testSectionsList(testSection(false))

	result := preserveSectionIsSplit(apiSections, map[int]bool{}, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	secObj := result.Elements()[0].(types.Object)
	isSplit := secObj.Attributes()["is_split"].(types.Bool)
	if isSplit.ValueBool() {
		t.Error("expected is_split to remain false with empty plan map")
	}
}

func TestPreserveSectionIsSplit_SectionCountMismatch(t *testing.T) {
	var diags diag.Diagnostics
	// API returns 2 sections, plan had 3 — index 2 has no corresponding API section
	apiSections := testSectionsList(testSection(false), testSection(false))
	planIsSplit := map[int]bool{0: true, 1: false, 2: true}

	result := preserveSectionIsSplit(apiSections, planIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	elements := result.Elements()
	if len(elements) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(elements))
	}

	// Index 0: plan=true, API=false -> should be overridden to true
	sec0 := elements[0].(types.Object).Attributes()["is_split"].(types.Bool)
	if !sec0.ValueBool() {
		t.Error("section 0: expected is_split=true")
	}

	// Index 1: plan=false, API=false -> no change needed
	sec1 := elements[1].(types.Object).Attributes()["is_split"].(types.Bool)
	if sec1.ValueBool() {
		t.Error("section 1: expected is_split=false")
	}
}

func TestPreserveSectionIsSplit_UntrackedSectionUntouched(t *testing.T) {
	var diags diag.Diagnostics
	// API returns 2 sections, plan only tracks index 1
	apiSections := testSectionsList(testSection(false), testSection(false))
	planIsSplit := map[int]bool{1: true}

	result := preserveSectionIsSplit(apiSections, planIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	// Index 0: not tracked, should remain false
	sec0 := result.Elements()[0].(types.Object).Attributes()["is_split"].(types.Bool)
	if sec0.ValueBool() {
		t.Error("section 0: expected is_split=false (untracked)")
	}

	// Index 1: tracked, plan=true -> overridden
	sec1 := result.Elements()[1].(types.Object).Attributes()["is_split"].(types.Bool)
	if !sec1.ValueBool() {
		t.Error("section 1: expected is_split=true")
	}
}

func TestPreserveSectionIsSplit_NoOpWhenValuesMatch(t *testing.T) {
	var diags diag.Diagnostics
	apiSections := testSectionsList(testSection(true))
	planIsSplit := map[int]bool{0: true}

	result := preserveSectionIsSplit(apiSections, planIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	// Should pass through without modification when values already match
	secObj := result.Elements()[0].(types.Object)
	isSplit := secObj.Attributes()["is_split"].(types.Bool)
	if !isSplit.ValueBool() {
		t.Error("expected is_split=true (values matched, should be unchanged)")
	}
}
