// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"fmt"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
)

// GenerateTerraform generates Terraform HCL configuration from conversion results.
func GenerateTerraform(result *converter.ConversionResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("# Terraform configuration generated from UptimeRobot migration\n")
	sb.WriteString("# Review and adjust as needed before applying\n")
	sb.WriteString("#\n")
	fmt.Fprintf(&sb, "# Total monitors: %d\n", len(result.Monitors))
	fmt.Fprintf(&sb, "# Total healthchecks: %d\n", len(result.Healthchecks))
	if len(result.Skipped) > 0 {
		fmt.Fprintf(&sb, "# Skipped resources: %d (see comments below)\n", len(result.Skipped))
	}
	sb.WriteString("\n")

	// Terraform and provider configuration
	sb.WriteString("terraform {\n")
	sb.WriteString("  required_providers {\n")
	sb.WriteString("    hyperping = {\n")
	sb.WriteString("      source  = \"develeap/hyperping\"\n")
	sb.WriteString("      version = \"~> 1.0\"\n")
	sb.WriteString("    }\n")
	sb.WriteString("  }\n")
	sb.WriteString("}\n\n")

	sb.WriteString("provider \"hyperping\" {\n")
	sb.WriteString("  # API key will be read from HYPERPING_API_KEY environment variable\n")
	sb.WriteString("}\n\n")

	// Variables for escalation policies
	if len(result.Monitors) > 0 || len(result.Healthchecks) > 0 {
		sb.WriteString("# Escalation Policy Configuration\n")
		sb.WriteString("# Create escalation policies in Hyperping dashboard first,\n")
		sb.WriteString("# then set their UUIDs here or via terraform.tfvars\n")
		sb.WriteString("variable \"escalation_policy\" {\n")
		sb.WriteString("  description = \"Default escalation policy UUID for alerts\"\n")
		sb.WriteString("  type        = string\n")
		sb.WriteString("  default     = \"\"  # Set this to your escalation policy UUID\n")
		sb.WriteString("}\n\n")
	}

	// Generate monitors
	if len(result.Monitors) > 0 {
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Monitors\n")
		sb.WriteString("# ============================================\n\n")

		for _, m := range result.Monitors {
			generateMonitorResource(&sb, m)
		}
	}

	// Generate healthchecks
	if len(result.Healthchecks) > 0 {
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Healthchecks (from Heartbeat monitors)\n")
		sb.WriteString("# ============================================\n\n")

		for _, h := range result.Healthchecks {
			generateHealthcheckResource(&sb, h)
		}
	}

	// Document skipped resources
	if len(result.Skipped) > 0 {
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Skipped Resources\n")
		sb.WriteString("# ============================================\n")
		sb.WriteString("# The following monitors could not be migrated:\n")
		sb.WriteString("#\n")
		for _, s := range result.Skipped {
			fmt.Fprintf(&sb, "# - %s (ID: %d, Type: %d): %s\n", s.Name, s.ID, s.Type, s.Reason)
		}
		sb.WriteString("\n")
	}

	// Outputs
	sb.WriteString("# ============================================\n")
	sb.WriteString("# Outputs\n")
	sb.WriteString("# ============================================\n\n")

	if len(result.Healthchecks) > 0 {
		sb.WriteString("# Healthcheck ping URLs\n")
		sb.WriteString("# Use these URLs to update your heartbeat scripts\n")
		for _, h := range result.Healthchecks {
			fmt.Fprintf(&sb, "output \"%s_ping_url\" {\n", h.ResourceName)
			fmt.Fprintf(&sb, "  description = \"Ping URL for %s\"\n", escapeString(h.Name))
			fmt.Fprintf(&sb, "  value       = hyperping_healthcheck.%s.ping_url\n", h.ResourceName)
			sb.WriteString("  sensitive   = true\n")
			sb.WriteString("}\n\n")
		}
	}

	return sb.String()
}

// generateMonitorResource generates HCL for a single monitor resource.
func generateMonitorResource(sb *strings.Builder, m converter.HyperpingMonitor) {
	fmt.Fprintf(sb, "# Original UptimeRobot Monitor ID: %d\n", m.OriginalID)
	if len(m.Warnings) > 0 {
		sb.WriteString("# Warnings:\n")
		for _, w := range m.Warnings {
			fmt.Fprintf(sb, "#   - %s\n", w)
		}
	}
	fmt.Fprintf(sb, "resource \"hyperping_monitor\" \"%s\" {\n", m.ResourceName)
	fmt.Fprintf(sb, "  name            = %q\n", m.Name)
	fmt.Fprintf(sb, "  url             = %q\n", m.URL)
	fmt.Fprintf(sb, "  protocol        = %q\n", m.Protocol)

	if m.HTTPMethod != "" {
		fmt.Fprintf(sb, "  http_method     = %q\n", m.HTTPMethod)
	}

	fmt.Fprintf(sb, "  check_frequency = %d\n", m.CheckFrequency)

	if m.ExpectedStatusCode != "" {
		fmt.Fprintf(sb, "  expected_status_code = %q\n", m.ExpectedStatusCode)
	}

	if m.RequiredKeyword != "" {
		fmt.Fprintf(sb, "  required_keyword     = %q\n", m.RequiredKeyword)
	}

	if m.Port > 0 {
		fmt.Fprintf(sb, "  port            = %d\n", m.Port)
	}

	if m.Protocol == "http" {
		fmt.Fprintf(sb, "  follow_redirects = %t\n", m.FollowRedirects)
	}

	if len(m.Regions) > 0 {
		sb.WriteString("  regions         = [")
		for i, r := range m.Regions {
			if i > 0 {
				sb.WriteString(", ")
			}
			fmt.Fprintf(sb, "%q", r)
		}
		sb.WriteString("]\n")
	}

	sb.WriteString("\n")
	sb.WriteString("  # Uncomment to enable alerting:\n")
	sb.WriteString("  # escalation_policy = var.escalation_policy\n")
	sb.WriteString("}\n\n")
}

// generateHealthcheckResource generates HCL for a single healthcheck resource.
func generateHealthcheckResource(sb *strings.Builder, h converter.HyperpingHealthcheck) {
	fmt.Fprintf(sb, "# Original UptimeRobot Heartbeat Monitor ID: %d\n", h.OriginalID)
	if len(h.Warnings) > 0 {
		sb.WriteString("# Warnings:\n")
		for _, w := range h.Warnings {
			fmt.Fprintf(sb, "#   - %s\n", w)
		}
	}
	fmt.Fprintf(sb, "resource \"hyperping_healthcheck\" \"%s\" {\n", h.ResourceName)
	fmt.Fprintf(sb, "  name               = %q\n", h.Name)
	fmt.Fprintf(sb, "  period_value       = %d\n", h.PeriodValue)
	fmt.Fprintf(sb, "  period_type        = %q\n", h.PeriodType)
	fmt.Fprintf(sb, "  grace_period_value = %d\n", h.GracePeriodValue)
	fmt.Fprintf(sb, "  grace_period_type  = %q\n", h.GracePeriodType)
	sb.WriteString("\n")
	sb.WriteString("  # Uncomment to enable alerting:\n")
	sb.WriteString("  # escalation_policy = var.escalation_policy\n")
	sb.WriteString("}\n\n")
}

// escapeString escapes a string for use in Terraform configuration.
func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
