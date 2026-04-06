// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"

	hyperping "github.com/develeap/hyperping-go"
)

// ---------------------------------------------------------------------------
// applyHealthcheckTimingFields
// ---------------------------------------------------------------------------

func TestApplyHealthcheckTimingFields_noChanges(t *testing.T) {
	t.Parallel()

	plan := &HealthcheckResourceModel{
		Cron:             types.StringValue("0 * * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := &HealthcheckResourceModel{
		Cron:             types.StringValue("0 * * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}

	var req hyperping.UpdateHealthcheckRequest
	changed := applyHealthcheckTimingFields(plan, state, &req)

	if changed {
		t.Error("expected no changes, got hasChanges=true")
	}
	if req.Cron != nil {
		t.Errorf("expected Cron nil, got %v", *req.Cron)
	}
	if req.Timezone != nil {
		t.Errorf("expected Timezone nil, got %v", *req.Timezone)
	}
	if req.PeriodValue != nil {
		t.Errorf("expected PeriodValue nil, got %v", *req.PeriodValue)
	}
	if req.PeriodType != nil {
		t.Errorf("expected PeriodType nil, got %v", *req.PeriodType)
	}
	if req.GracePeriodValue != nil {
		t.Errorf("expected GracePeriodValue nil, got %v", *req.GracePeriodValue)
	}
	if req.GracePeriodType != nil {
		t.Errorf("expected GracePeriodType nil, got %v", *req.GracePeriodType)
	}
}

func TestApplyHealthcheckTimingFields_cronChanged(t *testing.T) {
	t.Parallel()

	plan := &HealthcheckResourceModel{
		Cron:             types.StringValue("0 0 * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := &HealthcheckResourceModel{
		Cron:             types.StringValue("0 * * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}

	var req hyperping.UpdateHealthcheckRequest
	changed := applyHealthcheckTimingFields(plan, state, &req)

	if !changed {
		t.Error("expected hasChanges=true")
	}
	if req.Cron == nil || *req.Cron != "0 0 * * *" {
		t.Errorf("expected Cron=%q, got %v", "0 0 * * *", req.Cron)
	}
}

func TestApplyHealthcheckTimingFields_cronRemoved(t *testing.T) {
	t.Parallel()

	plan := &HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Value(60),
		PeriodType:       types.StringValue("minutes"),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := &HealthcheckResourceModel{
		Cron:             types.StringValue("0 * * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}

	var req hyperping.UpdateHealthcheckRequest
	changed := applyHealthcheckTimingFields(plan, state, &req)

	if !changed {
		t.Error("expected hasChanges=true")
	}
	if req.Cron == nil || *req.Cron != "" {
		t.Errorf("expected Cron empty string for removal, got %v", req.Cron)
	}
	if req.Timezone == nil || *req.Timezone != "" {
		t.Errorf("expected Timezone empty string for removal, got %v", req.Timezone)
	}
}

func TestApplyHealthcheckTimingFields_periodValueChanged(t *testing.T) {
	t.Parallel()

	plan := &HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Value(120),
		PeriodType:       types.StringValue("minutes"),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := &HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Value(60),
		PeriodType:       types.StringValue("minutes"),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}

	var req hyperping.UpdateHealthcheckRequest
	changed := applyHealthcheckTimingFields(plan, state, &req)

	if !changed {
		t.Error("expected hasChanges=true")
	}
	if req.PeriodValue == nil || *req.PeriodValue != 120 {
		t.Errorf("expected PeriodValue=120, got %v", req.PeriodValue)
	}
}

func TestApplyHealthcheckTimingFields_periodValueRemoved(t *testing.T) {
	t.Parallel()

	// When switching from period mode to cron mode, PeriodValue becomes null.
	// The function should NOT send PeriodValue (omits it); the comment in the
	// source says: "When clearing period_value (switching to cron mode), omit
	// the field entirely."
	plan := &HealthcheckResourceModel{
		Cron:             types.StringValue("0 * * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}
	state := &HealthcheckResourceModel{
		Cron:             types.StringNull(),
		Timezone:         types.StringNull(),
		PeriodValue:      types.Int64Value(60),
		PeriodType:       types.StringValue("minutes"),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}

	var req hyperping.UpdateHealthcheckRequest
	changed := applyHealthcheckTimingFields(plan, state, &req)

	if !changed {
		t.Error("expected hasChanges=true")
	}
	// PeriodValue is omitted (nil) when being cleared.
	if req.PeriodValue != nil {
		t.Errorf("expected PeriodValue nil (omitted), got %v", *req.PeriodValue)
	}
	// PeriodType should be sent as empty string to signal clearing.
	if req.PeriodType == nil || *req.PeriodType != "" {
		t.Errorf("expected PeriodType empty string, got %v", req.PeriodType)
	}
}

func TestApplyHealthcheckTimingFields_gracePeriodChanged(t *testing.T) {
	t.Parallel()

	plan := &HealthcheckResourceModel{
		Cron:             types.StringValue("0 * * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(60),
		GracePeriodType:  types.StringValue("hours"),
	}
	state := &HealthcheckResourceModel{
		Cron:             types.StringValue("0 * * * *"),
		Timezone:         types.StringValue("UTC"),
		PeriodValue:      types.Int64Null(),
		PeriodType:       types.StringNull(),
		GracePeriodValue: types.Int64Value(30),
		GracePeriodType:  types.StringValue("minutes"),
	}

	var req hyperping.UpdateHealthcheckRequest
	changed := applyHealthcheckTimingFields(plan, state, &req)

	if !changed {
		t.Error("expected hasChanges=true")
	}
	if req.GracePeriodValue == nil || *req.GracePeriodValue != 60 {
		t.Errorf("expected GracePeriodValue=60, got %v", req.GracePeriodValue)
	}
	if req.GracePeriodType == nil || *req.GracePeriodType != "hours" {
		t.Errorf("expected GracePeriodType=%q, got %v", "hours", req.GracePeriodType)
	}
}

// ---------------------------------------------------------------------------
// validateCronPeriodExclusivity
// ---------------------------------------------------------------------------

func TestValidateCronPeriodExclusivity_cronModeOnly(t *testing.T) {
	t.Parallel()

	model := &HealthcheckResourceModel{
		Cron:        types.StringValue("0 * * * *"),
		Timezone:    types.StringValue("UTC"),
		PeriodValue: types.Int64Null(),
		PeriodType:  types.StringNull(),
	}

	err := validateCronPeriodExclusivity(model)
	if err != nil {
		t.Errorf("expected no error for cron-only mode, got: %v", err)
	}
}

func TestValidateCronPeriodExclusivity_periodModeOnly(t *testing.T) {
	t.Parallel()

	model := &HealthcheckResourceModel{
		Cron:        types.StringNull(),
		Timezone:    types.StringNull(),
		PeriodValue: types.Int64Value(60),
		PeriodType:  types.StringValue("minutes"),
	}

	err := validateCronPeriodExclusivity(model)
	if err != nil {
		t.Errorf("expected no error for period-only mode, got: %v", err)
	}
}

func TestValidateCronPeriodExclusivity_bothSet(t *testing.T) {
	t.Parallel()

	model := &HealthcheckResourceModel{
		Cron:        types.StringValue("0 * * * *"),
		Timezone:    types.StringValue("UTC"),
		PeriodValue: types.Int64Value(60),
		PeriodType:  types.StringValue("minutes"),
	}

	err := validateCronPeriodExclusivity(model)
	if err == nil {
		t.Fatal("expected error when both cron and period are set")
	}
	expected := "specify either (cron + tz) or (period_value + period_type), not both"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateCronPeriodExclusivity_neitherSet(t *testing.T) {
	t.Parallel()

	model := &HealthcheckResourceModel{
		Cron:        types.StringNull(),
		Timezone:    types.StringNull(),
		PeriodValue: types.Int64Null(),
		PeriodType:  types.StringNull(),
	}

	err := validateCronPeriodExclusivity(model)
	if err == nil {
		t.Fatal("expected error when neither cron nor period is set")
	}
	expected := "either cron or period_value must be specified"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateCronPeriodExclusivity_cronWithoutTimezone(t *testing.T) {
	t.Parallel()

	model := &HealthcheckResourceModel{
		Cron:        types.StringValue("0 * * * *"),
		Timezone:    types.StringNull(),
		PeriodValue: types.Int64Null(),
		PeriodType:  types.StringNull(),
	}

	err := validateCronPeriodExclusivity(model)
	if err == nil {
		t.Fatal("expected error when cron is set without timezone")
	}
	expected := "tz is required when cron is set"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateCronPeriodExclusivity_timezoneWithoutCron(t *testing.T) {
	t.Parallel()

	model := &HealthcheckResourceModel{
		Cron:        types.StringNull(),
		Timezone:    types.StringValue("UTC"),
		PeriodValue: types.Int64Null(),
		PeriodType:  types.StringNull(),
	}

	err := validateCronPeriodExclusivity(model)
	if err == nil {
		t.Fatal("expected error when timezone is set without cron")
	}
	expected := "cron is required when tz is set"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateCronPeriodExclusivity_periodValueWithoutType(t *testing.T) {
	t.Parallel()

	model := &HealthcheckResourceModel{
		Cron:        types.StringNull(),
		Timezone:    types.StringNull(),
		PeriodValue: types.Int64Value(60),
		PeriodType:  types.StringNull(),
	}

	err := validateCronPeriodExclusivity(model)
	if err == nil {
		t.Fatal("expected error when period_value is set without period_type")
	}
	expected := "period_type is required when period_value is set"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}

func TestValidateCronPeriodExclusivity_periodTypeWithoutValue(t *testing.T) {
	t.Parallel()

	model := &HealthcheckResourceModel{
		Cron:        types.StringNull(),
		Timezone:    types.StringNull(),
		PeriodValue: types.Int64Null(),
		PeriodType:  types.StringValue("minutes"),
	}

	err := validateCronPeriodExclusivity(model)
	if err == nil {
		t.Fatal("expected error when period_type is set without period_value")
	}
	expected := "period_value is required when period_type is set"
	if err.Error() != expected {
		t.Errorf("expected error %q, got %q", expected, err.Error())
	}
}
