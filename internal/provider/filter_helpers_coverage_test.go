// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// --- Schema function tests ---

func TestNameFilterSchema(t *testing.T) {
	s := NameFilterSchema()

	assertSchemaIsOptional(t, s)
	assertSchemaDescription(t, s, "Filter criteria for resources")

	expectedAttrs := []string{"name_regex"}
	assertSchemaAttributeNames(t, s, expectedAttrs)
	assertStringAttrOptional(t, s, "name_regex")
}

func TestMonitorFilterSchema(t *testing.T) {
	s := MonitorFilterSchema()

	assertSchemaIsOptional(t, s)
	assertSchemaDescription(t, s, "Filter criteria for monitors")

	expectedAttrs := []string{"name_regex", "protocol", "paused", "status", "project_uuid"}
	assertSchemaAttributeNames(t, s, expectedAttrs)

	assertStringAttrOptional(t, s, "name_regex")
	assertStringAttrOptional(t, s, "protocol")
	assertStringAttrOptional(t, s, "status")
	assertStringAttrOptional(t, s, "project_uuid")

	// paused is a BoolAttribute
	pausedAttr, ok := s.Attributes["paused"]
	if !ok {
		t.Fatal("MonitorFilterSchema() missing attribute: paused")
	}
	boolAttr, ok := pausedAttr.(schema.BoolAttribute)
	if !ok {
		t.Fatal("MonitorFilterSchema() 'paused' is not a BoolAttribute")
	}
	if !boolAttr.Optional {
		t.Error("MonitorFilterSchema() 'paused' should be optional")
	}

	// status should have validators
	statusAttr, ok := s.Attributes["status"]
	if !ok {
		t.Fatal("MonitorFilterSchema() missing attribute: status")
	}
	statusStr, ok := statusAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("MonitorFilterSchema() 'status' is not a StringAttribute")
	}
	if len(statusStr.Validators) == 0 {
		t.Error("MonitorFilterSchema() 'status' should have validators")
	}
}

func TestIncidentFilterSchema(t *testing.T) {
	s := IncidentFilterSchema()

	assertSchemaIsOptional(t, s)
	assertSchemaDescription(t, s, "Filter criteria for incidents")

	expectedAttrs := []string{"name_regex", "status", "severity"}
	assertSchemaAttributeNames(t, s, expectedAttrs)

	assertStringAttrOptional(t, s, "name_regex")
	assertStringAttrOptional(t, s, "status")
	assertStringAttrOptional(t, s, "severity")
}

func TestMaintenanceFilterSchema(t *testing.T) {
	s := MaintenanceFilterSchema()

	assertSchemaIsOptional(t, s)
	assertSchemaDescription(t, s, "Filter criteria for maintenance windows")

	expectedAttrs := []string{"name_regex", "status"}
	assertSchemaAttributeNames(t, s, expectedAttrs)

	assertStringAttrOptional(t, s, "name_regex")
	assertStringAttrOptional(t, s, "status")
}

func TestHealthcheckFilterSchema(t *testing.T) {
	s := HealthcheckFilterSchema()

	assertSchemaIsOptional(t, s)
	assertSchemaDescription(t, s, "Filter criteria for healthchecks")

	expectedAttrs := []string{"name_regex", "status"}
	assertSchemaAttributeNames(t, s, expectedAttrs)

	assertStringAttrOptional(t, s, "name_regex")
	assertStringAttrOptional(t, s, "status")
}

func TestOutageFilterSchema(t *testing.T) {
	s := OutageFilterSchema()

	assertSchemaIsOptional(t, s)
	assertSchemaDescription(t, s, "Filter criteria for outages")

	expectedAttrs := []string{"name_regex", "monitor_uuid"}
	assertSchemaAttributeNames(t, s, expectedAttrs)

	assertStringAttrOptional(t, s, "name_regex")
	assertStringAttrOptional(t, s, "monitor_uuid")
}

func TestStatusPageFilterSchema(t *testing.T) {
	s := StatusPageFilterSchema()

	assertSchemaIsOptional(t, s)
	assertSchemaDescription(t, s, "Filter criteria for status pages")

	expectedAttrs := []string{"name_regex", "hostname"}
	assertSchemaAttributeNames(t, s, expectedAttrs)

	assertStringAttrOptional(t, s, "name_regex")
	assertStringAttrOptional(t, s, "hostname")
}

