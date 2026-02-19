// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package report

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

// MigrationReport contains the complete migration report.
type MigrationReport struct {
	Timestamp         time.Time      `json:"timestamp"`
	TotalChecks       int            `json:"total_checks"`
	SupportedChecks   int            `json:"supported_checks"`
	UnsupportedChecks int            `json:"unsupported_checks"`
	ChecksByType      map[string]int `json:"checks_by_type"`
	UnsupportedTypes  map[string]int `json:"unsupported_types"`
	ManualSteps       []ManualStep   `json:"manual_steps"`
	Warnings          []string       `json:"warnings"`
}

// ManualStep represents a manual action required.
type ManualStep struct {
	CheckID     int    `json:"check_id"`
	CheckName   string `json:"check_name"`
	CheckType   string `json:"check_type"`
	Description string `json:"description"`
	Action      string `json:"action"`
}

// Reporter generates migration reports.
type Reporter struct{}

// NewReporter creates a new Reporter.
func NewReporter() *Reporter {
	return &Reporter{}
}

// GenerateReport generates a comprehensive migration report.
func (r *Reporter) GenerateReport(checks []pingdom.Check, results []converter.ConversionResult) *MigrationReport {
	report := &MigrationReport{
		Timestamp:        time.Now(),
		TotalChecks:      len(checks),
		ChecksByType:     make(map[string]int),
		UnsupportedTypes: make(map[string]int),
		ManualSteps:      []ManualStep{},
		Warnings:         []string{},
	}

	for i, check := range checks {
		result := results[i]

		// Count by type
		report.ChecksByType[check.Type]++

		if result.Supported {
			report.SupportedChecks++

			// Add warnings for special handling
			if len(result.Notes) > 0 {
				for _, note := range result.Notes {
					report.Warnings = append(report.Warnings, fmt.Sprintf("Check %d (%s): %s", check.ID, check.Name, note))
				}
			}
		} else {
			report.UnsupportedChecks++
			report.UnsupportedTypes[result.UnsupportedType]++

			// Add manual step
			step := r.generateManualStep(check, result)
			report.ManualSteps = append(report.ManualSteps, step)
		}
	}

	return report
}

func (r *Reporter) generateManualStep(check pingdom.Check, _ converter.ConversionResult) ManualStep {
	step := ManualStep{
		CheckID:   check.ID,
		CheckName: check.Name,
		CheckType: check.Type,
	}

	switch check.Type {
	case "dns":
		step.Description = "DNS checks are not directly supported by Hyperping"
		step.Action = "Option 1: Create HTTP monitor to DNS-over-HTTPS service (e.g., https://dns.google/resolve?name=example.com&type=A)\n" +
			"Option 2: Monitor the service that relies on DNS instead\n" +
			"Option 3: Use external DNS monitoring tool with webhook to Hyperping healthcheck"

	case "udp":
		step.Description = "UDP checks are not supported by Hyperping"
		step.Action = "Option 1: Use TCP alternative if service supports it\n" +
			"Option 2: Monitor the application that uses the UDP service\n" +
			"Option 3: Use external monitoring tool with webhook to Hyperping healthcheck"

	case "transaction":
		step.Description = "Transaction/browser checks require external script"
		step.Action = "Create Playwright/Selenium script for transaction:\n" +
			"1. Write script simulating user journey\n" +
			"2. Deploy as Kubernetes CronJob or scheduled Lambda\n" +
			"3. Create Hyperping healthcheck\n" +
			"4. Script pings healthcheck URL on success\n" +
			"See: docs/guides/migrate-from-pingdom.md#transaction-check-equivalent"

	default:
		step.Description = fmt.Sprintf("Check type '%s' is not supported", check.Type)
		step.Action = "Manual review required. Contact support for migration options."
	}

	return step
}

// GenerateJSONReport generates a JSON report.
func (r *Reporter) GenerateJSONReport(report *MigrationReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshaling report: %w", err)
	}

	return string(data), nil
}

