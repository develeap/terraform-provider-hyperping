// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
)

// goldenResult is a deliberately small, deterministic conversion result that
// touches one of every code path (monitor with warning, healthcheck with
// warning, skipped resource). No map-iteration paths in the generator so the
// output is stable.
func goldenResult() *converter.ConversionResult {
	return &converter.ConversionResult{
		Monitors: []converter.HyperpingMonitor{
			{
				ResourceName:       "api_health",
				Name:               "API Health",
				URL:                "https://api.example.com/health",
				Protocol:           "http",
				HTTPMethod:         "GET",
				CheckFrequency:     300,
				ExpectedStatusCode: "2xx",
				FollowRedirects:    true,
				Regions:            []string{"london", "virginia", "singapore"},
				OriginalID:         101,
				Warnings:           []string{"Check frequency adjusted from 250s to 300s (nearest allowed value)"},
			},
			{
				ResourceName:    "db_port",
				Name:            "Database Port",
				URL:             "https://db.example.com",
				Protocol:        "port",
				CheckFrequency:  60,
				Port:            5432,
				FollowRedirects: false,
				Regions:         []string{"virginia"},
				OriginalID:      102,
			},
		},
		Healthchecks: []converter.HyperpingHealthcheck{
			{
				ResourceName:     "cron_heartbeat",
				Name:             "Cron Heartbeat",
				PeriodValue:      1,
				PeriodType:       "hours",
				GracePeriodValue: 1,
				GracePeriodType:  "hours",
				OriginalID:       201,
				Warnings:         []string{"Heartbeat monitor converted to healthcheck. Update your script to ping the new URL (see manual-steps.md)"},
			},
		},
		Skipped: []converter.SkippedMonitor{
			{ID: 999, Name: "Unsupported Foo", Type: 99, Reason: "unsupported monitor type: 99"},
		},
		ContactsMap: map[string][]string{},
	}
}

func TestGenerateTerraform_Golden(t *testing.T) {
	got := GenerateTerraform(goldenResult())
	goldenAssert(t, "hyperping.tf.golden", got)
}

func TestGenerateTerraform_EmptyResult(t *testing.T) {
	got := GenerateTerraform(&converter.ConversionResult{ContactsMap: map[string][]string{}})

	// Header + provider block must always be present.
	for _, want := range []string{
		"# Terraform configuration generated from UptimeRobot migration",
		`terraform {`,
		`provider "hyperping"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("empty result missing %q", want)
		}
	}
	// No resources or skipped section.
	for _, unwanted := range []string{
		`resource "hyperping_monitor"`,
		`resource "hyperping_healthcheck"`,
		"# Skipped Resources",
		"# Escalation Policy Configuration",
	} {
		if strings.Contains(got, unwanted) {
			t.Errorf("empty result should not contain %q", unwanted)
		}
	}
}

func TestGenerateTerraform_HealthcheckOutput(t *testing.T) {
	r := &converter.ConversionResult{
		Healthchecks: []converter.HyperpingHealthcheck{
			{ResourceName: "hb1", Name: "HB", PeriodValue: 1, PeriodType: "hours", GracePeriodValue: 1, GracePeriodType: "hours"},
		},
	}
	got := GenerateTerraform(r)
	if !strings.Contains(got, `output "hb1_ping_url"`) {
		t.Errorf("expected ping_url output for healthcheck, got:\n%s", got)
	}
	if !strings.Contains(got, "sensitive   = true") {
		t.Error("expected sensitive flag on ping_url output")
	}
}

func TestGenerateTerraform_FollowRedirectsOnlyForHTTP(t *testing.T) {
	r := &converter.ConversionResult{
		Monitors: []converter.HyperpingMonitor{
			{ResourceName: "p", Name: "P", URL: "host", Protocol: "icmp", CheckFrequency: 60, FollowRedirects: true},
		},
	}
	got := GenerateTerraform(r)
	if strings.Contains(got, "follow_redirects") {
		t.Errorf("follow_redirects should be omitted for non-HTTP protocols, got:\n%s", got)
	}
}
