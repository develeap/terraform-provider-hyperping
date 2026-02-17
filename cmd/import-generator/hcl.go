// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// buildOptionalStringField returns an HCL line for a string field only when
// the value is non-empty and differs from the given skip value.
// Returns an empty string when the field should be omitted.
func buildOptionalStringField(name, value, skipValue string) string {
	if value == "" || value == skipValue {
		return ""
	}
	return fmt.Sprintf("  %s = %q\n", name, value)
}

// buildOptionalIntField returns an HCL line for an int field only when
// the value differs from the skip value (typically 0 or a default).
func buildOptionalIntField(name string, value, skipValue int) string {
	if value == skipValue {
		return ""
	}
	return fmt.Sprintf("  %s = %d\n", name, value)
}

func (g *Generator) generateMonitorHCL(sb *strings.Builder, m client.Monitor) {
	name := g.terraformName(m.Name)

	fmt.Fprintf(sb, "resource \"hyperping_monitor\" %q {\n", name)
	fmt.Fprintf(sb, "  name     = %q\n", m.Name)
	fmt.Fprintf(sb, "  url      = %q\n", m.URL)
	fmt.Fprintf(sb, "  protocol = %q\n", m.Protocol)

	sb.WriteString(buildOptionalStringField("http_method", m.HTTPMethod, "GET"))
	sb.WriteString(buildOptionalIntField("check_frequency", m.CheckFrequency, 60))

	if len(m.Regions) > 0 {
		fmt.Fprintf(sb, "  regions = %s\n", formatStringList(m.Regions))
	}

	if m.Port != nil && *m.Port != 0 {
		fmt.Fprintf(sb, "  port = %d\n", *m.Port)
	}

	if !m.FollowRedirects {
		sb.WriteString("  follow_redirects = false\n")
	}

	if m.ExpectedStatusCode.String() != "" && m.ExpectedStatusCode.String() != "200" {
		fmt.Fprintf(sb, "  expected_status_code = %q\n", m.ExpectedStatusCode.String())
	}

	if m.RequiredKeyword != nil && *m.RequiredKeyword != "" {
		fmt.Fprintf(sb, "  required_keyword = %q\n", *m.RequiredKeyword)
	}

	if m.Paused {
		sb.WriteString("  paused = true\n")
	}

	sb.WriteString(buildOptionalIntField("alerts_wait", m.AlertsWait, 0))

	if m.EscalationPolicy != nil && *m.EscalationPolicy != "" {
		fmt.Fprintf(sb, "  escalation_policy_uuid = %q\n", *m.EscalationPolicy)
	}

	if len(m.RequestHeaders) > 0 {
		sb.WriteString("  request_headers = {\n")
		for _, h := range m.RequestHeaders {
			fmt.Fprintf(sb, "    %q = %q\n", h.Name, h.Value)
		}
		sb.WriteString("  }\n")
	}

	sb.WriteString(buildOptionalStringField("request_body", m.RequestBody, ""))

	sb.WriteString("}\n")
}

func (g *Generator) generateHealthcheckHCL(sb *strings.Builder, h client.Healthcheck) {
	name := g.terraformName(h.Name)

	fmt.Fprintf(sb, "resource \"hyperping_healthcheck\" %q {\n", name)
	fmt.Fprintf(sb, "  name = %q\n", h.Name)

	if h.Cron != "" {
		fmt.Fprintf(sb, "  cron = %q\n", h.Cron)
		if h.Timezone != "" {
			fmt.Fprintf(sb, "  timezone = %q\n", h.Timezone)
		}
	} else if h.PeriodValue != nil && *h.PeriodValue > 0 {
		fmt.Fprintf(sb, "  period_value = %d\n", *h.PeriodValue)
		fmt.Fprintf(sb, "  period_type = %q\n", h.PeriodType)
	}

	if h.GracePeriod > 0 {
		fmt.Fprintf(sb, "  grace_period = %d\n", h.GracePeriod)
	}

	if h.IsPaused {
		sb.WriteString("  paused = true\n")
	}

	sb.WriteString("}\n")
}

