// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/betterstack"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
)

// Report contains migration statistics and details.
type Report struct {
	Summary          Summary                     `json:"summary"`
	Monitors         []MonitorMapping            `json:"monitors"`
	Healthchecks     []HealthcheckMapping        `json:"healthchecks"`
	ConversionIssues []converter.ConversionIssue `json:"conversion_issues"`
}

// Summary contains high-level migration statistics.
type Summary struct {
	TotalMonitors         int `json:"total_monitors"`
	ConvertedMonitors     int `json:"converted_monitors"`
	TotalHeartbeats       int `json:"total_heartbeats"`
	ConvertedHealthchecks int `json:"converted_healthchecks"`
	TotalIssues           int `json:"total_issues"`
	CriticalIssues        int `json:"critical_issues"`
	Warnings              int `json:"warnings"`
}

// MonitorMapping maps Better Stack monitor to Hyperping monitor.
type MonitorMapping struct {
	BetterStackID   string   `json:"betterstack_id"`
	BetterStackName string   `json:"betterstack_name"`
	HyperpingName   string   `json:"hyperping_name"`
	ResourceName    string   `json:"resource_name"`
	Protocol        string   `json:"protocol"`
	Issues          []string `json:"issues,omitempty"`
}

// HealthcheckMapping maps Better Stack heartbeat to Hyperping healthcheck.
type HealthcheckMapping struct {
	BetterStackID   string   `json:"betterstack_id"`
	BetterStackName string   `json:"betterstack_name"`
	HyperpingName   string   `json:"hyperping_name"`
	ResourceName    string   `json:"resource_name"`
	Period          int      `json:"period"`
	Issues          []string `json:"issues,omitempty"`
}

// Generate creates a migration report.
func Generate(
	bsMonitors []betterstack.Monitor,
	bsHeartbeats []betterstack.Heartbeat,
	convertedMonitors []converter.ConvertedMonitor,
	convertedHealthchecks []converter.ConvertedHealthcheck,
	monitorIssues []converter.ConversionIssue,
	healthcheckIssues []converter.ConversionIssue,
) *Report {
	report := &Report{
		ConversionIssues: append(monitorIssues, healthcheckIssues...),
	}

	// Build monitor mappings
	for i, m := range convertedMonitors {
		mapping := MonitorMapping{
			BetterStackID:   bsMonitors[i].ID,
			BetterStackName: bsMonitors[i].Attributes.PronouncableName,
			HyperpingName:   m.Name,
			ResourceName:    m.ResourceName,
			Protocol:        m.Protocol,
			Issues:          m.Issues,
		}
		report.Monitors = append(report.Monitors, mapping)
	}

	// Build healthcheck mappings
	for i, h := range convertedHealthchecks {
		mapping := HealthcheckMapping{
			BetterStackID:   bsHeartbeats[i].ID,
			BetterStackName: bsHeartbeats[i].Attributes.Name,
			HyperpingName:   h.Name,
			ResourceName:    h.ResourceName,
			Period:          h.Period,
			Issues:          h.Issues,
		}
		report.Healthchecks = append(report.Healthchecks, mapping)
	}

	// Calculate summary
	report.Summary = Summary{
		TotalMonitors:         len(bsMonitors),
		ConvertedMonitors:     len(convertedMonitors),
		TotalHeartbeats:       len(bsHeartbeats),
		ConvertedHealthchecks: len(convertedHealthchecks),
		TotalIssues:           len(report.ConversionIssues),
	}

	// Count critical issues and warnings
	for _, issue := range report.ConversionIssues {
		if issue.Severity == "error" {
			report.Summary.CriticalIssues++
		} else {
			report.Summary.Warnings++
		}
	}

	return report
}

// JSON returns the report as JSON.
func (r *Report) JSON() string {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Sprintf("{\"error\": \"failed to marshal report: %v\"}", err)
	}
	return string(data)
}

// PrintSummary prints a human-readable summary to the writer.
func (r *Report) PrintSummary(w io.Writer) {
	fmt.Fprintln(w, "Migration Summary:")
	fmt.Fprintf(w, "  Monitors:     %d migrated from %d\n", r.Summary.ConvertedMonitors, r.Summary.TotalMonitors)
	fmt.Fprintf(w, "  Healthchecks: %d migrated from %d heartbeats\n", r.Summary.ConvertedHealthchecks, r.Summary.TotalHeartbeats)
	fmt.Fprintf(w, "  Total Issues: %d (%d critical, %d warnings)\n",
		r.Summary.TotalIssues,
		r.Summary.CriticalIssues,
		r.Summary.Warnings,
	)

	if r.Summary.CriticalIssues > 0 {
		fmt.Fprintln(w, "\n⚠️  Critical issues found! Review migration-report.json and manual-steps.md")
	} else if r.Summary.Warnings > 0 {
		fmt.Fprintln(w, "\n⚠️  Warnings found. Review migration-report.json for details")
	} else {
		fmt.Fprintln(w, "\n✓ No critical issues detected")
	}
}
