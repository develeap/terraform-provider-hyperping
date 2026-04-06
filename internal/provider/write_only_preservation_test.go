// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	hyperping "github.com/develeap/hyperping-go"
)

// =============================================================================
// Write-Only Field Preservation Tests
//
// The Hyperping API has several write-only fields that are accepted on
// POST/PUT but not returned (or returned as empty) in GET responses.
// The provider preserves these values from plan/state to prevent drift.
//
// Write-only fields:
//   - password  (statuspage resource) - handled in Read/Update callers
//   - text      (incident resource)   - handled in mapIncidentToModel
//   - text      (maintenance resource) - handled in mapMaintenanceToModel
//   - is_split  (statuspage sections)  - handled by preserveSectionIsSplit
// =============================================================================

// --- Password preservation (statuspage resource) ---
// The password field is preserved at the caller level in Read/Update.
// These tests verify the pattern: save before mapping, restore after.

func TestPreservePassword_ReadPattern(t *testing.T) {
	tests := []struct {
		name          string
		priorPassword types.String
		wantPreserved bool
		wantValue     string
	}{
		{
			name:          "non-null password is preserved",
			priorPassword: types.StringValue("my-secret-pass"),
			wantPreserved: true,
			wantValue:     "my-secret-pass",
		},
		{
			name:          "null password is not restored",
			priorPassword: types.StringNull(),
			wantPreserved: false,
		},
		{
			name:          "unknown password is not restored",
			priorPassword: types.StringUnknown(),
			wantPreserved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the Read pattern from statuspage_resource.go:
			// 1. Save password before API call
			priorPassword := tt.priorPassword

			// 2. After mapping, password is empty (API never returns it)
			state := &StatusPageResourceModel{}
			// (mapStatusPageToModel would run here but doesn't set password)

			// 3. Restore password from prior state
			if !priorPassword.IsNull() {
				state.Password = priorPassword
			}

			if tt.wantPreserved {
				if state.Password.IsNull() {
					t.Error("expected password to be preserved, got null")
				}
				if state.Password.ValueString() != tt.wantValue {
					t.Errorf("expected password %q, got %q", tt.wantValue, state.Password.ValueString())
				}
			} else {
				if !state.Password.IsNull() && state.Password.ValueString() != "" {
					t.Errorf("expected password to remain unset, got %q", state.Password.ValueString())
				}
			}
		})
	}
}

func TestPreservePassword_UpdatePattern(t *testing.T) {
	tests := []struct {
		name          string
		planPassword  types.String
		wantPreserved bool
		wantValue     string
	}{
		{
			name:          "plan password is preserved after mapping",
			planPassword:  types.StringValue("new-password"),
			wantPreserved: true,
			wantValue:     "new-password",
		},
		{
			name:          "null plan password is not restored",
			planPassword:  types.StringNull(),
			wantPreserved: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the Update pattern from statuspage_resource.go:
			// 1. Save password from plan before mapping
			planPassword := tt.planPassword

			// 2. Mapping overwrites plan (password is not set by mapping)
			plan := &StatusPageResourceModel{}

			// 3. Restore password from plan
			if !planPassword.IsNull() {
				plan.Password = planPassword
			}

			if tt.wantPreserved {
				if plan.Password.IsNull() {
					t.Error("expected password to be preserved, got null")
				}
				if plan.Password.ValueString() != tt.wantValue {
					t.Errorf("expected password %q, got %q", tt.wantValue, plan.Password.ValueString())
				}
			} else {
				if !plan.Password.IsNull() && plan.Password.ValueString() != "" {
					t.Errorf("expected password to remain unset, got %q", plan.Password.ValueString())
				}
			}
		})
	}
}

// --- Text preservation (incident resource) ---

func TestPreserveIncidentText_WriteOnlyBehavior(t *testing.T) {
	r := &IncidentResource{}

	tests := []struct {
		name         string
		apiText      string
		priorText    types.String
		expectedText string
	}{
		{
			name:         "API returns text - uses API value",
			apiText:      "API returned text",
			priorText:    types.StringValue("plan text"),
			expectedText: "API returned text",
		},
		{
			name:         "API returns empty - preserves plan value",
			apiText:      "",
			priorText:    types.StringValue("plan text that should survive"),
			expectedText: "plan text that should survive",
		},
		{
			name:         "API returns empty - no prior text - stays empty",
			apiText:      "",
			priorText:    types.StringValue(""),
			expectedText: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			incident := &hyperping.Incident{
				UUID:        "inci_test",
				Title:       hyperping.LocalizedText{En: "Test"},
				Text:        hyperping.LocalizedText{En: tt.apiText},
				Type:        "incident",
				StatusPages: []string{"sp-1"},
			}

			model := &IncidentResourceModel{
				Text: tt.priorText,
			}
			diags := &diag.Diagnostics{}
			r.mapIncidentToModel(incident, model, diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}
			if model.Text.ValueString() != tt.expectedText {
				t.Errorf("expected text %q, got %q", tt.expectedText, model.Text.ValueString())
			}
		})
	}
}

// --- Text preservation (maintenance resource) ---

func TestPreserveMaintenanceText_WriteOnlyBehavior(t *testing.T) {
	r := &MaintenanceResource{}
	startDate := "2025-12-20T02:00:00.000Z"
	endDate := "2025-12-20T06:00:00.000Z"

	tests := []struct {
		name         string
		apiText      string
		priorText    types.String
		expectedText string
		expectNull   bool
	}{
		{
			name:         "API returns text - uses API value",
			apiText:      "API returned text",
			priorText:    types.StringValue("plan text"),
			expectedText: "API returned text",
		},
		{
			name:         "API returns empty - preserves plan value",
			apiText:      "",
			priorText:    types.StringValue("plan text that should survive"),
			expectedText: "plan text that should survive",
		},
		{
			name:       "API returns empty - null prior text - stays null",
			apiText:    "",
			priorText:  types.StringNull(),
			expectNull: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maintenance := &hyperping.Maintenance{
				UUID:      "mw_test",
				Name:      "Test",
				Title:     hyperping.LocalizedText{En: "Title"},
				Text:      hyperping.LocalizedText{En: tt.apiText},
				StartDate: &startDate,
				EndDate:   &endDate,
				Monitors:  []string{"mon-1"},
			}

			model := &MaintenanceResourceModel{
				Text: tt.priorText,
			}
			diags := &diag.Diagnostics{}
			r.mapMaintenanceToModel(maintenance, model, diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}
			if tt.expectNull {
				if !model.Text.IsNull() {
					t.Errorf("expected text to be null, got %q", model.Text.ValueString())
				}
			} else if model.Text.ValueString() != tt.expectedText {
				t.Errorf("expected text %q, got %q", tt.expectedText, model.Text.ValueString())
			}
		})
	}
}
