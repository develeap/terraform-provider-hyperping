// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// buildTestService builds a TF service object for testing.
func buildTestService(uuid string, showUptime, showResponseTimes bool, isGroup bool, nested types.List) types.Object {
	attrs := map[string]attr.Value{
		"id":                  types.StringValue(""),
		"uuid":                types.StringValue(uuid),
		"name":                types.MapNull(types.StringType),
		"is_group":            types.BoolValue(isGroup),
		"show_uptime":         types.BoolValue(showUptime),
		"show_response_times": types.BoolValue(showResponseTimes),
		"services":            nested,
	}
	obj, _ := types.ObjectValue(ServiceAttrTypes(), attrs)
	return obj
}

// buildTestNestedService builds a TF nested service object for testing.
func buildTestNestedService(uuid string, showUptime, showResponseTimes bool) types.Object {
	attrs := map[string]attr.Value{
		"id":                  types.StringValue(""),
		"uuid":                types.StringValue(uuid),
		"name":                types.MapNull(types.StringType),
		"is_group":            types.BoolValue(false),
		"show_uptime":         types.BoolValue(showUptime),
		"show_response_times": types.BoolValue(showResponseTimes),
	}
	obj, _ := types.ObjectValue(NestedServiceAttrTypes(), attrs)
	return obj
}

// buildTestSection builds a TF section object for testing.
func buildTestSection(isSplit bool, services types.List) types.Object {
	attrs := map[string]attr.Value{
		"name":     types.MapNull(types.StringType),
		"is_split": types.BoolValue(isSplit),
		"services": services,
	}
	obj, _ := types.ObjectValue(SectionAttrTypes(), attrs)
	return obj
}

// buildServicesList builds a types.List of services from objects.
func buildServicesList(t *testing.T, services ...types.Object) types.List {
	t.Helper()

	values := make([]attr.Value, len(services))
	for i, s := range services {
		values[i] = s
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: ServiceAttrTypes()}, values)
	if diags.HasError() {
		t.Fatalf("failed to build services list: %v", diags)
	}
	return list
}

// buildNestedServicesList builds a types.List of nested services from objects.
func buildNestedServicesList(t *testing.T, services ...types.Object) types.List {
	t.Helper()

	values := make([]attr.Value, len(services))
	for i, s := range services {
		values[i] = s
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}, values)
	if diags.HasError() {
		t.Fatalf("failed to build nested services list: %v", diags)
	}
	return list
}

// buildSectionsList builds a types.List of sections from objects.
func buildSectionsList(t *testing.T, sections ...types.Object) types.List {
	t.Helper()

	values := make([]attr.Value, len(sections))
	for i, s := range sections {
		values[i] = s
	}
	list, diags := types.ListValue(types.ObjectType{AttrTypes: SectionAttrTypes()}, values)
	if diags.HasError() {
		t.Fatalf("failed to build sections list: %v", diags)
	}
	return list
}

