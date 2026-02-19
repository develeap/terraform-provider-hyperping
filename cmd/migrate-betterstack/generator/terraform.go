// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"fmt"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-betterstack/converter"
)

// Generator generates Terraform HCL and import scripts.
type Generator struct{}

// New creates a new generator.
func New() *Generator {
	return &Generator{}
}

// GenerateTerraform generates Terraform HCL configuration.
func (g *Generator) GenerateTerraform(monitors []converter.ConvertedMonitor, healthchecks []converter.ConvertedHealthcheck) string {
	var sb strings.Builder

	sb.WriteString("# Auto-generated from Better Stack migration\n")
	sb.WriteString("# Generated at: " + getCurrentTimestamp() + "\n")
	sb.WriteString("# Review and customize before applying\n\n")

	sb.WriteString("terraform {\n")
	sb.WriteString("  required_version = \">= 1.8\"\n\n")
	sb.WriteString("  required_providers {\n")
	sb.WriteString("    hyperping = {\n")
	sb.WriteString("      source  = \"develeap/hyperping\"\n")
	sb.WriteString("      version = \"~> 1.0\"\n")
	sb.WriteString("    }\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n\n")

	sb.WriteString("provider \"hyperping\" {\n")
	sb.WriteString("  # API key from HYPERPING_API_KEY environment variable\n")
	sb.WriteString("}\n\n")

	if len(monitors) > 0 {
		sb.WriteString("# ===== MONITORS =====\n\n")
		for _, m := range monitors {
			sb.WriteString(g.generateMonitorBlock(m))
		}
	}

	if len(healthchecks) > 0 {
		sb.WriteString("# ===== HEALTHCHECKS =====\n\n")
		for _, h := range healthchecks {
			sb.WriteString(g.generateHealthcheckBlock(h))
		}
	}

	return sb.String()
}

// monitorOptionalField describes a single optional HCL field for a monitor block.
type monitorOptionalField struct {
	name  string
	value string
	skip  bool
}

// buildMonitorOptionalFields collects all optional fields that should be emitted.
func buildMonitorOptionalFields(m converter.ConvertedMonitor) []monitorOptionalField {
	return []monitorOptionalField{
		{name: "protocol", value: quoteString(m.Protocol), skip: m.Protocol == "http"},
		{name: "http_method", value: quoteString(m.HTTPMethod), skip: m.HTTPMethod == "GET" || m.Protocol != "http"},
		{name: "expected_status_code", value: quoteString(m.ExpectedStatusCode), skip: m.ExpectedStatusCode == "200"},
		{name: "follow_redirects", value: "false", skip: m.FollowRedirects},
		{name: "paused", value: "true", skip: !m.Paused},
		{name: "port", value: fmt.Sprintf("%d", m.Port), skip: m.Port <= 0 || m.Protocol != "port"},
	}
}

// writeMonitorMigrationNotes writes any migration issue comments at the top of a block.
func writeMonitorMigrationNotes(sb *strings.Builder, issues []string) {
	if len(issues) == 0 {
		return
	}
	sb.WriteString("# MIGRATION NOTES:\n")
	for _, issue := range issues {
		fmt.Fprintf(sb, "# - %s\n", issue)
	}
}

// writeMonitorRegions writes the regions block if regions are present.
func writeMonitorRegions(sb *strings.Builder, regions []string) {
	if len(regions) == 0 {
		return
	}
	sb.WriteString("\n  regions = [\n")
	for _, region := range regions {
		fmt.Fprintf(sb, "    %s,\n", quoteString(region))
	}
	sb.WriteString("  ]\n")
}

// writeMonitorRequestHeaders writes the request_headers block if headers are present.
func writeMonitorRequestHeaders(sb *strings.Builder, headers []converter.RequestHeader) {
	if len(headers) == 0 {
		return
	}
	sb.WriteString("\n  request_headers = [\n")
	for _, header := range headers {
		sb.WriteString("    {\n")
		fmt.Fprintf(sb, "      name  = %s\n", quoteString(header.Name))
		fmt.Fprintf(sb, "      value = %s\n", quoteString(header.Value))
		sb.WriteString("    },\n")
	}
	sb.WriteString("  ]\n")
}

func (g *Generator) generateMonitorBlock(m converter.ConvertedMonitor) string {
	var sb strings.Builder

	writeMonitorMigrationNotes(&sb, m.Issues)

	fmt.Fprintf(&sb, "resource \"hyperping_monitor\" %q {\n", m.ResourceName)
	fmt.Fprintf(&sb, "  name                 = %s\n", quoteString(m.Name))
	fmt.Fprintf(&sb, "  url                  = %s\n", quoteString(m.URL))

	for _, f := range buildMonitorOptionalFields(m) {
		if !f.skip {
			fmt.Fprintf(&sb, "  %-20s = %s\n", f.name, f.value)
		}
	}

	fmt.Fprintf(&sb, "  check_frequency      = %d\n", m.CheckFrequency)

	writeMonitorRegions(&sb, m.Regions)
	writeMonitorRequestHeaders(&sb, m.RequestHeaders)

	if m.RequestBody != "" {
		fmt.Fprintf(&sb, "\n  request_body = %s\n", quoteString(m.RequestBody))
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

func (g *Generator) generateHealthcheckBlock(h converter.ConvertedHealthcheck) string {
	var sb strings.Builder

	if len(h.Issues) > 0 {
		sb.WriteString("# MIGRATION NOTES:\n")
		for _, issue := range h.Issues {
			fmt.Fprintf(&sb, "# - %s\n", issue)
		}
	}

	fmt.Fprintf(&sb, "resource \"hyperping_healthcheck\" %q {\n", h.ResourceName)
	fmt.Fprintf(&sb, "  name               = %s\n", quoteString(h.Name))

	cronExpr := periodToCron(h.Period)
	fmt.Fprintf(&sb, "  cron               = %s\n", quoteString(cronExpr))
	sb.WriteString("  timezone           = \"UTC\"\n")

	gracePeriodMinutes := h.Grace / 60
	if gracePeriodMinutes < 1 {
		gracePeriodMinutes = 1
	}
	fmt.Fprintf(&sb, "  grace_period_value = %d\n", gracePeriodMinutes)
	sb.WriteString("  grace_period_type  = \"minutes\"\n")

	if h.Paused {
		sb.WriteString("  paused             = true\n")
	}

	sb.WriteString("}\n\n")
	return sb.String()
}

func quoteString(s string) string {
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")
	return fmt.Sprintf("\"%s\"", escaped)
}

func periodToCron(periodSeconds int) string {
	switch {
	case periodSeconds <= 60:
		return "* * * * *"
	case periodSeconds < 3600:
		minutes := periodSeconds / 60
		return fmt.Sprintf("*/%d * * * *", minutes)
	case periodSeconds == 3600:
		return "0 * * * *"
	case periodSeconds < 86400:
		hours := periodSeconds / 3600
		return fmt.Sprintf("0 */%d * * *", hours)
	default:
		return "0 0 * * *"
	}
}

func getCurrentTimestamp() string {
	return "2026-02-13T00:00:00Z"
}
