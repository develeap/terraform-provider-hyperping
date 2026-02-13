// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package report

import (
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

// Report represents a migration report.
type Report struct {
	Timestamp string          `json:"timestamp"`
	Summary   Summary         `json:"summary"`
	Monitors  []MonitorReport `json:"monitors"`
	Warnings  []Warning       `json:"warnings"`
	Errors    []Error         `json:"errors"`
}

// Summary contains migration summary statistics.
type Summary struct {
	TotalMonitors        int `json:"total_monitors"`
	MigratedMonitors     int `json:"migrated_monitors"`
	MigratedHealthchecks int `json:"migrated_healthchecks"`
	SkippedMonitors      int `json:"skipped_monitors"`
	MonitorsWithWarnings int `json:"monitors_with_warnings"`
}

// MonitorReport contains details about a migrated monitor.
type MonitorReport struct {
	OriginalID      int      `json:"original_id"`
	OriginalName    string   `json:"original_name"`
	OriginalType    int      `json:"original_type"`
	ResourceType    string   `json:"resource_type"`
	ResourceName    string   `json:"resource_name"`
	MigrationStatus string   `json:"migration_status"`
	Warnings        []string `json:"warnings,omitempty"`
}

// Warning represents a migration warning.
type Warning struct {
	Resource string `json:"resource"`
	Message  string `json:"message"`
}

// Error represents a migration error.
type Error struct {
	Resource string `json:"resource"`
	Message  string `json:"message"`
}

// Generate generates a migration report.
func Generate(
	monitors []uptimerobot.Monitor,
	alertContacts []uptimerobot.AlertContact,
	result *converter.ConversionResult,
) *Report {
	report := &Report{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Summary: Summary{
			TotalMonitors:        len(monitors),
			MigratedMonitors:     len(result.Monitors),
			MigratedHealthchecks: len(result.Healthchecks),
			SkippedMonitors:      len(result.Skipped),
		},
		Monitors: []MonitorReport{},
		Warnings: []Warning{},
		Errors:   []Error{},
	}

	// Process migrated monitors
	for _, m := range result.Monitors {
		monitorReport := MonitorReport{
			OriginalID:      m.OriginalID,
			OriginalName:    m.Name,
			ResourceType:    "hyperping_monitor",
			ResourceName:    m.ResourceName,
			MigrationStatus: "migrated",
			Warnings:        m.Warnings,
		}

		// Determine original type from protocol
		switch m.Protocol {
		case "http":
			if m.RequiredKeyword != "" {
				monitorReport.OriginalType = 2 // Keyword
			} else {
				monitorReport.OriginalType = 1 // HTTP
			}
		case "icmp":
			monitorReport.OriginalType = 3 // Ping
		case "port":
			monitorReport.OriginalType = 4 // Port
		}

		report.Monitors = append(report.Monitors, monitorReport)

		// Collect warnings
		if len(m.Warnings) > 0 {
			report.Summary.MonitorsWithWarnings++
			for _, w := range m.Warnings {
				report.Warnings = append(report.Warnings, Warning{
					Resource: m.Name,
					Message:  w,
				})
			}
		}
	}

	// Process migrated healthchecks
	for _, h := range result.Healthchecks {
		monitorReport := MonitorReport{
			OriginalID:      h.OriginalID,
			OriginalName:    h.Name,
			OriginalType:    5, // Heartbeat
			ResourceType:    "hyperping_healthcheck",
			ResourceName:    h.ResourceName,
			MigrationStatus: "migrated",
			Warnings:        h.Warnings,
		}

		report.Monitors = append(report.Monitors, monitorReport)

		// Collect warnings
		if len(h.Warnings) > 0 {
			report.Summary.MonitorsWithWarnings++
			for _, w := range h.Warnings {
				report.Warnings = append(report.Warnings, Warning{
					Resource: h.Name,
					Message:  w,
				})
			}
		}
	}

	// Process skipped monitors
	for _, s := range result.Skipped {
		monitorReport := MonitorReport{
			OriginalID:      s.ID,
			OriginalName:    s.Name,
			OriginalType:    s.Type,
			ResourceType:    "",
			ResourceName:    "",
			MigrationStatus: "skipped",
			Warnings:        []string{s.Reason},
		}

		report.Monitors = append(report.Monitors, monitorReport)

		report.Errors = append(report.Errors, Error{
			Resource: s.Name,
			Message:  s.Reason,
		})
	}

	return report
}
