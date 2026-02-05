// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// Generator generates Terraform import commands and HCL from Hyperping resources.
type Generator struct {
	client    *client.Client
	prefix    string
	resources []string
}

// ResourceData holds fetched resource data for generation.
type ResourceData struct {
	Monitors     []client.Monitor
	Healthchecks []client.Healthcheck
	StatusPages  []client.StatusPage
	Incidents    []client.Incident
	Maintenance  []client.Maintenance
	Outages      []client.Outage
}

// Generate fetches resources and generates output in the specified format.
func (g *Generator) Generate(ctx context.Context, format string) (string, error) {
	data, err := g.fetchResources(ctx)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	switch format {
	case "import":
		g.generateImports(&sb, data)
	case "hcl":
		g.generateHCL(&sb, data)
	case "both":
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Terraform Import Commands\n")
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Run these commands to import existing resources:\n\n")
		g.generateImports(&sb, data)
		sb.WriteString("\n\n")
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Terraform HCL Configuration\n")
		sb.WriteString("# ============================================\n")
		sb.WriteString("# Add this to your .tf files:\n\n")
		g.generateHCL(&sb, data)
	default:
		return "", fmt.Errorf("unknown format: %s", format)
	}

	return sb.String(), nil
}

func (g *Generator) fetchResources(ctx context.Context) (*ResourceData, error) {
	data := &ResourceData{}

	for _, r := range g.resources {
		switch r {
		case "monitors":
			monitors, err := g.client.ListMonitors(ctx)
			if err != nil {
				return nil, fmt.Errorf("fetching monitors: %w", err)
			}
			data.Monitors = monitors

		case "healthchecks":
			healthchecks, err := g.client.ListHealthchecks(ctx)
			if err != nil {
				return nil, fmt.Errorf("fetching healthchecks: %w", err)
			}
			data.Healthchecks = healthchecks

		case "statuspages":
			resp, err := g.client.ListStatusPages(ctx, nil, nil)
			if err != nil {
				return nil, fmt.Errorf("fetching status pages: %w", err)
			}
			data.StatusPages = resp.StatusPages

		case "incidents":
			incidents, err := g.client.ListIncidents(ctx)
			if err != nil {
				return nil, fmt.Errorf("fetching incidents: %w", err)
			}
			data.Incidents = incidents

		case "maintenance":
			maintenance, err := g.client.ListMaintenance(ctx)
			if err != nil {
				return nil, fmt.Errorf("fetching maintenance: %w", err)
			}
			data.Maintenance = maintenance

		case "outages":
			outages, err := g.client.ListOutages(ctx)
			if err != nil {
				return nil, fmt.Errorf("fetching outages: %w", err)
			}
			data.Outages = outages
		}
	}

	return data, nil
}

func (g *Generator) generateImports(sb *strings.Builder, data *ResourceData) {
	for _, m := range data.Monitors {
		name := g.terraformName(m.Name)
		fmt.Fprintf(sb, "terraform import hyperping_monitor.%s %q\n", name, m.UUID)
	}

	for _, h := range data.Healthchecks {
		name := g.terraformName(h.Name)
		fmt.Fprintf(sb, "terraform import hyperping_healthcheck.%s %q\n", name, h.UUID)
	}

	for _, sp := range data.StatusPages {
		name := g.terraformName(sp.Name)
		fmt.Fprintf(sb, "terraform import hyperping_statuspage.%s %q\n", name, sp.UUID)
	}

	for _, i := range data.Incidents {
		name := g.terraformName(i.Title.En)
		fmt.Fprintf(sb, "terraform import hyperping_incident.%s %q\n", name, i.UUID)
	}

	for _, m := range data.Maintenance {
		titleText := m.Title.En
		if titleText == "" {
			titleText = m.Name
		}
		name := g.terraformName(titleText)
		fmt.Fprintf(sb, "terraform import hyperping_maintenance.%s %q\n", name, m.UUID)
	}

	for _, o := range data.Outages {
		name := g.terraformName(o.Monitor.Name)
		fmt.Fprintf(sb, "terraform import hyperping_outage.%s %q\n", name, o.UUID)
	}
}

func (g *Generator) generateHCL(sb *strings.Builder, data *ResourceData) {
	// Monitors
	for _, m := range data.Monitors {
		g.generateMonitorHCL(sb, m)
		sb.WriteString("\n")
	}

	// Healthchecks
	for _, h := range data.Healthchecks {
		g.generateHealthcheckHCL(sb, h)
		sb.WriteString("\n")
	}

	// Status Pages
	for _, sp := range data.StatusPages {
		g.generateStatusPageHCL(sb, sp)
		sb.WriteString("\n")
	}

	// Incidents
	for _, i := range data.Incidents {
		g.generateIncidentHCL(sb, i)
		sb.WriteString("\n")
	}

	// Maintenance
	for _, m := range data.Maintenance {
		g.generateMaintenanceHCL(sb, m)
		sb.WriteString("\n")
	}

	// Outages
	for _, o := range data.Outages {
		g.generateOutageHCL(sb, o)
		sb.WriteString("\n")
	}
}

// terraformName converts a resource name to a valid Terraform identifier.
func (g *Generator) terraformName(name string) string {
	// Replace non-alphanumeric characters with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	tfName := re.ReplaceAllString(name, "_")

	// Remove leading/trailing underscores
	tfName = strings.Trim(tfName, "_")

	// Ensure it starts with a letter
	if tfName != "" && (tfName[0] >= '0' && tfName[0] <= '9') {
		tfName = "r_" + tfName
	}

	// Convert to lowercase
	tfName = strings.ToLower(tfName)

	// Add prefix if specified
	if g.prefix != "" {
		tfName = g.prefix + tfName
	}

	// Fallback for empty names
	if tfName == "" {
		tfName = "resource"
	}

	return tfName
}

// escapeHCL escapes a string for HCL output.
func escapeHCL(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
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
