// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// TerraformGenerator generates Terraform HCL configuration.
type TerraformGenerator struct {
	prefix string
}

// NewTerraformGenerator creates a new TerraformGenerator.
func NewTerraformGenerator(prefix string) *TerraformGenerator {
	return &TerraformGenerator{
		prefix: prefix,
	}
}

// GenerateHCL generates Terraform HCL for converted monitors.
func (g *TerraformGenerator) GenerateHCL(checks []pingdom.Check, results []converter.ConversionResult) string {
	var sb strings.Builder

	sb.WriteString("# Generated from Pingdom export\n")
	sb.WriteString("# Review and adjust as needed before applying\n\n")

	for i, check := range checks {
		result := results[i]

		sb.WriteString(fmt.Sprintf("# Pingdom Check ID: %d\n", check.ID))
		sb.WriteString(fmt.Sprintf("# Original Name: %s\n", check.Name))
		sb.WriteString(fmt.Sprintf("# Type: %s\n", check.Type))

		if len(check.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("# Tags: %s\n", converter.TagsToString(check.Tags)))
		}

		if !result.Supported {
			sb.WriteString(fmt.Sprintf("# UNSUPPORTED: %s\n", result.UnsupportedType))
			for _, note := range result.Notes {
				sb.WriteString(fmt.Sprintf("# NOTE: %s\n", note))
			}
			sb.WriteString("\n")
			continue
		}

		if result.Monitor != nil {
			g.generateMonitorHCL(&sb, check, result.Monitor)
		}

		if len(result.Notes) > 0 {
			for _, note := range result.Notes {
				sb.WriteString(fmt.Sprintf("  # NOTE: %s\n", note))
			}
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *TerraformGenerator) generateMonitorHCL(sb *strings.Builder, _ pingdom.Check, monitor *client.CreateMonitorRequest) {
	tfName := g.terraformName(monitor.Name)

	fmt.Fprintf(sb, "resource \"hyperping_monitor\" %q {\n", tfName)
	fmt.Fprintf(sb, "  name     = %q\n", escapeHCL(monitor.Name))
	fmt.Fprintf(sb, "  url      = %q\n", escapeHCL(monitor.URL))
	fmt.Fprintf(sb, "  protocol = %q\n", monitor.Protocol)

	if monitor.HTTPMethod != "" && monitor.HTTPMethod != "GET" {
		fmt.Fprintf(sb, "  http_method = %q\n", monitor.HTTPMethod)
	}

	if monitor.CheckFrequency != 60 {
		fmt.Fprintf(sb, "  check_frequency = %d\n", monitor.CheckFrequency)
	}

	if len(monitor.Regions) > 0 {
		fmt.Fprintf(sb, "  regions = %s\n", formatStringList(monitor.Regions))
	}

	if monitor.Port != nil && *monitor.Port != 0 {
		fmt.Fprintf(sb, "  port = %d\n", *monitor.Port)
	}

	if monitor.FollowRedirects != nil && !*monitor.FollowRedirects {
		sb.WriteString("  follow_redirects = false\n")
	}

	if monitor.ExpectedStatusCode != "" && monitor.ExpectedStatusCode != "200" {
		fmt.Fprintf(sb, "  expected_status_code = %q\n", monitor.ExpectedStatusCode)
	}

	if monitor.RequiredKeyword != nil && *monitor.RequiredKeyword != "" {
		fmt.Fprintf(sb, "  required_keyword = %q\n", escapeHCL(*monitor.RequiredKeyword))
	}

	if len(monitor.RequestHeaders) > 0 {
		sb.WriteString("  request_headers = {\n")
		for _, h := range monitor.RequestHeaders {
			fmt.Fprintf(sb, "    %q = %q\n", escapeHCL(h.Name), escapeHCL(h.Value))
		}
		sb.WriteString("  }\n")
	}

	if monitor.RequestBody != nil && *monitor.RequestBody != "" {
		fmt.Fprintf(sb, "  request_body = %q\n", escapeHCL(*monitor.RequestBody))
	}

	if monitor.Paused {
		sb.WriteString("  paused = true\n")
	}

	sb.WriteString("}\n")
}

// terraformName converts a resource name to a valid Terraform identifier.
func (g *TerraformGenerator) terraformName(name string) string {
	// Remove brackets and their contents
	re := regexp.MustCompile(`\[.*?\]`)
	tfName := re.ReplaceAllString(name, "")

	// Replace non-alphanumeric characters with underscores
	re = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	tfName = re.ReplaceAllString(tfName, "_")

	// Remove leading/trailing underscores
	tfName = strings.Trim(tfName, "_")

	// Ensure it starts with a letter
	if tfName != "" && (tfName[0] >= '0' && tfName[0] <= '9') {
		tfName = "monitor_" + tfName
	}

	// Convert to lowercase
	tfName = strings.ToLower(tfName)

	// Add prefix if specified
	if g.prefix != "" {
		tfName = g.prefix + tfName
	}

	// Fallback for empty names
	if tfName == "" {
		tfName = "monitor"
	}

	return tfName
}

// escapeHCL escapes a string for HCL output.
func escapeHCL(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

// formatStringList formats a Go string slice as an HCL list.
func formatStringList(items []string) string {
	if len(items) == 0 {
		return "[]"
	}

	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = fmt.Sprintf("%q", item)
	}

	return "[" + strings.Join(quoted, ", ") + "]"
}
