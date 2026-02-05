// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package contract

import (
	"testing"
)

func TestCompareWithDocumentation(t *testing.T) {
	// Create a cassette schema with some fields
	schema := NewCassetteSchema("testdata")
	endpoint := schema.GetOrCreateEndpoint("GET", "/v1/healthchecks/{id}")
	endpoint.ResponseFields["uuid"] = APIFieldSchema{Name: "uuid", Type: "string"}
	endpoint.ResponseFields["name"] = APIFieldSchema{Name: "name", Type: "string"}
	endpoint.ResponseFields["pingUrl"] = APIFieldSchema{Name: "pingUrl", Type: "string"}
	endpoint.ResponseFields["lastLogStartDate"] = APIFieldSchema{Name: "lastLogStartDate", Type: "string"}
	endpoint.ResponseFields["period"] = APIFieldSchema{Name: "period", Type: "integer"}

	// Documented fields (simulating what the API docs say)
	docFields := map[string][]DocumentedField{
		"healthchecks": {
			{Name: "uuid", Type: "string"},
			{Name: "name", Type: "string"},
			{Name: "period", Type: "number"},   // Type differs
			{Name: "oldField", Type: "string"}, // Deprecated
		},
	}

	results := CompareWithDocumentation(schema, docFields)

	if len(results) == 0 {
		t.Fatal("expected at least one comparison result")
	}

	result := results[0]
	if result.Resource != "healthchecks" {
		t.Errorf("expected resource 'healthchecks', got %q", result.Resource)
	}

	// Check discoveries
	var foundUndocumented, foundDiffers, foundDeprecated bool
	for _, d := range result.Discoveries {
		switch d.FieldName {
		case "pingUrl", "lastLogStartDate":
			if d.DocStatus != DocStatusUndocumented {
				t.Errorf("expected %s to be undocumented, got %s", d.FieldName, d.DocStatus)
			}
			foundUndocumented = true
		case "period":
			if d.DocStatus != DocStatusDiffers {
				t.Errorf("expected period to differ, got %s", d.DocStatus)
			}
			foundDiffers = true
		case "oldField":
			if d.DocStatus != DocStatusDeprecated {
				t.Errorf("expected oldField to be deprecated, got %s", d.DocStatus)
			}
			foundDeprecated = true
		}
	}

	if !foundUndocumented {
		t.Error("expected to find undocumented fields")
	}
	if !foundDiffers {
		t.Error("expected to find type mismatch")
	}
	if !foundDeprecated {
		t.Error("expected to find deprecated field")
	}

	// Check summary stats
	if result.Summary.UndocumentedFields < 2 {
		t.Errorf("expected at least 2 undocumented fields, got %d", result.Summary.UndocumentedFields)
	}
	if result.Summary.TypeMismatches != 1 {
		t.Errorf("expected 1 type mismatch, got %d", result.Summary.TypeMismatches)
	}
}

func TestExtractResourceFromPath(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/v1/healthchecks", "healthchecks"},
		{"/v1/healthchecks/{id}", "healthchecks"},
		{"/v1/monitors/{id}/pause", "monitors"},
		{"/v3/incidents", "incidents"},
		{"/v1/statuspages/{id}/subscribers", "statuspages"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := extractResourceFromPath(tt.path)
			if result != tt.expected {
				t.Errorf("extractResourceFromPath(%q) = %q, want %q", tt.path, result, tt.expected)
			}
		})
	}
}

func TestNormalizeTypeForComparison(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"string", "string"},
		{"String", "string"},
		{"int", "integer"},
		{"integer", "integer"},
		{"int64", "integer"},
		{"float", "number"},
		{"number", "number"},
		{"bool", "boolean"},
		{"boolean", "boolean"},
		{"array", "array"},
		{"list", "array"},
		{"object", "object"},
		{"map", "object"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeTypeForComparison(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeTypeForComparison(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateDiscoveryReport(t *testing.T) {
	results := []ComparisonResult{
		{
			Resource: "healthchecks",
			Endpoint: "GET /v1/healthchecks/{id}",
			Discoveries: []FieldDiscovery{
				{
					FieldName:  "lastLogStartDate",
					Type:       "string",
					DocStatus:  DocStatusUndocumented,
					Suggestion: "Add 'lastLogStartDate' (string) to API documentation",
				},
				{
					FieldName:  "period",
					Type:       "integer",
					DocStatus:  DocStatusDiffers,
					DocType:    "number",
					Suggestion: "API returns integer but docs say number",
				},
			},
			Summary: ComparisonStats{
				UndocumentedFields: 1,
				TypeMismatches:     1,
			},
		},
	}

	report := GenerateDiscoveryReport(results)

	// Check report contains expected sections
	if !containsString(report, "# API Documentation Gaps") {
		t.Error("expected report header")
	}
	if !containsString(report, "## healthchecks") {
		t.Error("expected resource section")
	}
	if !containsString(report, "lastLogStartDate") {
		t.Error("expected undocumented field in report")
	}
	if !containsString(report, "Type Mismatches") {
		t.Error("expected type mismatch section")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
