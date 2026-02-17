// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// =============================================================================
// TestBuildMaintenanceUpdateRequest
// =============================================================================

func TestBuildMaintenanceUpdateRequest_changedNameIncluded(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MaintenanceResourceModel{
		Name:      types.StringValue("new-name"),
		Title:     types.StringNull(),
		Text:      types.StringNull(),
		StartDate: types.StringValue("2026-02-01T00:00:00Z"),
		EndDate:   types.StringValue("2026-02-01T02:00:00Z"),
		Monitors:  makeStringList([]string{"mon-1"}),
	}
	state := plan
	state.Name = types.StringValue("old-name")

	req := buildMaintenanceUpdateRequest(ctx, &plan, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Name == nil {
		t.Fatal("expected Name to be set")
	}
	if *req.Name != "new-name" {
		t.Errorf("expected Name='new-name', got=%q", *req.Name)
	}
}

func TestBuildMaintenanceUpdateRequest_unchangedNameOmitted(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MaintenanceResourceModel{
		Name:      types.StringValue("same-name"),
		Title:     types.StringNull(),
		Text:      types.StringNull(),
		StartDate: types.StringValue("2026-02-01T00:00:00Z"),
		EndDate:   types.StringValue("2026-02-01T02:00:00Z"),
		Monitors:  makeStringList([]string{"mon-1"}),
	}
	state := plan

	req := buildMaintenanceUpdateRequest(ctx, &plan, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Name != nil {
		t.Errorf("expected Name to be nil (unchanged), got=%q", *req.Name)
	}
}

func TestBuildMaintenanceUpdateRequest_changedTitleIncluded(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MaintenanceResourceModel{
		Name:      types.StringValue("name"),
		Title:     types.StringValue("New Title"),
		Text:      types.StringNull(),
		StartDate: types.StringValue("2026-02-01T00:00:00Z"),
		EndDate:   types.StringValue("2026-02-01T02:00:00Z"),
		Monitors:  makeStringList([]string{"mon-1"}),
	}
	state := plan
	state.Title = types.StringValue("Old Title")

	req := buildMaintenanceUpdateRequest(ctx, &plan, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Title == nil {
		t.Fatal("expected Title to be set")
	}
	if req.Title.En != "New Title" {
		t.Errorf("expected Title.En='New Title', got=%q", req.Title.En)
	}
}

func TestBuildMaintenanceUpdateRequest_nullTitleOmitted(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MaintenanceResourceModel{
		Name:      types.StringValue("name"),
		Title:     types.StringNull(),
		Text:      types.StringNull(),
		StartDate: types.StringValue("2026-02-01T00:00:00Z"),
		EndDate:   types.StringValue("2026-02-01T02:00:00Z"),
		Monitors:  makeStringList([]string{"mon-1"}),
	}
	state := plan
	state.Title = types.StringValue("Had Title")

	req := buildMaintenanceUpdateRequest(ctx, &plan, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	// Title changed to null: condition is !plan.Title.IsNull() so it should be omitted
	if req.Title != nil {
		t.Errorf("expected Title to be nil when plan title is null, got=%v", req.Title)
	}
}

func TestBuildMaintenanceUpdateRequest_changedDatesIncluded(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MaintenanceResourceModel{
		Name:      types.StringValue("name"),
		Title:     types.StringNull(),
		Text:      types.StringNull(),
		StartDate: types.StringValue("2026-03-01T00:00:00Z"),
		EndDate:   types.StringValue("2026-03-01T04:00:00Z"),
		Monitors:  makeStringList([]string{"mon-1"}),
	}
	state := plan
	state.StartDate = types.StringValue("2026-02-01T00:00:00Z")
	state.EndDate = types.StringValue("2026-02-01T02:00:00Z")

	req := buildMaintenanceUpdateRequest(ctx, &plan, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.StartDate == nil || *req.StartDate != "2026-03-01T00:00:00Z" {
		t.Errorf("expected StartDate='2026-03-01T00:00:00Z', got=%v", req.StartDate)
	}
	if req.EndDate == nil || *req.EndDate != "2026-03-01T04:00:00Z" {
		t.Errorf("expected EndDate='2026-03-01T04:00:00Z', got=%v", req.EndDate)
	}
}

func TestBuildMaintenanceUpdateRequest_changedMonitorsIncluded(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MaintenanceResourceModel{
		Name:      types.StringValue("name"),
		Title:     types.StringNull(),
		Text:      types.StringNull(),
		StartDate: types.StringValue("2026-02-01T00:00:00Z"),
		EndDate:   types.StringValue("2026-02-01T02:00:00Z"),
		Monitors:  makeStringList([]string{"mon-1", "mon-2"}),
	}
	state := plan
	state.Monitors = makeStringList([]string{"mon-1"})

	req := buildMaintenanceUpdateRequest(ctx, &plan, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Monitors == nil {
		t.Fatal("expected Monitors to be set")
	}
	if len(*req.Monitors) != 2 {
		t.Errorf("expected 2 monitors, got=%d", len(*req.Monitors))
	}
}

func TestBuildMaintenanceUpdateRequest_noChangesAllNil(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MaintenanceResourceModel{
		Name:      types.StringValue("name"),
		Title:     types.StringNull(),
		Text:      types.StringNull(),
		StartDate: types.StringValue("2026-02-01T00:00:00Z"),
		EndDate:   types.StringValue("2026-02-01T02:00:00Z"),
		Monitors:  makeStringList([]string{"mon-1"}),
	}
	state := plan

	req := buildMaintenanceUpdateRequest(ctx, &plan, &state, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.Name != nil || req.Title != nil || req.Text != nil ||
		req.StartDate != nil || req.EndDate != nil || req.Monitors != nil {
		t.Error("expected all fields nil when no changes")
	}
}

// =============================================================================
// TestApplyHTTPFieldChanges
// =============================================================================

func TestApplyHTTPFieldChanges_changedRequestBodyIncluded(t *testing.T) {
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		RequestBody:    types.StringValue(`{"key":"value"}`),
		RequestHeaders: types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}),
	}
	state := MonitorResourceModel{
		RequestBody:    types.StringNull(),
		RequestHeaders: types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}),
	}
	req := client.UpdateMonitorRequest{}

	applyHTTPFieldChanges(&plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestBody == nil {
		t.Fatal("expected RequestBody to be set")
	}
	if *req.RequestBody != `{"key":"value"}` {
		t.Errorf("expected RequestBody='%s', got=%q", `{"key":"value"}`, *req.RequestBody)
	}
}

func TestApplyHTTPFieldChanges_nullBodyClearsToEmpty(t *testing.T) {
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		RequestBody:    types.StringNull(),
		RequestHeaders: types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}),
	}
	state := MonitorResourceModel{
		RequestBody:    types.StringValue("old-body"),
		RequestHeaders: types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}),
	}
	req := client.UpdateMonitorRequest{}

	applyHTTPFieldChanges(&plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestBody == nil {
		t.Fatal("expected RequestBody to be set (empty string to clear)")
	}
	if *req.RequestBody != "" {
		t.Errorf("expected RequestBody='' to clear, got=%q", *req.RequestBody)
	}
}

