// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Test RequestHeaderAttrTypes returns expected keys
func TestRequestHeaderAttrTypes_Coverage(t *testing.T) {
	attrTypes := RequestHeaderAttrTypes()

	expectedKeys := []string{"name", "value"}
	if len(attrTypes) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d", len(expectedKeys), len(attrTypes))
	}

	for _, key := range expectedKeys {
		if _, ok := attrTypes[key]; !ok {
			t.Errorf("missing expected key: %s", key)
		}
	}
}

// Test monitorReferenceAttrTypes
func TestMonitorReferenceAttrTypes_Coverage(t *testing.T) {
	attrTypes := monitorReferenceAttrTypes()

	expectedKeys := []string{"uuid", "name", "url", "protocol"}
	for _, key := range expectedKeys {
		if _, ok := attrTypes[key]; !ok {
			t.Errorf("missing expected key: %s", key)
		}
	}
}

// Test acknowledgedByAttrTypes
func TestAcknowledgedByAttrTypes_Coverage(t *testing.T) {
	attrTypes := acknowledgedByAttrTypes()

	expectedKeys := []string{"uuid", "email", "name"}
	for _, key := range expectedKeys {
		if _, ok := attrTypes[key]; !ok {
			t.Errorf("missing expected key: %s", key)
		}
	}
}

// Comprehensive test for mapStringSliceToList
func TestMapStringSliceToList_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		wantNull bool
		wantLen  int
	}{
		{"nil", nil, true, 0},
		{"empty", []string{}, true, 0},
		{"single", []string{"london"}, false, 1},
		{"multiple", []string{"london", "frankfurt", "virginia"}, false, 3},
		{"with empty strings", []string{"", "test", ""}, false, 3},
		{"unicode", []string{"london", "東京", "مومباي"}, false, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapStringSliceToList(tt.input, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}

			if result.IsNull() != tt.wantNull {
				t.Errorf("IsNull = %v, want %v", result.IsNull(), tt.wantNull)
			}

			if !tt.wantNull {
				elements := result.Elements()
				if len(elements) != tt.wantLen {
					t.Errorf("got %d elements, want %d", len(elements), tt.wantLen)
				}
			}
		})
	}
}

// Comprehensive test for mapRequestHeadersToTFList
func TestMapRequestHeadersToTFList_Comprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    []client.RequestHeader
		wantNull bool
		wantLen  int
	}{
		{
			name:     "nil",
			input:    nil,
			wantNull: true,
		},
		{
			name:     "empty",
			input:    []client.RequestHeader{},
			wantNull: true,
		},
		{
			name: "single",
			input: []client.RequestHeader{
				{Name: "Authorization", Value: "Bearer token"},
			},
			wantNull: false,
			wantLen:  1,
		},
		{
			name: "multiple",
			input: []client.RequestHeader{
				{Name: "Authorization", Value: "Bearer token"},
				{Name: "Accept", Value: "application/json"},
				{Name: "X-Custom", Value: "value"},
			},
			wantNull: false,
			wantLen:  3,
		},
		{
			name: "with special characters",
			input: []client.RequestHeader{
				{Name: "X-Test", Value: "value with spaces"},
				{Name: "X-Unicode", Value: "测试值"},
			},
			wantNull: false,
			wantLen:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := mapRequestHeadersToTFList(tt.input, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}

			if result.IsNull() != tt.wantNull {
				t.Errorf("IsNull = %v, want %v", result.IsNull(), tt.wantNull)
			}

			if !tt.wantNull {
				elements := result.Elements()
				if len(elements) != tt.wantLen {
					t.Errorf("got %d elements, want %d", len(elements), tt.wantLen)
				}
			}
		})
	}
}

