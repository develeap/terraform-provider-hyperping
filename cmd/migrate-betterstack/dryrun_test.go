// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
	"github.com/develeap/terraform-provider-hyperping/pkg/dryrun"
)

func TestMonitorBridge(t *testing.T) {
	source := betterstack.Monitor{
		ID:   "123",
		Type: "monitor",
		Attributes: betterstack.MonitorAttributes{
			PronouncableName:    "Test Monitor",
			URL:                 "https://example.com",
			MonitorType:         "status",
			CheckFrequency:      30,
			RequestTimeout:      10,
			RequestMethod:       "GET",
			ExpectedStatusCodes: []int{200},
			Regions:             []string{"us", "eu"},
			FollowRedirects:     true,
			Paused:              false,
		},
	}

	converted := converter.ConvertedMonitor{
		ResourceName:       "test_monitor",
		Name:               "Test Monitor",
		URL:                "https://example.com",
		Protocol:           "http",
		HTTPMethod:         "GET",
		CheckFrequency:     30,
		Regions:            []string{"virginia", "london"},
		ExpectedStatusCode: "200",
		FollowRedirects:    true,
		Paused:             false,
		Issues:             []string{},
	}

	bridge := &monitorBridge{
		source:    source,
		converted: converted,
	}

	if bridge.GetResourceName() != "test_monitor" {
		t.Errorf("expected resource name test_monitor, got %s", bridge.GetResourceName())
	}

	if bridge.GetResourceType() != "hyperping_monitor" {
		t.Errorf("expected resource type hyperping_monitor, got %s", bridge.GetResourceType())
	}

	sourceData := bridge.GetSourceData()
	if sourceData["url"] != "https://example.com" {
		t.Error("source data missing URL")
	}

	targetData := bridge.GetTargetData()
	if targetData["url"] != "https://example.com" {
		t.Error("target data missing URL")
	}
}

func TestHealthcheckBridge(t *testing.T) {
	source := betterstack.Heartbeat{
		ID:   "456",
		Type: "heartbeat",
		Attributes: betterstack.HeartbeatAttributes{
			Name:   "Test Heartbeat",
			Period: 300,
			Grace:  60,
			Paused: false,
		},
	}

	converted := converter.ConvertedHealthcheck{
		ResourceName: "test_heartbeat",
		Name:         "Test Heartbeat",
		Period:       300,
		Grace:        60,
		Paused:       false,
		Issues:       []string{},
	}

	bridge := &healthcheckBridge{
		source:    source,
		converted: converted,
	}

	if bridge.GetResourceName() != "test_heartbeat" {
		t.Errorf("expected resource name test_heartbeat, got %s", bridge.GetResourceName())
	}

	if bridge.GetResourceType() != "hyperping_healthcheck" {
		t.Errorf("expected resource type hyperping_healthcheck, got %s", bridge.GetResourceType())
	}
}

func TestBuildDryRunReport(t *testing.T) {
	monitors := []betterstack.Monitor{
		{
			ID:   "123",
			Type: "monitor",
			Attributes: betterstack.MonitorAttributes{
				PronouncableName:    "Test Monitor",
				URL:                 "https://example.com",
				MonitorType:         "status",
				CheckFrequency:      30,
				ExpectedStatusCodes: []int{200},
			},
		},
	}

	heartbeats := []betterstack.Heartbeat{
		{
			ID:   "456",
			Type: "heartbeat",
			Attributes: betterstack.HeartbeatAttributes{
				Name:   "Test Heartbeat",
				Period: 300,
				Grace:  60,
			},
		},
	}

	convertedMonitors := []converter.ConvertedMonitor{
		{
			ResourceName:   "test_monitor",
			Name:           "Test Monitor",
			URL:            "https://example.com",
			Protocol:       "http",
			CheckFrequency: 30,
			Issues:         []string{},
		},
	}

	convertedHealthchecks := []converter.ConvertedHealthcheck{
		{
			ResourceName: "test_heartbeat",
			Name:         "Test Heartbeat",
			Period:       300,
			Grace:        60,
			Issues:       []string{},
		},
	}

	tfConfig := `resource "hyperping_monitor" "test_monitor" {
  name = "Test Monitor"
  url  = "https://example.com"
}`

	report := buildDryRunReport(
		monitors,
		heartbeats,
		convertedMonitors,
		convertedHealthchecks,
		[]converter.ConversionIssue{},
		[]converter.ConversionIssue{},
		tfConfig,
		"import script",
		"manual steps",
	)

	if report.SourcePlatform != "Better Stack" {
		t.Errorf("expected source platform Better Stack, got %s", report.SourcePlatform)
	}

	if report.TargetPlatform != "Hyperping" {
		t.Errorf("expected target platform Hyperping, got %s", report.TargetPlatform)
	}

	if report.Summary.TotalMonitors != 1 {
		t.Errorf("expected 1 monitor, got %d", report.Summary.TotalMonitors)
	}

	if report.Summary.TotalHealthchecks != 1 {
		t.Errorf("expected 1 healthcheck, got %d", report.Summary.TotalHealthchecks)
	}

	if len(report.Comparison) != 2 {
		t.Errorf("expected 2 comparisons, got %d", len(report.Comparison))
	}

	if report.Compatibility.OverallScore != 100.0 {
		t.Errorf("expected 100%% compatibility, got %.1f%%", report.Compatibility.OverallScore)
	}
}

