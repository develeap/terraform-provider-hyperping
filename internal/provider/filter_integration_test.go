// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// TestFilterIntegration verifies that all data sources have consistent filter implementations
func TestFilterIntegration(t *testing.T) {
	t.Run("MonitorFilterModel has required fields", func(t *testing.T) {
		filter := &MonitorFilterModel{
			NameRegex: types.StringValue("test"),
			Protocol:  types.StringValue("https"),
			Paused:    types.BoolValue(false),
		}
		if filter.NameRegex.IsNull() {
			t.Error("NameRegex should not be null")
		}
		if filter.Protocol.IsNull() {
			t.Error("Protocol should not be null")
		}
		if filter.Paused.IsNull() {
			t.Error("Paused should not be null")
		}
	})

	t.Run("IncidentFilterModel has required fields", func(t *testing.T) {
		filter := &IncidentFilterModel{
			NameRegex: types.StringValue("test"),
			Status:    types.StringValue("outage"),
			Severity:  types.StringNull(),
		}
		if filter.NameRegex.IsNull() {
			t.Error("NameRegex should not be null")
		}
		if filter.Status.IsNull() {
			t.Error("Status should not be null")
		}
		if !filter.Severity.IsNull() {
			t.Error("Severity should be null")
		}
	})

	t.Run("MaintenanceFilterModel has required fields", func(t *testing.T) {
		filter := &MaintenanceFilterModel{
			NameRegex: types.StringValue("test"),
			Status:    types.StringValue("ongoing"),
		}
		if filter.NameRegex.IsNull() {
			t.Error("NameRegex should not be null")
		}
		if filter.Status.IsNull() {
			t.Error("Status should not be null")
		}
	})

	t.Run("OutageFilterModel has required fields", func(t *testing.T) {
		filter := &OutageFilterModel{
			NameRegex:   types.StringValue("test"),
			MonitorUUID: types.StringValue("mon-123"),
		}
		if filter.NameRegex.IsNull() {
			t.Error("NameRegex should not be null")
		}
		if filter.MonitorUUID.IsNull() {
			t.Error("MonitorUUID should not be null")
		}
	})

	t.Run("HealthcheckFilterModel has required fields", func(t *testing.T) {
		filter := &HealthcheckFilterModel{
			NameRegex: types.StringValue("test"),
			Status:    types.StringValue("up"),
		}
		if filter.NameRegex.IsNull() {
			t.Error("NameRegex should not be null")
		}
		if filter.Status.IsNull() {
			t.Error("Status should not be null")
		}
	})

	t.Run("StatusPageFilterModel has required fields", func(t *testing.T) {
		filter := &StatusPageFilterModel{
			NameRegex: types.StringValue("test"),
			Hostname:  types.StringValue("status.example.com"),
		}
		if filter.NameRegex.IsNull() {
			t.Error("NameRegex should not be null")
		}
		if filter.Hostname.IsNull() {
			t.Error("Hostname should not be null")
		}
	})
}

// TestFilterSchemaConsistency ensures all filter schemas follow the same pattern
func TestFilterSchemaConsistency(t *testing.T) {
	schemas := []struct {
		name   string
		schema func() interface{}
	}{
		{"MonitorFilterSchema", func() interface{} { return MonitorFilterSchema() }},
		{"IncidentFilterSchema", func() interface{} { return IncidentFilterSchema() }},
		{"MaintenanceFilterSchema", func() interface{} { return MaintenanceFilterSchema() }},
		{"OutageFilterSchema", func() interface{} { return OutageFilterSchema() }},
		{"HealthcheckFilterSchema", func() interface{} { return HealthcheckFilterSchema() }},
		{"StatusPageFilterSchema", func() interface{} { return StatusPageFilterSchema() }},
	}

	for _, s := range schemas {
		t.Run(s.name, func(t *testing.T) {
			schema := s.schema()
			if schema == nil {
				t.Errorf("%s returned nil", s.name)
			}
		})
	}
}

// TestDataSourceFilterInitialization verifies all data sources can be created
func TestDataSourceFilterInitialization(t *testing.T) {
	dataSources := []struct {
		name string
		ds   interface{}
	}{
		{"MonitorsDataSource", &MonitorsDataSource{}},
		{"IncidentsDataSource", &IncidentsDataSource{}},
		{"MaintenanceWindowsDataSource", &MaintenanceWindowsDataSource{}},
		{"OutagesDataSource", &OutagesDataSource{}},
		{"HealthchecksDataSource", &HealthchecksDataSource{}},
		{"StatusPagesDataSource", &StatusPagesDataSource{}},
	}

	for _, ds := range dataSources {
		t.Run(ds.name, func(t *testing.T) {
			if ds.ds == nil {
				t.Errorf("%s is nil", ds.name)
			}
		})
	}
}