// Comprehensive test for MapMonitorCommonFields
func TestMapMonitorCommonFields_Comprehensive(t *testing.T) {
	port443 := 443
	port8080 := 8080
	requiredKeyword := "success"
	escalationPolicy := "ep_123"

	tests := []struct {
		name            string
		input           *client.Monitor
		checkPort       bool
		wantPort        int64
		checkKeyword    bool
		checkEscalation bool
	}{
		{
			name: "with port",
			input: &client.Monitor{
				UUID:               "mon_port",
				Name:               "With Port",
				URL:                "tcp://example.com",
				Protocol:           "TCP",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				Port:               &port443,
			},
			checkPort: true,
			wantPort:  443,
		},
		{
			name: "with custom port",
			input: &client.Monitor{
				UUID:               "mon_custom_port",
				Name:               "Custom Port",
				URL:                "tcp://example.com",
				Protocol:           "TCP",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				Port:               &port8080,
			},
			checkPort: true,
			wantPort:  8080,
		},
		{
			name: "with required keyword",
			input: &client.Monitor{
				UUID:               "mon_keyword",
				Name:               "With Keyword",
				URL:                "https://example.com",
				Protocol:           "HTTPS",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				RequiredKeyword:    &requiredKeyword,
			},
			checkKeyword: true,
		},
		{
			name: "with escalation policy",
			input: &client.Monitor{
				UUID:               "mon_ep",
				Name:               "With EP",
				URL:                "https://example.com",
				Protocol:           "HTTPS",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
				EscalationPolicy:   &escalationPolicy,
			},
			checkEscalation: true,
		},
		{
			name: "minimal (no optionals)",
			input: &client.Monitor{
				UUID:               "mon_minimal",
				Name:               "Minimal",
				URL:                "https://example.com",
				Protocol:           "HTTPS",
				HTTPMethod:         "GET",
				CheckFrequency:     60,
				ExpectedStatusCode: "200",
				FollowRedirects:    true,
			},
			checkPort:       false,
			checkKeyword:    false,
			checkEscalation: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			result := MapMonitorCommonFields(tt.input, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}

			if result.ID.ValueString() != tt.input.UUID {
				t.Errorf("ID = %s, want %s", result.ID.ValueString(), tt.input.UUID)
			}

			if tt.checkPort {
				if result.Port.IsNull() {
					t.Error("expected non-null port")
				} else if result.Port.ValueInt64() != tt.wantPort {
					t.Errorf("Port = %d, want %d", result.Port.ValueInt64(), tt.wantPort)
				}
			} else {
				if !result.Port.IsNull() {
					t.Error("expected null port")
				}
			}

			if tt.checkKeyword {
				if result.RequiredKeyword.IsNull() {
					t.Error("expected non-null required_keyword")
				}
			}

			if tt.checkEscalation {
				if result.EscalationPolicy.IsNull() {
					t.Error("expected non-null escalation_policy")
				}
			}
		})
	}
}

// Test MapHealthcheckCommonFields with all field combinations
func TestMapHealthcheckCommonFields_AllFields(t *testing.T) {
	periodValue := 60
	hcFull := &client.Healthcheck{
		UUID:             "hc_full",
		Name:             "Full Healthcheck",
		PingURL:          "https://ping.example.com/hc_full",
		Cron:             "*/5 * * * *",
		Timezone:         "America/New_York",
		PeriodValue:      &periodValue,
		PeriodType:       "seconds",
		GracePeriodValue: 120,
		GracePeriodType:  "seconds",
		IsPaused:         false,
		IsDown:           true,
		Period:           300,
		GracePeriod:      120,
		LastPing:         "2026-02-10T15:30:00Z",
		CreatedAt:        "2026-01-01T00:00:00Z",
		EscalationPolicy: &client.EscalationPolicyReference{UUID: "ep_test"},
	}

	result := MapHealthcheckCommonFields(hcFull)

	// Check all fields are populated correctly
	if result.ID.ValueString() != "hc_full" {
		t.Errorf("ID = %s, want hc_full", result.ID.ValueString())
	}
	if result.Name.ValueString() != "Full Healthcheck" {
		t.Errorf("Name = %s, want Full Healthcheck", result.Name.ValueString())
	}
	if result.Cron.IsNull() || result.Cron.ValueString() != "*/5 * * * *" {
		t.Error("Cron not set correctly")
	}
	if result.Timezone.IsNull() || result.Timezone.ValueString() != "America/New_York" {
		t.Error("Timezone not set correctly")
	}
	if result.PeriodValue.IsNull() || result.PeriodValue.ValueInt64() != 60 {
		t.Error("PeriodValue not set correctly")
	}
	if result.EscalationPolicy.IsNull() || result.EscalationPolicy.ValueString() != "ep_test" {
		t.Error("EscalationPolicy not set correctly")
	}
}

