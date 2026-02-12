// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// generateTestMonitors creates n test monitors for performance testing.
func generateTestMonitors(n int) []client.Monitor {
	monitors := make([]client.Monitor, n)
	protocols := []string{"http", "https", "tcp", "icmp"}

	for i := 0; i < n; i++ {
		monitors[i] = client.Monitor{
			UUID:     client.MonitorUUID(i),
			Name:     client.MonitorName(i, i%4),
			URL:      "https://example.com",
			Protocol: protocols[i%len(protocols)],
			Paused:   i%3 == 0,
		}
	}

	return monitors
}

// generateTestIncidents creates n test incidents for performance testing.
func generateTestIncidents(n int) []client.Incident {
	incidents := make([]client.Incident, n)
	types := []string{"incident", "maintenance"}

	for i := 0; i < n; i++ {
		title := client.IncidentTitle(i)
		text := "Test incident text"
		incidents[i] = client.Incident{
			UUID:  client.IncidentID(i),
			Title: client.LocalizedText{En: title},
			Text:  client.LocalizedText{En: text},
			Type:  types[i%len(types)],
		}
	}

	return incidents
}

// generateTestMaintenanceWindows creates n test maintenance windows.
func generateTestMaintenanceWindows(n int) []client.Maintenance {
	windows := make([]client.Maintenance, n)

	for i := 0; i < n; i++ {
		title := client.MaintenanceTitle(i)
		text := "Test maintenance text"
		windows[i] = client.Maintenance{
			UUID:  client.MaintenanceID(i),
			Name:  client.MaintenanceTitle(i),
			Title: client.LocalizedText{En: title},
			Text:  client.LocalizedText{En: text},
		}
	}

	return windows
}

// TestFilterPerformance_1000Monitors tests filtering performance with 1000 monitors.
func TestFilterPerformance_1000Monitors(t *testing.T) {
	monitors := generateTestMonitors(1000)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		NameRegex: types.StringValue("\\[PROD\\]-.*"),
		Protocol:  types.StringValue("https"),
		Paused:    types.BoolValue(false),
	}

	start := time.Now()

	var filtered []client.Monitor
	for _, m := range monitors {
		monitor := m
		if ds.shouldIncludeMonitor(&monitor, filter, diags) {
			filtered = append(filtered, monitor)
		}
	}

	elapsed := time.Since(start)

	if diags.HasError() {
		t.Fatalf("Filter returned errors: %v", diags.Errors())
	}

	if elapsed > 5*time.Second {
		t.Errorf("Filtering took too long: %v (threshold: 5s)", elapsed)
	}

	t.Logf("Filtered %d/%d monitors in %v", len(filtered), len(monitors), elapsed)
}

// TestFilterPerformance_1000Incidents tests filtering performance with 1000 incidents.
func TestFilterPerformance_1000Incidents(t *testing.T) {
	incidents := generateTestIncidents(1000)
	ds := &IncidentsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &IncidentFilterModel{
		NameRegex: types.StringValue("API.*"),
		Status:    types.StringValue("resolved"),
		Severity:  types.StringValue("major"),
	}

	start := time.Now()

	var filtered []client.Incident
	for _, inc := range incidents {
		incident := inc
		if ds.shouldIncludeIncident(&incident, filter, diags) {
			filtered = append(filtered, incident)
		}
	}

	elapsed := time.Since(start)

	if diags.HasError() {
		t.Fatalf("Filter returned errors: %v", diags.Errors())
	}

	if elapsed > 5*time.Second {
		t.Errorf("Filtering took too long: %v (threshold: 5s)", elapsed)
	}

	t.Logf("Filtered %d/%d incidents in %v", len(filtered), len(incidents), elapsed)
}

// TestFilterPerformance_1000MaintenanceWindows tests filtering performance with 1000 maintenance windows.
func TestFilterPerformance_1000MaintenanceWindows(t *testing.T) {
	windows := generateTestMaintenanceWindows(1000)
	ds := &MaintenanceWindowsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MaintenanceFilterModel{
		NameRegex: types.StringValue("Database.*"),
		Status:    types.StringValue("scheduled"),
	}

	start := time.Now()

	var filtered []client.Maintenance
	for _, w := range windows {
		window := w
		if ds.shouldIncludeMaintenance(&window, filter, diags) {
			filtered = append(filtered, window)
		}
	}

	elapsed := time.Since(start)

	if diags.HasError() {
		t.Fatalf("Filter returned errors: %v", diags.Errors())
	}

	if elapsed > 5*time.Second {
		t.Errorf("Filtering took too long: %v (threshold: 5s)", elapsed)
	}

	t.Logf("Filtered %d/%d maintenance windows in %v", len(filtered), len(windows), elapsed)
}

// TestFilterPerformance_10000Monitors tests filtering with very large dataset.
func TestFilterPerformance_10000Monitors(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping large dataset test in short mode")
	}

	monitors := generateTestMonitors(10000)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		Protocol: types.StringValue("https"),
		Paused:   types.BoolValue(false),
	}

	start := time.Now()

	var filtered []client.Monitor
	for _, m := range monitors {
		monitor := m
		if ds.shouldIncludeMonitor(&monitor, filter, diags) {
			filtered = append(filtered, monitor)
		}
	}

	elapsed := time.Since(start)

	if diags.HasError() {
		t.Fatalf("Filter returned errors: %v", diags.Errors())
	}

	// More lenient threshold for very large datasets
	if elapsed > 30*time.Second {
		t.Errorf("Filtering took too long: %v (threshold: 30s)", elapsed)
	}

	t.Logf("Filtered %d/%d monitors in %v", len(filtered), len(monitors), elapsed)
}

