// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestMapHealthcheckCommonFields_NilPointer(t *testing.T) {
	f := MapHealthcheckCommonFields(nil)
	if !f.ID.IsNull() {
		t.Error("expected ID to be null for nil healthcheck")
	}
	if !f.Name.IsNull() {
		t.Error("expected Name to be null for nil healthcheck")
	}
	if !f.IsPaused.IsNull() {
		t.Error("expected IsPaused to be null for nil healthcheck")
	}
}

func TestMapHealthcheckCommonFields_FullHealthcheck(t *testing.T) {
	periodValue := 60
	hc := &client.Healthcheck{
		UUID:             "hc_abc123",
		Name:             "Test HC",
		PingURL:          "https://hb.tinyping.io/hc_abc123",
		Cron:             "* * * * *",
		Tz:               "UTC",
		PeriodValue:      &periodValue,
		PeriodType:       "seconds",
		GracePeriodValue: 30,
		GracePeriodType:  "seconds",
		IsPaused:         false,
		IsDown:           true,
		Period:           60,
		GracePeriod:      30,
		LastPing:         "2026-01-28T10:00:00Z",
		CreatedAt:        "2026-01-01T00:00:00Z",
	}

	f := MapHealthcheckCommonFields(hc)

	if f.ID.ValueString() != "hc_abc123" {
		t.Errorf("expected ID hc_abc123, got %s", f.ID.ValueString())
	}
	if f.Name.ValueString() != "Test HC" {
		t.Errorf("expected Name 'Test HC', got %s", f.Name.ValueString())
	}
	if f.Cron.ValueString() != "* * * * *" {
		t.Errorf("expected Cron '* * * * *', got %s", f.Cron.ValueString())
	}
	if f.Tz.ValueString() != "UTC" {
		t.Errorf("expected Tz 'UTC', got %s", f.Tz.ValueString())
	}
	if f.PeriodValue.ValueInt64() != 60 {
		t.Errorf("expected PeriodValue 60, got %d", f.PeriodValue.ValueInt64())
	}
	if f.IsDown.ValueBool() != true {
		t.Error("expected IsDown true")
	}
	if f.IsPaused.ValueBool() != false {
		t.Error("expected IsPaused false")
	}
	if f.LastPing.ValueString() != "2026-01-28T10:00:00Z" {
		t.Errorf("expected LastPing '2026-01-28T10:00:00Z', got %s", f.LastPing.ValueString())
	}
}

func TestMapHealthcheckCommonFields_NullOptionalFields(t *testing.T) {
	hc := &client.Healthcheck{
		UUID:             "hc_min",
		Name:             "Minimal",
		GracePeriodValue: 10,
		GracePeriodType:  "seconds",
	}

	f := MapHealthcheckCommonFields(hc)

	if !f.Cron.IsNull() {
		t.Error("expected Cron to be null when empty")
	}
	if !f.Tz.IsNull() {
		t.Error("expected Tz to be null when empty")
	}
	if !f.PeriodValue.IsNull() {
		t.Error("expected PeriodValue to be null when nil")
	}
	if !f.PeriodType.IsNull() {
		t.Error("expected PeriodType to be null when empty")
	}
	if !f.EscalationPolicy.IsNull() {
		t.Error("expected EscalationPolicy to be null when nil")
	}
	if !f.LastPing.IsNull() {
		t.Error("expected LastPing to be null when empty")
	}
	if !f.CreatedAt.IsNull() {
		t.Error("expected CreatedAt to be null when empty")
	}
}

func TestMapOutageNestedObjects_NilOutage(t *testing.T) {
	var diags diag.Diagnostics
	monitorObj, ackObj := MapOutageNestedObjects(nil, &diags)

	if !monitorObj.IsNull() {
		t.Error("expected monitor object to be null for nil outage")
	}
	if !ackObj.IsNull() {
		t.Error("expected acknowledged_by object to be null for nil outage")
	}
	if diags.HasError() {
		t.Errorf("unexpected errors: %s", diags.Errors())
	}
}

func TestMapOutageNestedObjects_ZeroValueMonitor(t *testing.T) {
	outage := &client.Outage{
		UUID: "out_123",
		// Monitor is zero-value: all empty strings
	}

	var diags diag.Diagnostics
	monitorObj, ackObj := MapOutageNestedObjects(outage, &diags)

	if !monitorObj.IsNull() {
		t.Error("expected monitor object to be null for zero-value MonitorReference")
	}
	if !ackObj.IsNull() {
		t.Error("expected acknowledged_by object to be null when AcknowledgedBy is nil")
	}
}

