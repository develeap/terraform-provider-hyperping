// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"strings"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func (g *Generator) generateMonitorHCL(sb *strings.Builder, m client.Monitor) {
	name := g.terraformName(m.Name)

	fmt.Fprintf(sb, "resource \"hyperping_monitor\" %q {\n", name)
	fmt.Fprintf(sb, "  name     = %q\n", m.Name)
	fmt.Fprintf(sb, "  url      = %q\n", m.URL)
	fmt.Fprintf(sb, "  protocol = %q\n", m.Protocol)

	if m.HTTPMethod != "" && m.HTTPMethod != "GET" {
		fmt.Fprintf(sb, "  http_method = %q\n", m.HTTPMethod)
	}

	if m.CheckFrequency != 60 {
		fmt.Fprintf(sb, "  check_frequency = %d\n", m.CheckFrequency)
	}

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

	if m.AlertsWait > 0 {
		fmt.Fprintf(sb, "  alerts_wait = %d\n", m.AlertsWait)
	}

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

	if m.RequestBody != "" {
		fmt.Fprintf(sb, "  request_body = %q\n", m.RequestBody)
	}

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

	if sp.Settings.Theme != "" && sp.Settings.Theme != "system" {
		fmt.Fprintf(sb, "    theme = %q\n", sp.Settings.Theme)
	}

	if sp.Settings.Font != "" && sp.Settings.Font != "Inter" {
		fmt.Fprintf(sb, "    font = %q\n", sp.Settings.Font)
	}

	if sp.Settings.AccentColor != "" && sp.Settings.AccentColor != "#36b27e" {
		fmt.Fprintf(sb, "    accent_color = %q\n", sp.Settings.AccentColor)
	}

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

	if i.Text.En != "" {
		fmt.Fprintf(sb, "  text = %q\n", i.Text.En)
	}

	if i.Type != "" && i.Type != "incident" {
		fmt.Fprintf(sb, "  type = %q\n", i.Type)
	}

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

	if m.Text.En != "" {
		fmt.Fprintf(sb, "  text = %q\n", m.Text.En)
	}

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