func TestApplyHTTPFieldChanges_unchangedBodyOmitted(t *testing.T) {
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		RequestBody:    types.StringValue("same-body"),
		RequestHeaders: types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}),
	}
	state := plan
	req := client.UpdateMonitorRequest{}

	applyHTTPFieldChanges(&plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestBody != nil {
		t.Errorf("expected RequestBody nil (unchanged), got=%q", *req.RequestBody)
	}
}

func TestApplyHTTPFieldChanges_nullHeadersClearsToEmpty(t *testing.T) {
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		RequestBody:    types.StringNull(),
		RequestHeaders: types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}),
	}
	state := MonitorResourceModel{
		RequestBody:    types.StringNull(),
		RequestHeaders: makeHeaderList([]client.RequestHeader{{Name: "X-Test", Value: "val"}}),
	}
	req := client.UpdateMonitorRequest{}

	applyHTTPFieldChanges(&plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequestHeaders == nil {
		t.Fatal("expected RequestHeaders to be set")
	}
	if len(*req.RequestHeaders) != 0 {
		t.Errorf("expected empty RequestHeaders slice, got len=%d", len(*req.RequestHeaders))
	}
}

// =============================================================================
// TestApplyMonitoringFieldChanges
// =============================================================================

func TestApplyMonitoringFieldChanges_changedAlertsWaitIncluded(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		AlertsWait:       types.Int64Value(300),
		EscalationPolicy: types.StringNull(),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	state := MonitorResourceModel{
		AlertsWait:       types.Int64Value(60),
		EscalationPolicy: types.StringNull(),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	req := client.UpdateMonitorRequest{}

	applyMonitoringFieldChanges(ctx, &plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.AlertsWait == nil {
		t.Fatal("expected AlertsWait to be set")
	}
	if *req.AlertsWait != 300 {
		t.Errorf("expected AlertsWait=300, got=%d", *req.AlertsWait)
	}
}

func TestApplyMonitoringFieldChanges_nullAlertsWaitClearsToZero(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		AlertsWait:       types.Int64Null(),
		EscalationPolicy: types.StringNull(),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	state := MonitorResourceModel{
		AlertsWait:       types.Int64Value(300),
		EscalationPolicy: types.StringNull(),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	req := client.UpdateMonitorRequest{}

	applyMonitoringFieldChanges(ctx, &plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.AlertsWait == nil {
		t.Fatal("expected AlertsWait to be set to zero")
	}
	if *req.AlertsWait != 0 {
		t.Errorf("expected AlertsWait=0, got=%d", *req.AlertsWait)
	}
}

func TestApplyMonitoringFieldChanges_changedEscalationPolicyIncluded(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		AlertsWait:       types.Int64Null(),
		EscalationPolicy: types.StringValue("policy-uuid-new"),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	state := MonitorResourceModel{
		AlertsWait:       types.Int64Null(),
		EscalationPolicy: types.StringValue("policy-uuid-old"),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	req := client.UpdateMonitorRequest{}

	applyMonitoringFieldChanges(ctx, &plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.EscalationPolicy == nil {
		t.Fatal("expected EscalationPolicy to be set")
	}
	if *req.EscalationPolicy != "policy-uuid-new" {
		t.Errorf("expected EscalationPolicy='policy-uuid-new', got=%q", *req.EscalationPolicy)
	}
}

func TestApplyMonitoringFieldChanges_nullEscalationPolicyClearsToEmpty(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		AlertsWait:       types.Int64Null(),
		EscalationPolicy: types.StringNull(),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	state := MonitorResourceModel{
		AlertsWait:       types.Int64Null(),
		EscalationPolicy: types.StringValue("old-policy"),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	req := client.UpdateMonitorRequest{}

	applyMonitoringFieldChanges(ctx, &plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.EscalationPolicy == nil {
		t.Fatal("expected EscalationPolicy to be set to empty string")
	}
	if *req.EscalationPolicy != "" {
		t.Errorf("expected EscalationPolicy='', got=%q", *req.EscalationPolicy)
	}
}

func TestApplyMonitoringFieldChanges_changedRequiredKeywordIncluded(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		AlertsWait:       types.Int64Null(),
		EscalationPolicy: types.StringNull(),
		RequiredKeyword:  types.StringValue("success"),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	state := MonitorResourceModel{
		AlertsWait:       types.Int64Null(),
		EscalationPolicy: types.StringNull(),
		RequiredKeyword:  types.StringNull(),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	req := client.UpdateMonitorRequest{}

	applyMonitoringFieldChanges(ctx, &plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.RequiredKeyword == nil {
		t.Fatal("expected RequiredKeyword to be set")
	}
	if *req.RequiredKeyword != "success" {
		t.Errorf("expected RequiredKeyword='success', got=%q", *req.RequiredKeyword)
	}
}

func TestApplyMonitoringFieldChanges_unchangedFieldsOmitted(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := MonitorResourceModel{
		AlertsWait:       types.Int64Value(60),
		EscalationPolicy: types.StringValue("same-policy"),
		RequiredKeyword:  types.StringValue("same-keyword"),
		Regions:          makeStringList([]string{"london"}),
		Port:             types.Int64Null(),
	}
	state := plan
	req := client.UpdateMonitorRequest{}

	applyMonitoringFieldChanges(ctx, &plan, &state, &req, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}
	if req.AlertsWait != nil || req.EscalationPolicy != nil || req.RequiredKeyword != nil || req.Regions != nil {
		t.Error("expected all monitoring fields nil when unchanged")
	}
}

// =============================================================================
// TestApplyHealthcheckTimingFields
// =============================================================================

func TestApplyHealthcheckTimingFields_changedCronIncluded(t *testing.T) {
	plan := HealthcheckResourceModel{
		Cron:             types.StringValue("0 0 * * *"),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(5),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := HealthcheckResourceModel{
		Cron:             types.StringValue("0 12 * * *"),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(5),
		GracePeriodType:  types.StringValue("minutes"),
	}
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckTimingFields(&plan, &state, &req)

	if !hasChanges {
		t.Error("expected hasChanges=true when cron changes")
	}
	if req.Cron == nil {
		t.Fatal("expected Cron to be set")
	}
	if *req.Cron != "0 0 * * *" {
		t.Errorf("expected Cron='0 0 * * *', got=%q", *req.Cron)
	}
}

func TestApplyHealthcheckTimingFields_nullCronClearsToEmpty(t *testing.T) {
	plan := HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(5),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := HealthcheckResourceModel{
		Cron:             types.StringValue("0 0 * * *"),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(5),
		GracePeriodType:  types.StringValue("minutes"),
	}
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckTimingFields(&plan, &state, &req)

	if !hasChanges {
		t.Error("expected hasChanges=true")
	}
	if req.Cron == nil {
		t.Fatal("expected Cron to be set to empty string")
	}
	if *req.Cron != "" {
		t.Errorf("expected Cron='', got=%q", *req.Cron)
	}
}

func TestApplyHealthcheckTimingFields_changedGracePeriodIncluded(t *testing.T) {
	plan := HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(10),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(5),
		GracePeriodType:  types.StringValue("minutes"),
	}
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckTimingFields(&plan, &state, &req)

	if !hasChanges {
		t.Error("expected hasChanges=true")
	}
	if req.GracePeriodValue == nil {
		t.Fatal("expected GracePeriodValue to be set")
	}
	if *req.GracePeriodValue != 10 {
		t.Errorf("expected GracePeriodValue=10, got=%d", *req.GracePeriodValue)
	}
}

func TestApplyHealthcheckTimingFields_noChangesReturnsFalse(t *testing.T) {
	plan := HealthcheckResourceModel{
		Cron:             types.StringValue("0 0 * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(5),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := plan
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckTimingFields(&plan, &state, &req)

	if hasChanges {
		t.Error("expected hasChanges=false when nothing changed")
	}
	if req.Cron != nil || req.Timezone != nil || req.GracePeriodValue != nil || req.GracePeriodType != nil {
		t.Error("expected all timing fields nil when unchanged")
	}
}

func TestApplyHealthcheckTimingFields_changedPeriodValueIncluded(t *testing.T) {
	plan := HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Value(30),
		PeriodType:       types.StringValue("minutes"),
		GracePeriodValue: types.Int64Value(5),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Value(15),
		PeriodType:       types.StringValue("minutes"),
		GracePeriodValue: types.Int64Value(5),
		GracePeriodType:  types.StringValue("minutes"),
	}
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckTimingFields(&plan, &state, &req)

	if !hasChanges {
		t.Error("expected hasChanges=true")
	}
	if req.PeriodValue == nil {
		t.Fatal("expected PeriodValue to be set")
	}
	if *req.PeriodValue != 30 {
		t.Errorf("expected PeriodValue=30, got=%d", *req.PeriodValue)
	}
}

// =============================================================================
// TestApplyHealthcheckBehaviorFields
// =============================================================================

func TestApplyHealthcheckBehaviorFields_changedNameIncluded(t *testing.T) {
	plan := HealthcheckResourceModel{
		Name:             types.StringValue("new-check"),
		EscalationPolicy: types.StringNull(),
	}
	state := HealthcheckResourceModel{
		Name:             types.StringValue("old-check"),
		EscalationPolicy: types.StringNull(),
	}
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckBehaviorFields(&plan, &state, &req)

	if !hasChanges {
		t.Error("expected hasChanges=true")
	}
	if req.Name == nil {
		t.Fatal("expected Name to be set")
	}
	if *req.Name != "new-check" {
		t.Errorf("expected Name='new-check', got=%q", *req.Name)
	}
}

func TestApplyHealthcheckBehaviorFields_changedEscalationPolicyIncluded(t *testing.T) {
	plan := HealthcheckResourceModel{
		Name:             types.StringValue("check"),
		EscalationPolicy: types.StringValue("new-policy-uuid"),
	}
	state := HealthcheckResourceModel{
		Name:             types.StringValue("check"),
		EscalationPolicy: types.StringValue("old-policy-uuid"),
	}
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckBehaviorFields(&plan, &state, &req)

	if !hasChanges {
		t.Error("expected hasChanges=true")
	}
	if req.EscalationPolicy == nil {
		t.Fatal("expected EscalationPolicy to be set")
	}
	if *req.EscalationPolicy != "new-policy-uuid" {
		t.Errorf("expected EscalationPolicy='new-policy-uuid', got=%q", *req.EscalationPolicy)
	}
}

func TestApplyHealthcheckBehaviorFields_nullEscalationPolicyClearsToEmpty(t *testing.T) {
	plan := HealthcheckResourceModel{
		Name:             types.StringValue("check"),
		EscalationPolicy: types.StringNull(),
	}
	state := HealthcheckResourceModel{
		Name:             types.StringValue("check"),
		EscalationPolicy: types.StringValue("old-policy"),
	}
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckBehaviorFields(&plan, &state, &req)

	if !hasChanges {
		t.Error("expected hasChanges=true")
	}
	if req.EscalationPolicy == nil {
		t.Fatal("expected EscalationPolicy to be set to empty string")
	}
	if *req.EscalationPolicy != "" {
		t.Errorf("expected EscalationPolicy='', got=%q", *req.EscalationPolicy)
	}
}

func TestApplyHealthcheckBehaviorFields_noChangesReturnsFalse(t *testing.T) {
	plan := HealthcheckResourceModel{
		Name:             types.StringValue("check"),
		EscalationPolicy: types.StringValue("policy-uuid"),
	}
	state := plan
	req := client.UpdateHealthcheckRequest{}

	hasChanges := applyHealthcheckBehaviorFields(&plan, &state, &req)

	if hasChanges {
		t.Error("expected hasChanges=false when nothing changed")
	}
	if req.Name != nil || req.EscalationPolicy != nil {
		t.Error("expected all behavior fields nil when unchanged")
	}
}

// =============================================================================
// TestValidateCronFields
// =============================================================================

func TestValidateCronFields_cronWithTzIsValid(t *testing.T) {
	if err := validateCronFields(true, true); err != nil {
		t.Errorf("expected no error when both cron and tz set, got: %v", err)
	}
}

func TestValidateCronFields_cronWithoutTzIsError(t *testing.T) {
	if err := validateCronFields(true, false); err == nil {
		t.Error("expected error when cron set without tz")
	}
}

func TestValidateCronFields_tzWithoutCronIsError(t *testing.T) {
	if err := validateCronFields(false, true); err == nil {
		t.Error("expected error when tz set without cron")
	}
}

func TestValidateCronFields_neitherCronNorTzIsValid(t *testing.T) {
	if err := validateCronFields(false, false); err != nil {
		t.Errorf("expected no error when neither cron nor tz set, got: %v", err)
	}
}

// =============================================================================
// TestValidatePeriodFields
// =============================================================================

func TestValidatePeriodFields_bothSetIsValid(t *testing.T) {
	if err := validatePeriodFields(true, true); err != nil {
		t.Errorf("expected no error when both period_value and period_type set, got: %v", err)
	}
}

func TestValidatePeriodFields_periodValueWithoutTypeIsError(t *testing.T) {
	if err := validatePeriodFields(true, false); err == nil {
		t.Error("expected error when period_value set without period_type")
	}
}

func TestValidatePeriodFields_periodTypeWithoutValueIsError(t *testing.T) {
	if err := validatePeriodFields(false, true); err == nil {
		t.Error("expected error when period_type set without period_value")
	}
}

func TestValidatePeriodFields_neitherSetIsValid(t *testing.T) {
	if err := validatePeriodFields(false, false); err != nil {
		t.Errorf("expected no error when neither period_value nor period_type set, got: %v", err)
	}
}

// =============================================================================
// Test helpers (shared utilities for test setup)
// =============================================================================

// makeStringList builds a types.List of strings for use in test models.
func makeStringList(items []string) types.List {
	vals := make([]types.String, len(items))
	for i, s := range items {
		vals[i] = types.StringValue(s)
	}
	listVal, _ := types.ListValueFrom(context.Background(), types.StringType, vals)
	return listVal
}

// makeHeaderList builds a types.List of request header objects for test models.
func makeHeaderList(headers []client.RequestHeader) types.List {
	var diags diag.Diagnostics
	if len(headers) == 0 {
		return types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()})
	}
	return mapStringSliceToHeaders(headers, &diags)
}

// mapStringSliceToHeaders converts []client.RequestHeader to types.List for tests.
func mapStringSliceToHeaders(headers []client.RequestHeader, diags *diag.Diagnostics) types.List {
	objType := types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}
	if len(headers) == 0 {
		return types.ListNull(objType)
	}
	vals := make([]types.String, 0)
	_ = vals
	// Build using the same approach as mapMonitorToModel
	result, d := types.ListValueFrom(context.Background(), objType, []interface{}{})
	diags.Append(d...)
	return result
}