// --- Table-driven schema attribute tests ---

func TestSchemaFunctions_AttributeCounts(t *testing.T) {
	tests := []struct {
		name          string
		schemaFn      func() schema.SingleNestedAttribute
		expectedCount int
	}{
		{
			name:          "NameFilterSchema has 1 attribute",
			schemaFn:      NameFilterSchema,
			expectedCount: 1,
		},
		{
			name:          "MonitorFilterSchema has 5 attributes",
			schemaFn:      MonitorFilterSchema,
			expectedCount: 5,
		},
		{
			name:          "IncidentFilterSchema has 3 attributes",
			schemaFn:      IncidentFilterSchema,
			expectedCount: 3,
		},
		{
			name:          "MaintenanceFilterSchema has 2 attributes",
			schemaFn:      MaintenanceFilterSchema,
			expectedCount: 2,
		},
		{
			name:          "HealthcheckFilterSchema has 2 attributes",
			schemaFn:      HealthcheckFilterSchema,
			expectedCount: 2,
		},
		{
			name:          "OutageFilterSchema has 2 attributes",
			schemaFn:      OutageFilterSchema,
			expectedCount: 2,
		},
		{
			name:          "StatusPageFilterSchema has 2 attributes",
			schemaFn:      StatusPageFilterSchema,
			expectedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.schemaFn()
			got := len(s.Attributes)
			if got != tt.expectedCount {
				t.Errorf("attribute count = %d, want %d", got, tt.expectedCount)
			}
		})
	}
}

func TestSchemaFunctions_AllOptional(t *testing.T) {
	tests := []struct {
		name     string
		schemaFn func() schema.SingleNestedAttribute
	}{
		{"NameFilterSchema", NameFilterSchema},
		{"MonitorFilterSchema", MonitorFilterSchema},
		{"IncidentFilterSchema", IncidentFilterSchema},
		{"MaintenanceFilterSchema", MaintenanceFilterSchema},
		{"HealthcheckFilterSchema", HealthcheckFilterSchema},
		{"OutageFilterSchema", OutageFilterSchema},
		{"StatusPageFilterSchema", StatusPageFilterSchema},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.schemaFn()
			if !s.Optional {
				t.Errorf("%s() should be optional", tt.name)
			}
		})
	}
}

func TestSchemaFunctions_AllHaveNameRegex(t *testing.T) {
	tests := []struct {
		name     string
		schemaFn func() schema.SingleNestedAttribute
	}{
		{"NameFilterSchema", NameFilterSchema},
		{"MonitorFilterSchema", MonitorFilterSchema},
		{"IncidentFilterSchema", IncidentFilterSchema},
		{"MaintenanceFilterSchema", MaintenanceFilterSchema},
		{"HealthcheckFilterSchema", HealthcheckFilterSchema},
		{"OutageFilterSchema", OutageFilterSchema},
		{"StatusPageFilterSchema", StatusPageFilterSchema},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.schemaFn()
			attr, ok := s.Attributes["name_regex"]
			if !ok {
				t.Fatalf("%s() missing name_regex attribute", tt.name)
			}
			strAttr, ok := attr.(schema.StringAttribute)
			if !ok {
				t.Fatalf("%s() name_regex is not a StringAttribute", tt.name)
			}
			if !strAttr.Optional {
				t.Errorf("%s() name_regex should be optional", tt.name)
			}
		})
	}
}

func TestMonitorFilterSchema_StatusValidation(t *testing.T) {
	s := MonitorFilterSchema()

	statusAttr, ok := s.Attributes["status"]
	if !ok {
		t.Fatal("MonitorFilterSchema() missing attribute: status")
	}
	strAttr, ok := statusAttr.(schema.StringAttribute)
	if !ok {
		t.Fatal("MonitorFilterSchema() status is not a StringAttribute")
	}
	if strAttr.MarkdownDescription == "" {
		t.Error("MonitorFilterSchema() status should have MarkdownDescription")
	}
	if len(strAttr.Validators) != 1 {
		t.Errorf("MonitorFilterSchema() status validator count = %d, want 1", len(strAttr.Validators))
	}
}

