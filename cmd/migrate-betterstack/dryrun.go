// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
	"github.com/develeap/terraform-provider-hyperping/pkg/dryrun"
)

// monitorBridge adapts Better Stack monitors to dryrun interface.
type monitorBridge struct {
	source    betterstack.Monitor
	converted converter.ConvertedMonitor
}

func (m *monitorBridge) GetResourceName() string {
	return m.converted.ResourceName
}

func (m *monitorBridge) GetResourceType() string {
	return "hyperping_monitor"
}

func (m *monitorBridge) GetIssues() []string {
	return m.converted.Issues
}

func (m *monitorBridge) GetSourceData() map[string]interface{} {
	return map[string]interface{}{
		"url":              m.source.Attributes.URL,
		"monitor_type":     m.source.Attributes.MonitorType,
		"check_frequency":  m.source.Attributes.CheckFrequency,
		"request_timeout":  m.source.Attributes.RequestTimeout,
		"request_method":   m.source.Attributes.RequestMethod,
		"expected_status":  m.source.Attributes.ExpectedStatusCodes,
		"regions":          m.source.Attributes.Regions,
		"follow_redirects": m.source.Attributes.FollowRedirects,
		"paused":           m.source.Attributes.Paused,
	}
}

func (m *monitorBridge) GetTargetData() map[string]interface{} {
	return map[string]interface{}{
		"url":              m.converted.URL,
		"protocol":         m.converted.Protocol,
		"frequency":        m.converted.CheckFrequency,
		"http_method":      m.converted.HTTPMethod,
		"expected_status":  m.converted.ExpectedStatusCode,
		"regions":          m.converted.Regions,
		"follow_redirects": m.converted.FollowRedirects,
		"paused":           m.converted.Paused,
	}
}

// healthcheckBridge adapts Better Stack heartbeats to dryrun interface.
type healthcheckBridge struct {
	source    betterstack.Heartbeat
	converted converter.ConvertedHealthcheck
}

func (h *healthcheckBridge) GetResourceName() string {
	return h.converted.ResourceName
}

func (h *healthcheckBridge) GetResourceType() string {
	return "hyperping_healthcheck"
}

func (h *healthcheckBridge) GetIssues() []string {
	return h.converted.Issues
}

func (h *healthcheckBridge) GetSourceData() map[string]interface{} {
	return map[string]interface{}{
		"name":   h.source.Attributes.Name,
		"period": h.source.Attributes.Period,
		"grace":  h.source.Attributes.Grace,
		"paused": h.source.Attributes.Paused,
	}
}

func (h *healthcheckBridge) GetTargetData() map[string]interface{} {
	return map[string]interface{}{
		"name":   h.converted.Name,
		"period": h.converted.Period,
		"grace":  h.converted.Grace,
		"paused": h.converted.Paused,
	}
}

// buildDryRunReport creates an enhanced dry-run report.
func buildDryRunReport(
	monitors []betterstack.Monitor,
	heartbeats []betterstack.Heartbeat,
	convertedMonitors []converter.ConvertedMonitor,
	convertedHealthchecks []converter.ConvertedHealthcheck,
	monitorIssues []converter.ConversionIssue,
	healthcheckIssues []converter.ConversionIssue,
	tfConfig string,
	importScript string,
	manualSteps string,
) dryrun.Report {
	// Build bridge converters
	var bridges []dryrun.BridgeConverter

	for i, m := range convertedMonitors {
		if i < len(monitors) {
			bridges = append(bridges, &monitorBridge{
				source:    monitors[i],
				converted: m,
			})
		}
	}

	for i, h := range convertedHealthchecks {
		if i < len(heartbeats) {
			bridges = append(bridges, &healthcheckBridge{
				source:    heartbeats[i],
				converted: h,
			})
		}
	}

	// Build comparisons
	comparisons := dryrun.BuildComparisons(bridges)

	// Convert issues to warnings
	warnings := convertIssuesToWarnings(monitorIssues, healthcheckIssues)

	// Build resource breakdown
	breakdown := map[string]int{
		"hyperping_monitor":     len(convertedMonitors),
		"hyperping_healthcheck": len(convertedHealthchecks),
	}

	// Build summary
	summary := dryrun.BuildSummary(
		len(monitors),
		len(heartbeats),
		tfConfig,
		breakdown,
	)

	// Calculate frequency distribution
	summary.FrequencyDistribution = calculateFrequencyDistribution(convertedMonitors)

	// Calculate region distribution
	summary.RegionDistribution = calculateRegionDistribution(convertedMonitors)

	// Estimate performance
	sourceAPICalls := 2 // monitors + heartbeats
	estimates := dryrun.EstimatePerformance(
		len(convertedMonitors)+len(convertedHealthchecks),
		sourceAPICalls,
		len(tfConfig),
	)
	estimates.ImportScriptSize = int64(len(importScript))
	estimates.ManualStepsSize = int64(len(manualSteps))

	// Build report
	reporter := dryrun.NewReporter(true)
	return reporter.GenerateReport(
		"Better Stack",
		comparisons,
		warnings,
		tfConfig,
		estimates,
		summary,
	)
}

func convertIssuesToWarnings(
	monitorIssues []converter.ConversionIssue,
	healthcheckIssues []converter.ConversionIssue,
) []dryrun.Warning {
	var warnings []dryrun.Warning

	allIssues := append(monitorIssues, healthcheckIssues...)

	for _, issue := range allIssues {
		severity := "warning"
		if issue.Severity == "error" {
			severity = "critical"
		}

		warnings = append(warnings, dryrun.Warning{
			Severity:     severity,
			ResourceName: issue.ResourceName,
			ResourceType: issue.ResourceType,
			Category:     "conversion",
			Message:      issue.Message,
			Action:       "Review and update manually after migration",
			Impact:       "May require manual configuration",
		})
	}

	return warnings
}

func calculateFrequencyDistribution(monitors []converter.ConvertedMonitor) map[int]int {
	dist := make(map[int]int)
	for _, m := range monitors {
		dist[m.CheckFrequency]++
	}
	return dist
}

func calculateRegionDistribution(monitors []converter.ConvertedMonitor) map[string]int {
	dist := make(map[string]int)
	for _, m := range monitors {
		for _, region := range m.Regions {
			dist[region]++
		}
	}
	return dist
}

// printEnhancedDryRun prints the enhanced dry-run report.
func printEnhancedDryRun(report dryrun.Report, verbose bool, jsonFormat bool) {
	reporter := dryrun.NewReporter(true)

	if jsonFormat {
		jsonOutput, err := reporter.FormatJSON(report)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error formatting JSON: %v\n", err)
			return
		}
		fmt.Println(jsonOutput)
		return
	}

	opts := dryrun.Options{
		Verbose:          verbose,
		ShowAllResources: verbose,
		PreviewLimit:     3,
		FormatJSON:       false,
		IncludeExamples:  true,
	}

	if err := reporter.PrintReport(os.Stderr, report, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error printing dry-run report: %v\n", err)
	}
}