func (g *Generator) generateStatusPageHCL(sb *strings.Builder, sp client.StatusPage) {
	name := g.terraformName(sp.Name)

	fmt.Fprintf(sb, "resource \"hyperping_statuspage\" %q {\n", name)
	fmt.Fprintf(sb, "  name             = %q\n", sp.Name)
	fmt.Fprintf(sb, "  hosted_subdomain = %q\n", sp.HostedSubdomain)

	if sp.Hostname != nil && *sp.Hostname != "" {
		fmt.Fprintf(sb, "  hostname = %q\n", *sp.Hostname)
	}

	// Settings block
	sb.WriteString("\n  settings = {\n")
	fmt.Fprintf(sb, "    name      = %q\n", sp.Name)

	if len(sp.Settings.Languages) > 0 {
		fmt.Fprintf(sb, "    languages = %s\n", formatStringList(sp.Settings.Languages))
	} else {
		sb.WriteString("    languages = [\"en\"]\n")
	}

	sb.WriteString(buildOptionalStringField("theme", sp.Settings.Theme, "system"))
	sb.WriteString(buildOptionalStringField("font", sp.Settings.Font, "Inter"))
	sb.WriteString(buildOptionalStringField("accent_color", sp.Settings.AccentColor, "#36b27e"))

	sb.WriteString("  }\n")

	// Sections (simplified - just note they exist)
	if len(sp.Sections) > 0 {
		sb.WriteString("\n  # Note: Sections imported - review and adjust as needed\n")
		sb.WriteString("  # sections = [...]\n")
	}

	sb.WriteString("}\n")
}

func (g *Generator) generateIncidentHCL(sb *strings.Builder, i client.Incident) {
	name := g.terraformName(i.Title.En)

	fmt.Fprintf(sb, "resource \"hyperping_incident\" %q {\n", name)
	fmt.Fprintf(sb, "  title = %q\n", i.Title.En)

	sb.WriteString(buildOptionalStringField("text", i.Text.En, ""))
	sb.WriteString(buildOptionalStringField("type", i.Type, "incident"))

	if len(i.StatusPages) > 0 {
		fmt.Fprintf(sb, "  status_pages = %s\n", formatStringList(i.StatusPages))
	}

	if len(i.AffectedComponents) > 0 {
		fmt.Fprintf(sb, "  affected_components = %s\n", formatStringList(i.AffectedComponents))
	}

	sb.WriteString("}\n")
}

func (g *Generator) generateMaintenanceHCL(sb *strings.Builder, m client.Maintenance) {
	// Use Name if Title is empty
	titleText := m.Title.En
	if titleText == "" {
		titleText = m.Name
	}
	name := g.terraformName(titleText)

	fmt.Fprintf(sb, "resource \"hyperping_maintenance\" %q {\n", name)
	fmt.Fprintf(sb, "  title = %q\n", titleText)

	sb.WriteString(buildOptionalStringField("text", m.Text.En, ""))

	if m.StartDate != nil {
		fmt.Fprintf(sb, "  start_date = %q\n", *m.StartDate)
	}
	if m.EndDate != nil {
		fmt.Fprintf(sb, "  end_date   = %q\n", *m.EndDate)
	}

	if len(m.StatusPages) > 0 {
		fmt.Fprintf(sb, "  status_pages = %s\n", formatStringList(m.StatusPages))
	}

	sb.WriteString("}\n")
}

func (g *Generator) generateOutageHCL(sb *strings.Builder, o client.Outage) {
	name := g.terraformName(o.Monitor.Name)

	fmt.Fprintf(sb, "resource \"hyperping_outage\" %q {\n", name)
	fmt.Fprintf(sb, "  monitor_uuid = %q\n", o.Monitor.UUID)

	if o.Description != "" {
		fmt.Fprintf(sb, "  # description = %q\n", o.Description)
	}

	// Note: Most outage fields are read-only/computed
	sb.WriteString("  # Note: Outages are mostly read-only. Review fields after import.\n")

	sb.WriteString("}\n")
}