func TestMapOutageNestedObjects_FullOutage(t *testing.T) {
	outage := &client.Outage{
		UUID: "out_456",
		Monitor: client.MonitorReference{
			UUID:     "mon_789",
			Name:     "API Monitor",
			URL:      "https://api.example.com",
			Protocol: "HTTPS",
		},
		AcknowledgedBy: &client.AcknowledgedByUser{
			UUID:  "user_001",
			Email: "ops@example.com",
			Name:  "Ops Team",
		},
	}

	var diags diag.Diagnostics
	monitorObj, ackObj := MapOutageNestedObjects(outage, &diags)

	if monitorObj.IsNull() {
		t.Error("expected monitor object to be populated")
	}
	if ackObj.IsNull() {
		t.Error("expected acknowledged_by object to be populated")
	}
	if diags.HasError() {
		t.Errorf("unexpected errors: %s", diags.Errors())
	}

	monAttrs := monitorObj.Attributes()
	if monAttrs["uuid"].String() != `"mon_789"` {
		t.Errorf("expected monitor uuid 'mon_789', got %s", monAttrs["uuid"].String())
	}

	ackAttrs := ackObj.Attributes()
	if ackAttrs["email"].String() != `"ops@example.com"` {
		t.Errorf("expected ack email 'ops@example.com', got %s", ackAttrs["email"].String())
	}
}

func TestValidateCronPeriodExclusivity_BothSet(t *testing.T) {
	plan := buildTestPlan("* * * * *", "UTC", intPtr(60), "seconds")
	err := validateCronPeriodExclusivity(&plan)
	if err == nil {
		t.Error("expected error when both cron and period are set")
	}
}

func TestValidateCronPeriodExclusivity_CronOnly(t *testing.T) {
	plan := buildTestPlan("* * * * *", "UTC", nil, "")
	err := validateCronPeriodExclusivity(&plan)
	if err != nil {
		t.Errorf("unexpected error for cron-only: %s", err)
	}
}

func TestValidateCronPeriodExclusivity_PeriodOnly(t *testing.T) {
	plan := buildTestPlan("", "", intPtr(60), "seconds")
	err := validateCronPeriodExclusivity(&plan)
	if err != nil {
		t.Errorf("unexpected error for period-only: %s", err)
	}
}

func TestValidateCronPeriodExclusivity_CronWithoutTz(t *testing.T) {
	plan := buildTestPlan("* * * * *", "", nil, "")
	err := validateCronPeriodExclusivity(&plan)
	if err == nil {
		t.Error("expected error: tz required when cron is set")
	}
}

func TestValidateCronPeriodExclusivity_TzWithoutCron(t *testing.T) {
	plan := buildTestPlan("", "UTC", nil, "")
	err := validateCronPeriodExclusivity(&plan)
	if err == nil {
		t.Error("expected error: cron required when tz is set")
	}
}

func TestValidateCronPeriodExclusivity_PeriodValueWithoutType(t *testing.T) {
	plan := buildTestPlan("", "", intPtr(60), "")
	err := validateCronPeriodExclusivity(&plan)
	if err == nil {
		t.Error("expected error: period_type required when period_value is set")
	}
}

func TestValidateCronPeriodExclusivity_PeriodTypeWithoutValue(t *testing.T) {
	plan := buildTestPlan("", "", nil, "seconds")
	err := validateCronPeriodExclusivity(&plan)
	if err == nil {
		t.Error("expected error: period_value required when period_type is set")
	}
}

func TestValidateCronPeriodExclusivity_NeitherSet(t *testing.T) {
	plan := buildTestPlan("", "", nil, "")
	err := validateCronPeriodExclusivity(&plan)
	if err == nil {
		t.Error("expected error: either cron or period_value must be specified")
	}
}

func TestValidateCronPeriodExclusivity_EmptyCronString(t *testing.T) {
	// Empty string cron should be treated as "not set"
	plan := buildTestPlan("", "UTC", nil, "")
	err := validateCronPeriodExclusivity(&plan)
	if err == nil {
		t.Error("expected error: tz without cron (empty string cron is not set)")
	}
}

func TestValidateCronPeriodExclusivity_EmptyTzString(t *testing.T) {
	// Empty string tz should be treated as "not set"
	plan := buildTestPlan("* * * * *", "", nil, "")
	err := validateCronPeriodExclusivity(&plan)
	if err == nil {
		t.Error("expected error: cron without tz (empty string tz is not set)")
	}
}

// buildTestPlan creates a HealthcheckResourceModel with the given scheduling fields.
func buildTestPlan(cron, tz string, periodValue *int, periodType string) HealthcheckResourceModel {
	plan := HealthcheckResourceModel{
		GracePeriodValue: tfInt64(10),
		GracePeriodType:  tfString("seconds"),
	}

	if cron != "" {
		plan.Cron = tfString(cron)
	} else {
		plan.Cron = tfStringNull()
	}
	if tz != "" {
		plan.Tz = tfString(tz)
	} else {
		plan.Tz = tfStringNull()
	}
	if periodValue != nil {
		plan.PeriodValue = tfInt64(int64(*periodValue))
	} else {
		plan.PeriodValue = tfInt64Null()
	}
	if periodType != "" {
		plan.PeriodType = tfString(periodType)
	} else {
		plan.PeriodType = tfStringNull()
	}

	return plan
}

func intPtr(v int) *int { return &v }

// Terraform types helpers for tests.
func tfString(v string) types.String { return types.StringValue(v) }
func tfStringNull() types.String     { return types.StringNull() }
func tfInt64(v int64) types.Int64    { return types.Int64Value(v) }
func tfInt64Null() types.Int64       { return types.Int64Null() }
