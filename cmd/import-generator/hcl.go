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

	sb.WriteString(fmt.Sprintf("resource \"hyperping_monitor\" \"%s\" {\n", name))
	sb.WriteString(fmt.Sprintf("  name     = \"%s\"\n", escapeHCL(m.Name)))
	sb.WriteString(fmt.Sprintf("  url      = \"%s\"\n", escapeHCL(m.URL)))
	sb.WriteString(fmt.Sprintf("  protocol = \"%s\"\n", m.Protocol))

	if m.HTTPMethod != "" && m.HTTPMethod != "GET" {
		sb.WriteString(fmt.Sprintf("  http_method = \"%s\"\n", m.HTTPMethod))
	}

	if m.CheckFrequency != 60 {
		sb.WriteString(fmt.Sprintf("  check_frequency = %d\n", m.CheckFrequency))
	}

	if len(m.Regions) > 0 {
		sb.WriteString(fmt.Sprintf("  regions = %s\n", formatStringList(m.Regions)))
	}

	if m.Port != nil && *m.Port != 0 {
		sb.WriteString(fmt.Sprintf("  port = %d\n", *m.Port))
	}

	if !m.FollowRedirects {
		sb.WriteString("  follow_redirects = false\n")
	}

	if m.ExpectedStatusCode.String() != "" && m.ExpectedStatusCode.String() != "200" {
		sb.WriteString(fmt.Sprintf("  expected_status_code = \"%s\"\n", m.ExpectedStatusCode.String()))
	}

	if m.RequiredKeyword != nil && *m.RequiredKeyword != "" {
		sb.WriteString(fmt.Sprintf("  required_keyword = \"%s\"\n", escapeHCL(*m.RequiredKeyword)))
	}

	if m.Paused {
		sb.WriteString("  paused = true\n")
	}

	if m.AlertsWait > 0 {
		sb.WriteString(fmt.Sprintf("  alerts_wait = %d\n", m.AlertsWait))
	}

	if m.EscalationPolicy != nil && *m.EscalationPolicy != "" {
		sb.WriteString(fmt.Sprintf("  escalation_policy_uuid = \"%s\"\n", *m.EscalationPolicy))
	}

	if len(m.RequestHeaders) > 0 {
		sb.WriteString("  request_headers = {\n")
		for _, h := range m.RequestHeaders {
			sb.WriteString(fmt.Sprintf("    \"%s\" = \"%s\"\n", escapeHCL(h.Name), escapeHCL(h.Value)))
		}
		sb.WriteString("  }\n")
	}

	if m.RequestBody != "" {
		sb.WriteString(fmt.Sprintf("  request_body = \"%s\"\n", escapeHCL(m.RequestBody)))
	}

	sb.WriteString("}\n")
}

func (g *Generator) generateHealthcheckHCL(sb *strings.Builder, h client.Healthcheck) {
	name := g.terraformName(h.Name)

	sb.WriteString(fmt.Sprintf("resource \"hyperping_healthcheck\" \"%s\" {\n", name))
	sb.WriteString(fmt.Sprintf("  name = \"%s\"\n", escapeHCL(h.Name)))

	if h.Cron != "" {
		sb.WriteString(fmt.Sprintf("  cron = \"%s\"\n", h.Cron))
		if h.Timezone != "" {
			sb.WriteString(fmt.Sprintf("  timezone = \"%s\"\n", h.Timezone))
		}
	} else if h.PeriodValue != nil && *h.PeriodValue > 0 {
		sb.WriteString(fmt.Sprintf("  period_value = %d\n", *h.PeriodValue))
		sb.WriteString(fmt.Sprintf("  period_type = \"%s\"\n", h.PeriodType))
	}

	if h.GracePeriod > 0 {
		sb.WriteString(fmt.Sprintf("  grace_period = %d\n", h.GracePeriod))
	}

	if h.IsPaused {
		sb.WriteString("  paused = true\n")
	}

	sb.WriteString("}\n")
}

