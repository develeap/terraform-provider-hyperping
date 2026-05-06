// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/uptimerobot"
)

func TestGenerateManualSteps_FullCoverage(t *testing.T) {
	contacts := []uptimerobot.AlertContact{
		{Type: 2, Value: "ops@example.com"},
		{Type: 3, Value: "+15555555555"},
		{Type: 4, Value: "https://hook.example.com"},
		{Type: 11, FriendlyName: "Slack #ops"},
		{Type: 14, FriendlyName: "PagerDuty Primary"},
	}
	got := GenerateManualSteps(goldenResult(), contacts)

	for _, want := range []string{
		"# Manual Migration Steps",
		"## Table of Contents",
		"## Escalation Policy Setup",
		"### Your UptimeRobot Alert Contacts",
		"**Email Contacts:**",
		"ops@example.com",
		"**SMS Contacts:**",
		"+15555555555",
		"**Webhook URLs:**",
		"https://hook.example.com",
		"**Slack Integrations:**",
		"Slack #ops",
		"**PagerDuty Integrations:**",
		"PagerDuty Primary",
		"## Healthcheck Configuration",
		"## Monitor Warnings",
		"## Skipped Resources",
		"Unsupported Foo",
		"## Testing and Validation",
		"## Decommissioning UptimeRobot",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing section/text %q", want)
		}
	}
}

func TestGenerateManualSteps_OmitsEmptySections(t *testing.T) {
	// No healthchecks, no warnings, no skipped → those sections must be absent
	// from the body and from the table of contents.
	r := &converter.ConversionResult{
		Monitors: []converter.HyperpingMonitor{
			{ResourceName: "m", Name: "M", Protocol: "http", CheckFrequency: 60, Warnings: nil},
		},
	}
	got := GenerateManualSteps(r, nil)

	for _, unwanted := range []string{
		"## Healthcheck Configuration",
		"## Monitor Warnings",
		"## Skipped Resources",
		"- [Healthcheck Configuration]",
		"- [Monitor Warnings]",
		"- [Skipped Resources]",
	} {
		if strings.Contains(got, unwanted) {
			t.Errorf("did not expect %q in output", unwanted)
		}
	}
	// Sections that always appear.
	for _, want := range []string{
		"## Escalation Policy Setup",
		"## Testing and Validation",
		"## Decommissioning UptimeRobot",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("expected always-on section %q", want)
		}
	}
}

func TestGenerateManualSteps_NoAlertContactsSubsection(t *testing.T) {
	got := GenerateManualSteps(goldenResult(), nil)
	if strings.Contains(got, "### Your UptimeRobot Alert Contacts") {
		t.Errorf("alert-contacts subsection should be omitted when contacts list is empty")
	}
}

func TestHasWarnings(t *testing.T) {
	tests := []struct {
		name string
		r    *converter.ConversionResult
		want bool
	}{
		{"no warnings", &converter.ConversionResult{}, false},
		{
			name: "monitor warning",
			r: &converter.ConversionResult{
				Monitors: []converter.HyperpingMonitor{{Warnings: []string{"x"}}},
			},
			want: true,
		},
		{
			name: "healthcheck warning",
			r: &converter.ConversionResult{
				Healthchecks: []converter.HyperpingHealthcheck{{Warnings: []string{"y"}}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasWarnings(tt.r); got != tt.want {
				t.Errorf("hasWarnings = %v, want %v", got, tt.want)
			}
		})
	}
}
