// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestMaintenanceValidateConfig_SchemaConsistency verifies that the attributes
// accessed by ValidateConfig exist in the real MaintenanceResource schema.
func TestMaintenanceValidateConfig_SchemaConsistency(t *testing.T) {
	r := &MaintenanceResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	realSchema := schemaResp.Schema

	expectations := []struct {
		name     string
		wantType string
	}{
		{"start_date", "schema.StringAttribute"},
		{"end_date", "schema.StringAttribute"},
	}

	for _, exp := range expectations {
		attr, ok := realSchema.Attributes[exp.name]
		if !ok {
			t.Errorf("validation references attribute %q but it does not exist in the real schema", exp.name)
			continue
		}
		gotType := fmt.Sprintf("%T", attr)
		if gotType != exp.wantType {
			t.Errorf("attribute %q: expected type %s, got %s", exp.name, exp.wantType, gotType)
		}
	}
}

// maintenanceConfigBuilder constructs tftypes.Value objects for ValidateConfig tests.
type maintenanceConfigBuilder struct {
	startDate interface{} // string, nil (null), or tftypes.UnknownValue
	endDate   interface{} // string, nil (null), or tftypes.UnknownValue
}

func (b *maintenanceConfigBuilder) buildConfigValue(s schema.Schema) tftypes.Value {
	attrTypes := make(map[string]tftypes.Type)
	for name, attr := range s.Attributes {
		attrTypes[name] = attr.GetType().TerraformType(context.Background())
	}
	objType := tftypes.Object{AttributeTypes: attrTypes}

	vals := make(map[string]tftypes.Value)
	for name, attrType := range attrTypes {
		vals[name] = tftypes.NewValue(attrType, nil) // null by default
	}

	// Required fields need values to avoid unrelated errors.
	vals["name"] = tftypes.NewValue(tftypes.String, "test-maintenance")
	vals["start_date"] = buildMaintenanceTFStringValue(b.startDate)
	vals["end_date"] = buildMaintenanceTFStringValue(b.endDate)

	return tftypes.NewValue(objType, vals)
}

func buildMaintenanceTFStringValue(v interface{}) tftypes.Value {
	switch val := v.(type) {
	case string:
		return tftypes.NewValue(tftypes.String, val)
	case nil:
		return tftypes.NewValue(tftypes.String, nil)
	default:
		return tftypes.NewValue(tftypes.String, tftypes.UnknownValue)
	}
}

func runMaintenanceValidateConfig(t *testing.T, b *maintenanceConfigBuilder) *resource.ValidateConfigResponse {
	t.Helper()

	r := &MaintenanceResource{}
	ctx := context.Background()

	schemaResp := &resource.SchemaResponse{}
	r.Schema(ctx, resource.SchemaRequest{}, schemaResp)

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    b.buildConfigValue(schemaResp.Schema),
	}

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	return resp
}

func TestMaintenanceValidateConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    maintenanceConfigBuilder
		wantError bool
		errMatch  string
	}{
		{
			name: "valid date range",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T00:00:00Z",
				endDate:   "2026-06-01T04:00:00Z",
			},
		},
		{
			name: "end_date before start_date",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T04:00:00Z",
				endDate:   "2026-06-01T00:00:00Z",
			},
			wantError: true,
			errMatch:  "end_date must be after start_date",
		},
		{
			name: "end_date equal to start_date",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T04:00:00Z",
				endDate:   "2026-06-01T04:00:00Z",
			},
			wantError: true,
			errMatch:  "end_date must be after start_date",
		},
		{
			name: "start_date unknown skips validation",
			config: maintenanceConfigBuilder{
				startDate: tftypes.UnknownValue,
				endDate:   "2026-06-01T04:00:00Z",
			},
		},
		{
			name: "end_date unknown skips validation",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T00:00:00Z",
				endDate:   tftypes.UnknownValue,
			},
		},
		{
			name: "both dates unknown skips validation",
			config: maintenanceConfigBuilder{
				startDate: tftypes.UnknownValue,
				endDate:   tftypes.UnknownValue,
			},
		},
		{
			name: "start_date null skips validation",
			config: maintenanceConfigBuilder{
				startDate: nil,
				endDate:   "2026-06-01T04:00:00Z",
			},
		},
		{
			name: "end_date null skips validation",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T00:00:00Z",
				endDate:   nil,
			},
		},
		{
			name: "unparseable start_date defers to ISO8601 validator",
			config: maintenanceConfigBuilder{
				startDate: "not-a-date",
				endDate:   "2026-06-01T04:00:00Z",
			},
		},
		{
			name: "unparseable end_date defers to ISO8601 validator",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T00:00:00Z",
				endDate:   "not-a-date",
			},
		},
		{
			name: "date without timezone defers to ISO8601 validator",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T00:00:00",
				endDate:   "2026-06-01T04:00:00",
			},
		},
		{
			name: "1 second difference is valid",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T00:00:00Z",
				endDate:   "2026-06-01T00:00:01Z",
			},
		},
		{
			name: "different timezones end after start",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T12:00:00+00:00",
				endDate:   "2026-06-01T10:00:00-04:00", // 14:00 UTC > 12:00 UTC
			},
		},
		{
			name: "different timezones end before start",
			config: maintenanceConfigBuilder{
				startDate: "2026-06-01T12:00:00+00:00",
				endDate:   "2026-06-01T10:00:00+00:00",
			},
			wantError: true,
			errMatch:  "end_date must be after start_date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := runMaintenanceValidateConfig(t, &tt.config)

			if tt.wantError && !resp.Diagnostics.HasError() {
				t.Error("expected validation error, got none")
			}
			if !tt.wantError && resp.Diagnostics.HasError() {
				t.Errorf("unexpected validation error: %v", resp.Diagnostics)
			}
			if tt.wantError && tt.errMatch != "" {
				found := false
				for _, d := range resp.Diagnostics.Errors() {
					if containsString(d.Detail(), tt.errMatch) || containsString(d.Summary(), tt.errMatch) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error containing %q, got: %v", tt.errMatch, resp.Diagnostics)
				}
			}
		})
	}
}