func (g *Generator) generateStatusPageHCL(sb *strings.Builder, sp client.StatusPage) {
	name := g.terraformName(sp.Name)

	sb.WriteString(fmt.Sprintf("resource \"hyperping_statuspage\" \"%s\" {\n", name))
	sb.WriteString(fmt.Sprintf("  name             = \"%s\"\n", escapeHCL(sp.Name)))
	sb.WriteString(fmt.Sprintf("  hosted_subdomain = \"%s\"\n", escapeHCL(sp.HostedSubdomain)))

	if sp.Hostname != nil && *sp.Hostname != "" {
		sb.WriteString(fmt.Sprintf("  hostname = \"%s\"\n", escapeHCL(*sp.Hostname)))
	}

	// Settings block
	sb.WriteString("\n  settings = {\n")
	sb.WriteString(fmt.Sprintf("    name      = \"%s\"\n", escapeHCL(sp.Name)))

	if len(sp.Settings.Languages) > 0 {
		sb.WriteString(fmt.Sprintf("    languages = %s\n", formatStringList(sp.Settings.Languages)))
	} else {
		sb.WriteString("    languages = [\"en\"]\n")
	}

	if sp.Settings.Theme != "" && sp.Settings.Theme != "system" {
		sb.WriteString(fmt.Sprintf("    theme = \"%s\"\n", sp.Settings.Theme))
	}

	if sp.Settings.Font != "" && sp.Settings.Font != "Inter" {
		sb.WriteString(fmt.Sprintf("    font = \"%s\"\n", sp.Settings.Font))
	}

	if sp.Settings.AccentColor != "" && sp.Settings.AccentColor != "#36b27e" {
		sb.WriteString(fmt.Sprintf("    accent_color = \"%s\"\n", sp.Settings.AccentColor))
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

	sb.WriteString(fmt.Sprintf("resource \"hyperping_incident\" \"%s\" {\n", name))
	sb.WriteString(fmt.Sprintf("  title = \"%s\"\n", escapeHCL(i.Title.En)))

	if i.Text.En != "" {
		sb.WriteString(fmt.Sprintf("  text = \"%s\"\n", escapeHCL(i.Text.En)))
	}

	if i.Type != "" && i.Type != "incident" {
		sb.WriteString(fmt.Sprintf("  type = \"%s\"\n", i.Type))
	}

	if len(i.StatusPages) > 0 {
		sb.WriteString(fmt.Sprintf("  status_pages = %s\n", formatStringList(i.StatusPages)))
	}

	if len(i.AffectedComponents) > 0 {
		sb.WriteString(fmt.Sprintf("  affected_components = %s\n", formatStringList(i.AffectedComponents)))
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

	sb.WriteString(fmt.Sprintf("resource \"hyperping_maintenance\" \"%s\" {\n", name))
	sb.WriteString(fmt.Sprintf("  title = \"%s\"\n", escapeHCL(titleText)))

	if m.Text.En != "" {
		sb.WriteString(fmt.Sprintf("  text = \"%s\"\n", escapeHCL(m.Text.En)))
	}

	if m.StartDate != nil {
		sb.WriteString(fmt.Sprintf("  start_date = \"%s\"\n", *m.StartDate))
	}
	if m.EndDate != nil {
		sb.WriteString(fmt.Sprintf("  end_date   = \"%s\"\n", *m.EndDate))
	}

	if len(m.StatusPages) > 0 {
		sb.WriteString(fmt.Sprintf("  status_pages = %s\n", formatStringList(m.StatusPages)))
	}

	sb.WriteString("}\n")
}

func (g *Generator) generateOutageHCL(sb *strings.Builder, o client.Outage) {
	name := g.terraformName(o.Monitor.Name)

	sb.WriteString(fmt.Sprintf("resource \"hyperping_outage\" \"%s\" {\n", name))
	sb.WriteString(fmt.Sprintf("  monitor_uuid = \"%s\"\n", o.Monitor.UUID))

	if o.Description != "" {
		sb.WriteString(fmt.Sprintf("  # description = \"%s\"\n", escapeHCL(o.Description)))
	}

	// Note: Most outage fields are read-only/computed
	sb.WriteString("  # Note: Outages are mostly read-only. Review fields after import.\n")

	sb.WriteString("}\n")
}
