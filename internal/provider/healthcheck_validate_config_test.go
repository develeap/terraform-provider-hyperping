// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// TestHealthcheckValidateConfig_SchemaConsistency verifies that the attributes
// accessed by ValidateConfig exist in the real HealthcheckResource schema.
func TestHealthcheckValidateConfig_SchemaConsistency(t *testing.T) {
	r := &HealthcheckResource{}
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)
	realSchema := schemaResp.Schema

	expectations := []struct {
		name     string
		wantType string
	}{
		{"cron", "schema.StringAttribute"},
		{"period_value", "schema.Int64Attribute"},
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

// healthcheckConfigBuilder constructs tftypes.Value objects for ValidateConfig tests.
type healthcheckConfigBuilder struct {
	cron        interface{} // string, nil (null), or tftypes.UnknownValue
	periodValue interface{} // int64, nil (null), or tftypes.UnknownValue
}

func (b *healthcheckConfigBuilder) buildConfigValue(s schema.Schema) tftypes.Value {
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
	vals["name"] = tftypes.NewValue(tftypes.String, "test-healthcheck")
	vals["grace_period_value"] = tftypes.NewValue(tftypes.Number, int64(5))
	vals["grace_period_type"] = tftypes.NewValue(tftypes.String, "minutes")

	vals["cron"] = buildHealthcheckTFValue(b.cron, tftypes.String)
	vals["period_value"] = buildHealthcheckTFValue(b.periodValue, tftypes.Number)

	return tftypes.NewValue(objType, vals)
}

func buildHealthcheckTFValue(v interface{}, tfType tftypes.Type) tftypes.Value {
	switch val := v.(type) {
	case string:
		return tftypes.NewValue(tfType, val)
	case int64:
		return tftypes.NewValue(tfType, val)
	case nil:
		return tftypes.NewValue(tfType, nil)
	default:
		return tftypes.NewValue(tfType, tftypes.UnknownValue)
	}
}

func runHealthcheckValidateConfig(t *testing.T, b *healthcheckConfigBuilder) *resource.ValidateConfigResponse {
	t.Helper()

	r := &HealthcheckResource{}
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

func TestHealthcheckValidateConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		config    healthcheckConfigBuilder
		wantError bool
		errMatch  string
	}{
		{
			name: "cron only is valid",
			config: healthcheckConfigBuilder{
				cron:        "0 0 * * *",
				periodValue: nil,
			},
		},
		{
			name: "period_value only is valid",
			config: healthcheckConfigBuilder{
				cron:        nil,
				periodValue: int64(60),
			},
		},
		{
			name: "both cron and period_value is invalid",
			config: healthcheckConfigBuilder{
				cron:        "0 0 * * *",
				periodValue: int64(60),
			},
			wantError: true,
			errMatch:  "not both",
		},
		{
			name: "neither cron nor period_value is invalid",
			config: healthcheckConfigBuilder{
				cron:        nil,
				periodValue: nil,
			},
			wantError: true,
			errMatch:  "Either cron or period_value must be specified",
		},
		{
			name: "cron unknown skips validation",
			config: healthcheckConfigBuilder{
				cron:        tftypes.UnknownValue,
				periodValue: int64(60),
			},
		},
		{
			name: "period_value unknown skips validation",
			config: healthcheckConfigBuilder{
				cron:        "0 0 * * *",
				periodValue: tftypes.UnknownValue,
			},
		},
		{
			name: "both unknown skips validation",
			config: healthcheckConfigBuilder{
				cron:        tftypes.UnknownValue,
				periodValue: tftypes.UnknownValue,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := runHealthcheckValidateConfig(t, &tt.config)

			if tt.wantError && !resp.Diagnostics.HasError() {
				t.Error("expected validation error, got none")
			}
			if !tt.wantError && resp.Diagnostics.HasError() {
				t.Errorf("unexpected validation error: %v", resp.Diagnostics)
			}
			if tt.wantError && tt.errMatch != "" {
				found := false
				for _, d := range resp.Diagnostics.Errors() {
					if strings.Contains(d.Detail(), tt.errMatch) || strings.Contains(d.Summary(), tt.errMatch) {
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