func TestMonitorFilterSchema_ProjectUUID(t *testing.T) {
	s := MonitorFilterSchema()

	attr, ok := s.Attributes["project_uuid"]
	if !ok {
		t.Fatal("MonitorFilterSchema() missing attribute: project_uuid")
	}
	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("MonitorFilterSchema() project_uuid is not a StringAttribute")
	}
	if !strAttr.Optional {
		t.Error("MonitorFilterSchema() project_uuid should be optional")
	}
	if strAttr.MarkdownDescription == "" {
		t.Error("MonitorFilterSchema() project_uuid should have MarkdownDescription")
	}
}

func TestIncidentFilterSchema_Descriptions(t *testing.T) {
	s := IncidentFilterSchema()

	tests := []struct {
		attrName string
	}{
		{"name_regex"},
		{"status"},
		{"severity"},
	}

	for _, tt := range tests {
		t.Run(tt.attrName, func(t *testing.T) {
			attr, ok := s.Attributes[tt.attrName]
			if !ok {
				t.Fatalf("missing attribute: %s", tt.attrName)
			}
			strAttr, ok := attr.(schema.StringAttribute)
			if !ok {
				t.Fatalf("attribute %s is not a StringAttribute", tt.attrName)
			}
			if strAttr.Description == "" {
				t.Errorf("attribute %s should have a description", tt.attrName)
			}
		})
	}
}

func TestOutageFilterSchema_MonitorUUID(t *testing.T) {
	s := OutageFilterSchema()

	attr, ok := s.Attributes["monitor_uuid"]
	if !ok {
		t.Fatal("OutageFilterSchema() missing attribute: monitor_uuid")
	}
	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("OutageFilterSchema() monitor_uuid is not a StringAttribute")
	}
	if !strAttr.Optional {
		t.Error("OutageFilterSchema() monitor_uuid should be optional")
	}
	if strAttr.Description == "" {
		t.Error("OutageFilterSchema() monitor_uuid should have a description")
	}
}

func TestStatusPageFilterSchema_Hostname(t *testing.T) {
	s := StatusPageFilterSchema()

	attr, ok := s.Attributes["hostname"]
	if !ok {
		t.Fatal("StatusPageFilterSchema() missing attribute: hostname")
	}
	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatal("StatusPageFilterSchema() hostname is not a StringAttribute")
	}
	if !strAttr.Optional {
		t.Error("StatusPageFilterSchema() hostname should be optional")
	}
	if strAttr.Description == "" {
		t.Error("StatusPageFilterSchema() hostname should have a description")
	}
}

// --- Test helpers ---

func assertSchemaIsOptional(t *testing.T, s schema.SingleNestedAttribute) {
	t.Helper()
	if !s.Optional {
		t.Error("schema should be optional")
	}
}

func assertSchemaDescription(t *testing.T, s schema.SingleNestedAttribute, expected string) {
	t.Helper()
	if s.Description != expected {
		t.Errorf("schema description = %q, want %q", s.Description, expected)
	}
}

func assertSchemaAttributeNames(t *testing.T, s schema.SingleNestedAttribute, expectedNames []string) {
	t.Helper()

	if len(s.Attributes) != len(expectedNames) {
		t.Errorf("attribute count = %d, want %d", len(s.Attributes), len(expectedNames))
	}

	for _, name := range expectedNames {
		if _, ok := s.Attributes[name]; !ok {
			t.Errorf("missing expected attribute: %s", name)
		}
	}
}

func assertStringAttrOptional(t *testing.T, s schema.SingleNestedAttribute, attrName string) {
	t.Helper()

	attr, ok := s.Attributes[attrName]
	if !ok {
		t.Fatalf("missing attribute: %s", attrName)
	}

	strAttr, ok := attr.(schema.StringAttribute)
	if !ok {
		t.Fatalf("attribute %s is not a StringAttribute", attrName)
	}

	if !strAttr.Optional {
		t.Errorf("attribute %s should be optional", attrName)
	}
}