// BenchmarkRegexFilter benchmarks regex filter compilation and matching.
func BenchmarkRegexFilter(b *testing.B) {
	monitors := generateTestMonitors(100)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		NameRegex: types.StringValue("\\[PROD\\]-API-.*"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, m := range monitors {
			monitor := m
			ds.shouldIncludeMonitor(&monitor, filter, diags)
		}
	}
}

// BenchmarkExactMatchFilter benchmarks exact string matching.
func BenchmarkExactMatchFilter(b *testing.B) {
	monitors := generateTestMonitors(100)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		Protocol: types.StringValue("https"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, m := range monitors {
			monitor := m
			ds.shouldIncludeMonitor(&monitor, filter, diags)
		}
	}
}

// BenchmarkBooleanFilter benchmarks boolean matching.
func BenchmarkBooleanFilter(b *testing.B) {
	monitors := generateTestMonitors(100)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		Paused: types.BoolValue(false),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, m := range monitors {
			monitor := m
			ds.shouldIncludeMonitor(&monitor, filter, diags)
		}
	}
}

// BenchmarkCombinedFilters benchmarks multiple filters together.
func BenchmarkCombinedFilters(b *testing.B) {
	monitors := generateTestMonitors(100)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		NameRegex: types.StringValue("\\[PROD\\]-.*"),
		Protocol:  types.StringValue("https"),
		Paused:    types.BoolValue(false),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, m := range monitors {
			monitor := m
			ds.shouldIncludeMonitor(&monitor, filter, diags)
		}
	}
}

// BenchmarkComplexRegex benchmarks complex regex patterns.
func BenchmarkComplexRegex(b *testing.B) {
	monitors := generateTestMonitors(100)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		NameRegex: types.StringValue("^\\[(PROD|STAGING|DEV)\\]-(API|WEB|DATABASE)-[A-Z]+(-[0-9]+)?$"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, m := range monitors {
			monitor := m
			ds.shouldIncludeMonitor(&monitor, filter, diags)
		}
	}
}

// TestFilterPerformance_emptyFilter tests that empty filter has minimal overhead.
func TestFilterPerformance_emptyFilter(t *testing.T) {
	monitors := generateTestMonitors(1000)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		// All null - no filtering
	}

	start := time.Now()

	var filtered []client.Monitor
	for _, m := range monitors {
		monitor := m
		if ds.shouldIncludeMonitor(&monitor, filter, diags) {
			filtered = append(filtered, monitor)
		}
	}

	elapsed := time.Since(start)

	if diags.HasError() {
		t.Fatalf("Filter returned errors: %v", diags.Errors())
	}

	// Empty filter should be very fast
	if elapsed > 1*time.Second {
		t.Errorf("Empty filter took too long: %v (threshold: 1s)", elapsed)
	}

	// Should return all monitors
	if len(filtered) != len(monitors) {
		t.Errorf("Empty filter returned %d monitors, expected %d", len(filtered), len(monitors))
	}

	t.Logf("Empty filter processed %d monitors in %v", len(monitors), elapsed)
}

// TestFilterPerformance_singleFieldFilter tests single field filtering performance.
func TestFilterPerformance_singleFieldFilter(t *testing.T) {
	monitors := generateTestMonitors(1000)
	ds := &MonitorsDataSource{}
	diags := &diag.Diagnostics{}

	filter := &MonitorFilterModel{
		Protocol: types.StringValue("https"),
	}

	start := time.Now()

	var filtered []client.Monitor
	for _, m := range monitors {
		monitor := m
		if ds.shouldIncludeMonitor(&monitor, filter, diags) {
			filtered = append(filtered, monitor)
		}
	}

	elapsed := time.Since(start)

	if diags.HasError() {
		t.Fatalf("Filter returned errors: %v", diags.Errors())
	}

	if elapsed > 2*time.Second {
		t.Errorf("Single field filter took too long: %v (threshold: 2s)", elapsed)
	}

	t.Logf("Single field filter processed %d monitors in %v, matched %d", len(monitors), elapsed, len(filtered))
}

// TestFilterPerformance_regexCompilation tests regex compilation overhead.
func TestFilterPerformance_regexCompilation(t *testing.T) {
	monitors := generateTestMonitors(100)
	ds := &MonitorsDataSource{}

	// Test that regex is compiled multiple times (inefficient pattern)
	patterns := []string{
		"\\[PROD\\]-.*",
		"\\[STAGING\\]-.*",
		"\\[DEV\\]-.*",
		".*-API$",
		".*-WEB$",
	}

	start := time.Now()

	for _, pattern := range patterns {
		diags := &diag.Diagnostics{}
		filter := &MonitorFilterModel{
			NameRegex: types.StringValue(pattern),
		}

		for _, m := range monitors {
			monitor := m
			ds.shouldIncludeMonitor(&monitor, filter, diags)
		}

		if diags.HasError() {
			t.Fatalf("Filter returned errors: %v", diags.Errors())
		}
	}

	elapsed := time.Since(start)

	if elapsed > 5*time.Second {
		t.Errorf("Regex compilation took too long: %v (threshold: 5s)", elapsed)
	}

	t.Logf("Compiled and matched %d patterns against %d monitors in %v", len(patterns), len(monitors), elapsed)
}