// Test MapOutageNestedObjects with various monitor states
func TestMapOutageNestedObjects_MonitorStates(t *testing.T) {
	tests := []struct {
		name            string
		outage          *client.Outage
		wantMonitorNull bool
		wantAckNull     bool
	}{
		{
			name: "full monitor and ack",
			outage: &client.Outage{
				UUID: "out_full",
				Monitor: client.MonitorReference{
					UUID:     "mon_123",
					Name:     "API Monitor",
					URL:      "https://api.example.com",
					Protocol: "HTTPS",
				},
				AcknowledgedBy: &client.AcknowledgedByUser{
					UUID:  "user_1",
					Email: "ops@example.com",
					Name:  "Ops Team",
				},
			},
			wantMonitorNull: false,
			wantAckNull:     false,
		},
		{
			name: "monitor only",
			outage: &client.Outage{
				UUID: "out_mon_only",
				Monitor: client.MonitorReference{
					UUID:     "mon_456",
					Name:     "DB Monitor",
					URL:      "postgres://db.example.com",
					Protocol: "TCP",
				},
			},
			wantMonitorNull: false,
			wantAckNull:     true,
		},
		{
			name: "partial monitor (just uuid and name)",
			outage: &client.Outage{
				UUID: "out_partial",
				Monitor: client.MonitorReference{
					UUID: "mon_789",
					Name: "Partial",
				},
			},
			wantMonitorNull: false,
			wantAckNull:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var diags diag.Diagnostics
			monitorObj, ackObj := MapOutageNestedObjects(tt.outage, &diags)

			if diags.HasError() {
				t.Fatalf("unexpected error: %v", diags.Errors())
			}

			if monitorObj.IsNull() != tt.wantMonitorNull {
				t.Errorf("monitorObj.IsNull() = %v, want %v", monitorObj.IsNull(), tt.wantMonitorNull)
			}

			if ackObj.IsNull() != tt.wantAckNull {
				t.Errorf("ackObj.IsNull() = %v, want %v", ackObj.IsNull(), tt.wantAckNull)
			}

			if !tt.wantMonitorNull {
				attrs := monitorObj.Attributes()
				if attrs["uuid"] == nil {
					t.Error("monitor object missing uuid attribute")
				}
			}
		})
	}
}

// Test mapTFListToRequestHeaders handles edge cases
func TestMapTFListToRequestHeaders_EdgeCases(t *testing.T) {
	t.Run("null list", func(t *testing.T) {
		var diags diag.Diagnostics
		result := mapTFListToRequestHeaders(types.ListNull(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}), &diags)

		if result != nil {
			t.Error("expected nil for null list")
		}
	})

	t.Run("unknown list", func(t *testing.T) {
		var diags diag.Diagnostics
		result := mapTFListToRequestHeaders(types.ListUnknown(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}), &diags)

		if result != nil {
			t.Error("expected nil for unknown list")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		var diags diag.Diagnostics
		emptyList, _ := types.ListValue(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}, []attr.Value{})
		result := mapTFListToRequestHeaders(emptyList, &diags)

		if len(result) != 0 {
			t.Errorf("expected empty slice, got %d elements", len(result))
		}
	})

	t.Run("list with null values in header", func(t *testing.T) {
		var diags diag.Diagnostics

		headerWithNulls, _ := types.ObjectValue(RequestHeaderAttrTypes(), map[string]attr.Value{
			"name":  types.StringNull(),
			"value": types.StringNull(),
		})

		headerList, _ := types.ListValue(types.ObjectType{AttrTypes: RequestHeaderAttrTypes()}, []attr.Value{headerWithNulls})
		result := mapTFListToRequestHeaders(headerList, &diags)

		// Headers with null name/value should be skipped
		if len(result) != 0 {
			t.Errorf("expected 0 headers (null values skipped), got %d", len(result))
		}
	})
}