func TestExtractWriteOnlyBooleans_BasicService(t *testing.T) {
	svc := buildTestService("mon_abc123", true, true, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	svcList := buildServicesList(t, svc)
	section := buildTestSection(true, svcList)
	sections := buildSectionsList(t, section)

	serviceMap, sectionIsSplit := extractWriteOnlyBooleans(sections)

	// Check service booleans
	wb, ok := serviceMap["mon_abc123"]
	if !ok {
		t.Fatal("expected mon_abc123 in serviceMap")
	}
	if wb.showUptime == nil || !*wb.showUptime {
		t.Error("expected showUptime=true")
	}
	if wb.showResponseTimes == nil || !*wb.showResponseTimes {
		t.Error("expected showResponseTimes=true")
	}

	// Check section is_split
	if !sectionIsSplit[0] {
		t.Error("expected sectionIsSplit[0]=true")
	}
}

func TestExtractWriteOnlyBooleans_NestedService(t *testing.T) {
	nested := buildTestNestedService("mon_nested1", true, true)
	nestedList := buildNestedServicesList(nested)
	group := buildTestService("", false, false, true, nestedList)
	svcList := buildServicesList(group)
	section := buildTestSection(false, svcList)
	sections := buildSectionsList(section)

	serviceMap, _ := extractWriteOnlyBooleans(sections)

	wb, ok := serviceMap["mon_nested1"]
	if !ok {
		t.Fatal("expected mon_nested1 in serviceMap from nested services")
	}
	if wb.showUptime == nil || !*wb.showUptime {
		t.Error("expected showUptime=true for nested service")
	}
}

func TestExtractWriteOnlyBooleans_NullSections(t *testing.T) {
	nullSections := types.ListNull(types.ObjectType{AttrTypes: SectionAttrTypes()})
	serviceMap, sectionIsSplit := extractWriteOnlyBooleans(nullSections)

	if len(serviceMap) != 0 {
		t.Errorf("expected empty serviceMap, got %d entries", len(serviceMap))
	}
	if len(sectionIsSplit) != 0 {
		t.Errorf("expected empty sectionIsSplit, got %d entries", len(sectionIsSplit))
	}
}

func TestApplyWriteOnlyBooleans_PlanTrueApiFalse(t *testing.T) {
	r := &StatusPageResource{}
	var diags diag.Diagnostics

	// API returns false for booleans
	apiSvc := buildTestService("mon_abc123", false, false, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	apiSvcList := buildServicesList(apiSvc)
	apiSection := buildTestSection(false, apiSvcList)
	apiSections := buildSectionsList(apiSection)

	// Plan had true
	serviceMap := map[string]writeOnlyBooleans{
		"mon_abc123": {
			showUptime:        boolPtr(true),
			showResponseTimes: boolPtr(true),
		},
	}
	sectionIsSplit := map[int]bool{0: true}

	result := r.applyWriteOnlyBooleans(apiSections, serviceMap, sectionIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	// Check that booleans were preserved
	secObj := result.Elements()[0].(types.Object)
	secAttrs := secObj.Attributes()

	// is_split should be overridden to true
	isSplit := secAttrs["is_split"].(types.Bool)
	if !isSplit.ValueBool() {
		t.Error("expected is_split=true after preservation")
	}

	// Check service booleans
	svcList := secAttrs["services"].(types.List)
	svcObj := svcList.Elements()[0].(types.Object)
	svcAttrs := svcObj.Attributes()

	showUptime := svcAttrs["show_uptime"].(types.Bool)
	if !showUptime.ValueBool() {
		t.Error("expected show_uptime=true after preservation")
	}

	showRT := svcAttrs["show_response_times"].(types.Bool)
	if !showRT.ValueBool() {
		t.Error("expected show_response_times=true after preservation")
	}
}

func TestApplyWriteOnlyBooleans_UUIDBasedReordering(t *testing.T) {
	r := &StatusPageResource{}
	var diags diag.Diagnostics

	// API returns services in different order than plan
	apiSvc1 := buildTestService("mon_def456", false, false, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	apiSvc2 := buildTestService("mon_abc123", false, false, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	apiSvcList := buildServicesList(apiSvc1, apiSvc2)
	apiSection := buildTestSection(false, apiSvcList)
	apiSections := buildSectionsList(apiSection)

	// Plan had mon_abc123=true, mon_def456=false
	serviceMap := map[string]writeOnlyBooleans{
		"mon_abc123": {showUptime: boolPtr(true), showResponseTimes: boolPtr(true)},
		"mon_def456": {showUptime: boolPtr(false), showResponseTimes: boolPtr(false)},
	}
	sectionIsSplit := map[int]bool{}

	result := r.applyWriteOnlyBooleans(apiSections, serviceMap, sectionIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	secObj := result.Elements()[0].(types.Object)
	svcList := secObj.Attributes()["services"].(types.List)

	// First service (mon_def456) should NOT be overridden (plan=false)
	svc1Attrs := svcList.Elements()[0].(types.Object).Attributes()
	if svc1Attrs["show_uptime"].(types.Bool).ValueBool() {
		t.Error("mon_def456 show_uptime should remain false")
	}

	// Second service (mon_abc123) should be overridden (plan=true)
	svc2Attrs := svcList.Elements()[1].(types.Object).Attributes()
	if !svc2Attrs["show_uptime"].(types.Bool).ValueBool() {
		t.Error("mon_abc123 show_uptime should be true after preservation")
	}
}

func TestApplyWriteOnlyBooleans_NestedServices(t *testing.T) {
	r := &StatusPageResource{}
	var diags diag.Diagnostics

	// API returns nested service with false
	nestedSvc := buildTestNestedService("mon_nested1", false, false)
	nestedList := buildNestedServicesList(nestedSvc)
	groupSvc := buildTestService("", false, false, true, nestedList)
	apiSvcList := buildServicesList(groupSvc)
	apiSection := buildTestSection(false, apiSvcList)
	apiSections := buildSectionsList(apiSection)

	// Plan had true for nested service
	serviceMap := map[string]writeOnlyBooleans{
		"mon_nested1": {showUptime: boolPtr(true), showResponseTimes: boolPtr(true)},
	}
	sectionIsSplit := map[int]bool{}

	result := r.applyWriteOnlyBooleans(apiSections, serviceMap, sectionIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	secObj := result.Elements()[0].(types.Object)
	svcList := secObj.Attributes()["services"].(types.List)
	groupObj := svcList.Elements()[0].(types.Object)
	nestedSvcList := groupObj.Attributes()["services"].(types.List)
	nestedAttrs := nestedSvcList.Elements()[0].(types.Object).Attributes()

	if !nestedAttrs["show_uptime"].(types.Bool).ValueBool() {
		t.Error("nested service show_uptime should be true after preservation")
	}
	if !nestedAttrs["show_response_times"].(types.Bool).ValueBool() {
		t.Error("nested service show_response_times should be true after preservation")
	}
}

func TestApplyWriteOnlyBooleans_LengthMismatch(t *testing.T) {
	r := &StatusPageResource{}
	var diags diag.Diagnostics

	// API returns 2 services, plan had 3 — UUID matching should still work
	apiSvc1 := buildTestService("mon_abc123", false, false, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	apiSvc2 := buildTestService("mon_def456", false, false, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	apiSvcList := buildServicesList(apiSvc1, apiSvc2)
	apiSection := buildTestSection(false, apiSvcList)
	apiSections := buildSectionsList(apiSection)

	// Plan had 3 services including mon_ghi789 which API didn't return
	serviceMap := map[string]writeOnlyBooleans{
		"mon_abc123": {showUptime: boolPtr(true)},
		"mon_def456": {showUptime: boolPtr(true)},
		"mon_ghi789": {showUptime: boolPtr(true)},
	}
	sectionIsSplit := map[int]bool{}

	result := r.applyWriteOnlyBooleans(apiSections, serviceMap, sectionIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	secObj := result.Elements()[0].(types.Object)
	svcList := secObj.Attributes()["services"].(types.List)

	// Both existing services should be preserved
	svc1Attrs := svcList.Elements()[0].(types.Object).Attributes()
	if !svc1Attrs["show_uptime"].(types.Bool).ValueBool() {
		t.Error("mon_abc123 show_uptime should be true")
	}

	svc2Attrs := svcList.Elements()[1].(types.Object).Attributes()
	if !svc2Attrs["show_uptime"].(types.Bool).ValueBool() {
		t.Error("mon_def456 show_uptime should be true")
	}
}

func TestApplyWriteOnlyBooleans_EmptyServiceMap(t *testing.T) {
	r := &StatusPageResource{}
	var diags diag.Diagnostics

	apiSvc := buildTestService("mon_abc123", false, false, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	apiSvcList := buildServicesList(apiSvc)
	apiSection := buildTestSection(false, apiSvcList)
	apiSections := buildSectionsList(apiSection)

	// Empty maps (import scenario)
	result := r.applyWriteOnlyBooleans(apiSections, map[string]writeOnlyBooleans{}, map[int]bool{}, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	// Values should remain as API returned them
	secObj := result.Elements()[0].(types.Object)
	svcList := secObj.Attributes()["services"].(types.List)
	svcAttrs := svcList.Elements()[0].(types.Object).Attributes()

	if svcAttrs["show_uptime"].(types.Bool).ValueBool() {
		t.Error("show_uptime should remain false with empty serviceMap")
	}
}

func TestApplyWriteOnlyBooleans_MixedPreservation(t *testing.T) {
	r := &StatusPageResource{}
	var diags diag.Diagnostics

	// API returns false for all
	apiSvc1 := buildTestService("mon_abc123", false, false, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	apiSvc2 := buildTestService("mon_def456", false, false, false, types.ListNull(types.ObjectType{AttrTypes: NestedServiceAttrTypes()}))
	apiSvcList := buildServicesList(apiSvc1, apiSvc2)
	apiSection := buildTestSection(false, apiSvcList)
	apiSections := buildSectionsList(apiSection)

	// Plan: mon_abc123=true, mon_def456=false
	serviceMap := map[string]writeOnlyBooleans{
		"mon_abc123": {showUptime: boolPtr(true), showResponseTimes: boolPtr(false)},
		"mon_def456": {showUptime: boolPtr(false), showResponseTimes: boolPtr(false)},
	}
	sectionIsSplit := map[int]bool{}

	result := r.applyWriteOnlyBooleans(apiSections, serviceMap, sectionIsSplit, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", diags.Errors())
	}

	secObj := result.Elements()[0].(types.Object)
	svcList := secObj.Attributes()["services"].(types.List)

	// mon_abc123: showUptime preserved (plan=true), showResponseTimes NOT preserved (plan=false)
	svc1Attrs := svcList.Elements()[0].(types.Object).Attributes()
	if !svc1Attrs["show_uptime"].(types.Bool).ValueBool() {
		t.Error("mon_abc123 show_uptime should be true")
	}
	if svc1Attrs["show_response_times"].(types.Bool).ValueBool() {
		t.Error("mon_abc123 show_response_times should remain false (plan=false)")
	}

	// mon_def456: both should remain false
	svc2Attrs := svcList.Elements()[1].(types.Object).Attributes()
	if svc2Attrs["show_uptime"].(types.Bool).ValueBool() {
		t.Error("mon_def456 show_uptime should remain false")
	}
}

func boolPtr(v bool) *bool {
	return &v
}
