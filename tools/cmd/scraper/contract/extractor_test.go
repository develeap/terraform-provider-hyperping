// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package contract

import (
	"testing"
)

func TestExtractFromCassettes(t *testing.T) {
	schema, err := ExtractFromCassettes("testdata")
	if err != nil {
		t.Fatalf("ExtractFromCassettes failed: %v", err)
	}

	// Should have extracted endpoints
	if len(schema.Endpoints) == 0 {
		t.Fatal("expected at least one endpoint")
	}

	// Check POST /v1/healthchecks endpoint
	postEndpoint := schema.Endpoints["POST /v1/healthchecks"]
	if postEndpoint == nil {
		t.Fatal("expected POST /v1/healthchecks endpoint")
	}

	// Verify request fields
	expectedReqFields := []string{"name", "periodValue", "periodType", "gracePeriodValue", "gracePeriodType"}
	for _, field := range expectedReqFields {
		if _, exists := postEndpoint.RequestFields[field]; !exists {
			t.Errorf("expected request field %q", field)
		}
	}

	// Verify response fields
	expectedRespFields := []string{"uuid", "name", "pingUrl", "period", "gracePeriod", "isPaused", "isDown", "createdAt"}
	for _, field := range expectedRespFields {
		if _, exists := postEndpoint.ResponseFields[field]; !exists {
			t.Errorf("expected response field %q", field)
		}
	}

	// Check status codes
	if len(postEndpoint.StatusCodes) == 0 {
		t.Error("expected status codes to be tracked")
	}
	if !contains(postEndpoint.StatusCodes, 201) {
		t.Error("expected 201 in status codes")
	}

	// Check GET endpoint (should have {id} placeholder)
	getEndpoint := schema.Endpoints["GET /v1/healthchecks/{id}"]
	if getEndpoint == nil {
		t.Fatal("expected GET /v1/healthchecks/{id} endpoint")
	}

	// GET response should have nullable fields like lastPing
	if lastPing, exists := getEndpoint.ResponseFields["lastPing"]; exists {
		if !lastPing.Nullable {
			t.Error("expected lastPing to be nullable")
		}
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://api.hyperping.io/v1/healthchecks",
			expected: "/v1/healthchecks",
		},
		{
			input:    "https://api.hyperping.io/v1/healthchecks/hc_abc123",
			expected: "/v1/healthchecks/{id}",
		},
		{
			input:    "https://api.hyperping.io/v1/monitors/mon_xyz789",
			expected: "/v1/monitors/{id}",
		},
		{
			input:    "/v1/incidents?page=1",
			expected: "/v1/incidents",
		},
		{
			input:    "https://api.hyperping.io/v1/statuspages/sp_123abc/subscribers",
			expected: "/v1/statuspages/{id}/subscribers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizePath(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestInferType(t *testing.T) {
	tests := []struct {
		name      string
		value     any
		wantType  string
		wantChild string
		wantNull  bool
	}{
		{"nil", nil, "", "", true},
		{"string", "hello", "string", "", false},
		{"integer", float64(42), "integer", "", false},
		{"float", float64(3.14), "number", "", false},
		{"boolean", true, "boolean", "", false},
		{"empty array", []any{}, "array", "", false},
		{"string array", []any{"a", "b"}, "array", "string", false},
		{"object", map[string]any{"key": "value"}, "object", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotChild, gotNull := inferType(tt.value)
			if gotType != tt.wantType {
				t.Errorf("inferType() type = %q, want %q", gotType, tt.wantType)
			}
			if gotChild != tt.wantChild {
				t.Errorf("inferType() childType = %q, want %q", gotChild, tt.wantChild)
			}
			if gotNull != tt.wantNull {
				t.Errorf("inferType() isNull = %v, want %v", gotNull, tt.wantNull)
			}
		})
	}
}

func TestFieldSource(t *testing.T) {
	tests := []struct {
		source   FieldSource
		expected string
	}{
		{SourceRequest, "request"},
		{SourceResponse, "response"},
		{SourceBoth, "both"},
		{FieldSource(0), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.source.String(); got != tt.expected {
				t.Errorf("FieldSource.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExtractFields(t *testing.T) {
	obj := map[string]any{
		"uuid":      "abc123",
		"count":     float64(42),
		"enabled":   true,
		"price":     float64(19.99),
		"nullField": nil,
		"nested": map[string]any{
			"inner": "value",
		},
		"items": []any{
			map[string]any{"id": float64(1), "name": "first"},
		},
	}

	fields := make(map[string]APIFieldSchema)
	extractFields(obj, fields, SourceResponse, "")

	// Check basic types
	if fields["uuid"].Type != "string" {
		t.Errorf("uuid type = %q, want string", fields["uuid"].Type)
	}
	if fields["count"].Type != "integer" {
		t.Errorf("count type = %q, want integer", fields["count"].Type)
	}
	if fields["enabled"].Type != "boolean" {
		t.Errorf("enabled type = %q, want boolean", fields["enabled"].Type)
	}
	if fields["price"].Type != "number" {
		t.Errorf("price type = %q, want number", fields["price"].Type)
	}
	if !fields["nullField"].Nullable {
		t.Error("nullField should be nullable")
	}

	// Check nested object
	if fields["nested"].Type != "object" {
		t.Errorf("nested type = %q, want object", fields["nested"].Type)
	}
	if fields["nested.inner"].Type != "string" {
		t.Errorf("nested.inner type = %q, want string", fields["nested.inner"].Type)
	}

	// Check array of objects
	if fields["items"].Type != "array" {
		t.Errorf("items type = %q, want array", fields["items"].Type)
	}
	if fields["items"].ChildType != "object" {
		t.Errorf("items childType = %q, want object", fields["items"].ChildType)
	}
	if fields["items[].id"].Type != "integer" {
		t.Errorf("items[].id type = %q, want integer", fields["items[].id"].Type)
	}
}