func TestConvertIssuesToWarnings(t *testing.T) {
	monitorIssues := []converter.ConversionIssue{
		{
			ResourceName: "test",
			ResourceType: "monitor",
			Severity:     "warning",
			Message:      "Test warning",
		},
	}

	healthcheckIssues := []converter.ConversionIssue{
		{
			ResourceName: "test2",
			ResourceType: "healthcheck",
			Severity:     "error",
			Message:      "Test error",
		},
	}

	warnings := convertIssuesToWarnings(monitorIssues, healthcheckIssues)

	if len(warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d", len(warnings))
	}

	if warnings[0].Severity != "warning" {
		t.Errorf("expected warning severity, got %s", warnings[0].Severity)
	}

	if warnings[1].Severity != "critical" {
		t.Errorf("expected critical severity for error, got %s", warnings[1].Severity)
	}
}

func TestCalculateFrequencyDistribution(t *testing.T) {
	monitors := []converter.ConvertedMonitor{
		{CheckFrequency: 30},
		{CheckFrequency: 30},
		{CheckFrequency: 60},
		{CheckFrequency: 120},
	}

	dist := calculateFrequencyDistribution(monitors)

	if dist[30] != 2 {
		t.Errorf("expected 2 monitors at 30s, got %d", dist[30])
	}

	if dist[60] != 1 {
		t.Errorf("expected 1 monitor at 60s, got %d", dist[60])
	}

	if dist[120] != 1 {
		t.Errorf("expected 1 monitor at 120s, got %d", dist[120])
	}
}

func TestCalculateRegionDistribution(t *testing.T) {
	monitors := []converter.ConvertedMonitor{
		{Regions: []string{"london", "virginia"}},
		{Regions: []string{"london", "tokyo"}},
		{Regions: []string{"virginia"}},
	}

	dist := calculateRegionDistribution(monitors)

	if dist["london"] != 2 {
		t.Errorf("expected london in 2 monitors, got %d", dist["london"])
	}

	if dist["virginia"] != 2 {
		t.Errorf("expected virginia in 2 monitors, got %d", dist["virginia"])
	}

	if dist["tokyo"] != 1 {
		t.Errorf("expected tokyo in 1 monitor, got %d", dist["tokyo"])
	}
}

func TestPrintEnhancedDryRun(_ *testing.T) {
	report := dryrun.Report{
		Summary: dryrun.Summary{
			TotalMonitors:       5,
			TotalHealthchecks:   2,
			ExpectedTFResources: 7,
		},
		Compatibility: dryrun.CompatibilityScore{
			OverallScore:    90.0,
			Complexity:      "Simple",
			CleanMigrations: 7,
			WarningCount:    0,
			ErrorCount:      0,
		},
		SourcePlatform: "Better Stack",
		TargetPlatform: "Hyperping",
	}

	// Test that it doesn't panic
	// We're not capturing output, just ensuring it runs
	printEnhancedDryRun(report, false, false)
}

func TestBridgeGetIssues(t *testing.T) {
	converted := converter.ConvertedMonitor{
		Issues: []string{"warning 1", "warning 2"},
	}

	bridge := &monitorBridge{
		converted: converted,
	}

	issues := bridge.GetIssues()
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}

	if !strings.Contains(issues[0], "warning 1") {
		t.Error("expected first warning in issues")
	}
}