// GenerateTextReport generates a human-readable text report.
func (r *Reporter) GenerateTextReport(report *MigrationReport) string {
	var sb strings.Builder

	sb.WriteString("=================================================================\n")
	sb.WriteString("Pingdom to Hyperping Migration Report\n")
	sb.WriteString("=================================================================\n\n")

	fmt.Fprintf(&sb, "Generated: %s\n\n", report.Timestamp.Format(time.RFC3339))

	sb.WriteString("Summary\n")
	sb.WriteString("-------\n")
	fmt.Fprintf(&sb, "Total Checks:       %d\n", report.TotalChecks)
	fmt.Fprintf(&sb, "Supported:          %d (%.1f%%)\n", report.SupportedChecks, float64(report.SupportedChecks)/float64(report.TotalChecks)*100)
	fmt.Fprintf(&sb, "Unsupported:        %d (%.1f%%)\n", report.UnsupportedChecks, float64(report.UnsupportedChecks)/float64(report.TotalChecks)*100)
	fmt.Fprintf(&sb, "Manual Steps:       %d\n\n", len(report.ManualSteps))

	if len(report.ChecksByType) > 0 {
		sb.WriteString("Checks by Type\n")
		sb.WriteString("--------------\n")
		for checkType, count := range report.ChecksByType {
			fmt.Fprintf(&sb, "%-15s %d\n", checkType+":", count)
		}
		sb.WriteString("\n")
	}

	if len(report.UnsupportedTypes) > 0 {
		sb.WriteString("Unsupported Check Types\n")
		sb.WriteString("-----------------------\n")
		for checkType, count := range report.UnsupportedTypes {
			fmt.Fprintf(&sb, "%-15s %d check(s)\n", checkType+":", count)
		}
		sb.WriteString("\n")
	}

	if len(report.Warnings) > 0 {
		sb.WriteString("Warnings\n")
		sb.WriteString("--------\n")
		for i, warning := range report.Warnings {
			fmt.Fprintf(&sb, "%d. %s\n", i+1, warning)
		}
		sb.WriteString("\n")
	}

	if len(report.ManualSteps) > 0 {
		sb.WriteString("Manual Steps Required\n")
		sb.WriteString("=====================\n\n")

		for i, step := range report.ManualSteps {
			fmt.Fprintf(&sb, "%d. Check ID %d: %s\n", i+1, step.CheckID, step.CheckName)
			fmt.Fprintf(&sb, "   Type: %s\n", step.CheckType)
			fmt.Fprintf(&sb, "   Issue: %s\n", step.Description)
			sb.WriteString("   Action:\n")
			for _, line := range strings.Split(step.Action, "\n") {
				fmt.Fprintf(&sb, "   %s\n", line)
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("=================================================================\n")

	return sb.String()
}

// GenerateManualStepsMarkdown generates a markdown file for manual steps.
func (r *Reporter) GenerateManualStepsMarkdown(report *MigrationReport) string {
	var sb strings.Builder

	sb.WriteString("# Manual Migration Steps\n\n")
	fmt.Fprintf(&sb, "Generated: %s\n\n", report.Timestamp.Format(time.RFC1123))

	if len(report.ManualSteps) == 0 {
		sb.WriteString("No manual steps required. All checks were successfully converted!\n")
		return sb.String()
	}

	fmt.Fprintf(&sb, "The following %d check(s) require manual intervention:\n\n", len(report.ManualSteps))

	sb.WriteString("---\n\n")

	for i, step := range report.ManualSteps {
		fmt.Fprintf(&sb, "## %d. %s (ID: %d)\n\n", i+1, step.CheckName, step.CheckID)
		fmt.Fprintf(&sb, "**Type:** `%s`\n\n", step.CheckType)
		fmt.Fprintf(&sb, "**Issue:** %s\n\n", step.Description)
		sb.WriteString("**Action Required:**\n\n")
		sb.WriteString(step.Action)
		sb.WriteString("\n\n---\n\n")
	}

	sb.WriteString("## Additional Resources\n\n")
	sb.WriteString("- [Pingdom Migration Guide](../docs/guides/migrate-from-pingdom.md)\n")
	sb.WriteString("- [Hyperping Documentation](https://hyperping.io/docs)\n")
	sb.WriteString("- [Transaction Check Alternatives](../docs/guides/migrate-from-pingdom.md#transaction-check-conversion)\n")

	return sb.String()
}
