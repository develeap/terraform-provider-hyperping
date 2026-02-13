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

	// Generate monitors
	if len(monitors) > 0 {
		sb.WriteString("# ===== MONITORS =====\n\n")
		for _, m := range monitors {
			sb.WriteString(g.generateMonitorBlock(m))
		}
	}

	// Generate healthchecks
	if len(healthchecks) > 0 {
		sb.WriteString("# ===== HEALTHCHECKS =====\n\n")
		for _, h := range healthchecks {
			sb.WriteString(g.generateHealthcheckBlock(h))
		}
	}

	return sb.String()
}

func (g *Generator) generateMonitorBlock(m converter.ConvertedMonitor) string {
	var sb strings.Builder

	// Add issues as comments if any
	if len(m.Issues) > 0 {
		sb.WriteString("# MIGRATION NOTES:\n")
		for _, issue := range m.Issues {
			sb.WriteString(fmt.Sprintf("# - %s\n", issue))
		}
	}

	sb.WriteString(fmt.Sprintf("resource \"hyperping_monitor\" %q {\n", m.ResourceName))
	sb.WriteString(fmt.Sprintf("  name                 = %s\n", quoteString(m.Name)))
	sb.WriteString(fmt.Sprintf("  url                  = %s\n", quoteString(m.URL)))

	// Protocol (only if not default)
	if m.Protocol != "http" {
		sb.WriteString(fmt.Sprintf("  protocol             = %s\n", quoteString(m.Protocol)))
	}

	// HTTP method (only if not GET)
	if m.HTTPMethod != "GET" && m.Protocol == "http" {
		sb.WriteString(fmt.Sprintf("  http_method          = %s\n", quoteString(m.HTTPMethod)))
	}

	sb.WriteString(fmt.Sprintf("  check_frequency      = %d\n", m.CheckFrequency))

	// Expected status code (only if not 200)
	if m.ExpectedStatusCode != "200" {
		sb.WriteString(fmt.Sprintf("  expected_status_code = %s\n", quoteString(m.ExpectedStatusCode)))
	}

	// Follow redirects (only if false)
	if !m.FollowRedirects {
		sb.WriteString("  follow_redirects     = false\n")
	}

	// Paused (only if true)
	if m.Paused {
		sb.WriteString("  paused               = true\n")
	}

	// Port (only for port protocol)
	if m.Port > 0 && m.Protocol == "port" {
		sb.WriteString(fmt.Sprintf("  port                 = %d\n", m.Port))
	}

	// Regions
	if len(m.Regions) > 0 {
		sb.WriteString("\n  regions = [\n")
		for _, region := range m.Regions {
			sb.WriteString(fmt.Sprintf("    %s,\n", quoteString(region)))
		}
		sb.WriteString("  ]\n")
	}

	// Request headers
	if len(m.RequestHeaders) > 0 {
		sb.WriteString("\n  request_headers = [\n")
		for _, header := range m.RequestHeaders {
			sb.WriteString("    {\n")
			sb.WriteString(fmt.Sprintf("      name  = %s\n", quoteString(header.Name)))
			sb.WriteString(fmt.Sprintf("      value = %s\n", quoteString(header.Value)))
			sb.WriteString("    },\n")
		}
		sb.WriteString("  ]\n")
	}

	// Request body
	if m.RequestBody != "" {
		sb.WriteString(fmt.Sprintf("\n  request_body = %s\n", quoteString(m.RequestBody)))
	}

	sb.WriteString("}\n\n")

	return sb.String()
}

func (g *Generator) generateHealthcheckBlock(h converter.ConvertedHealthcheck) string {
	var sb strings.Builder

	// Add issues as comments if any
	if len(h.Issues) > 0 {
		sb.WriteString("# MIGRATION NOTES:\n")
		for _, issue := range h.Issues {
			sb.WriteString(fmt.Sprintf("# - %s\n", issue))
		}
	}

	sb.WriteString(fmt.Sprintf("resource \"hyperping_healthcheck\" %q {\n", h.ResourceName))
	sb.WriteString(fmt.Sprintf("  name               = %s\n", quoteString(h.Name)))

	// Note: Hyperping uses cron format, Better Stack uses simple period
	// Convert period to cron expression
	cronExpr := periodToCron(h.Period)
	sb.WriteString(fmt.Sprintf("  cron               = %s\n", quoteString(cronExpr)))
	sb.WriteString("  timezone           = \"UTC\"\n")

	// Grace period
	gracePeriodMinutes := h.Grace / 60
	if gracePeriodMinutes < 1 {
		gracePeriodMinutes = 1
	}
	sb.WriteString(fmt.Sprintf("  grace_period_value = %d\n", gracePeriodMinutes))
	sb.WriteString("  grace_period_type  = \"minutes\"\n")

	// Paused (only if true)
	if h.Paused {
		sb.WriteString("  paused             = true\n")
	}

	sb.WriteString("}\n\n")

	return sb.String()
}

func quoteString(s string) string {
	// Escape special characters
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")
	return fmt.Sprintf("\"%s\"", escaped)
}

func periodToCron(periodSeconds int) string {
	// Convert period in seconds to cron expression
	switch {
	case periodSeconds <= 60:
		return "* * * * *" // Every minute
	case periodSeconds < 3600:
		minutes := periodSeconds / 60
		return fmt.Sprintf("*/%d * * * *", minutes) // Every N minutes
	case periodSeconds == 3600:
		return "0 * * * *" // Every hour at minute 0
	case periodSeconds < 86400:
		hours := periodSeconds / 3600
		return fmt.Sprintf("0 */%d * * *", hours) // Every N hours
	default:
		// Daily or longer - default to daily at midnight
		return "0 0 * * *"
	}
}

func getCurrentTimestamp() string {
	return "2026-02-13T00:00:00Z"
}
