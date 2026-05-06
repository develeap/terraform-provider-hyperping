// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package report

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

func sampleInputs() ([]pingdom.Check, []converter.ConversionResult) {
	checks := []pingdom.Check{
		{ID: 1, Name: "API", Type: "http", Hostname: "api.example.com", Tags: []pingdom.Tag{{Name: "production"}, {Name: "api"}}},
		{ID: 2, Name: "DNS", Type: "dns", Hostname: "example.com"},
		{ID: 3, Name: "UDP", Type: "udp", Hostname: "u.example.com"},
		{ID: 4, Name: "TX", Type: "transaction", Hostname: "x.example.com"},
		{ID: 5, Name: "Mail", Type: "smtp", Hostname: "mail.example.com"},
		{ID: 6, Name: "Mystery", Type: "weirdo", Hostname: "x"},
	}
	conv := converter.NewCheckConverter()
	results := make([]converter.ConversionResult, len(checks))
	for i, c := range checks {
		results[i] = conv.Convert(c)
	}
	return checks, results
}

func TestGenerateReport_Counts(t *testing.T) {
	checks, results := sampleInputs()

	r := NewReporter().GenerateReport(checks, results)

	if r.TotalChecks != 6 {
		t.Errorf("TotalChecks = %d, want 6", r.TotalChecks)
	}
	if r.SupportedChecks != 2 { // http + smtp
		t.Errorf("SupportedChecks = %d, want 2", r.SupportedChecks)
	}
	if r.UnsupportedChecks != 4 {
		t.Errorf("UnsupportedChecks = %d, want 4", r.UnsupportedChecks)
	}
	if r.ChecksByType["http"] != 1 || r.ChecksByType["dns"] != 1 || r.ChecksByType["smtp"] != 1 {
		t.Errorf("ChecksByType = %v", r.ChecksByType)
	}
	if r.UnsupportedTypes["dns"] != 1 || r.UnsupportedTypes["udp"] != 1 || r.UnsupportedTypes["transaction"] != 1 || r.UnsupportedTypes["weirdo"] != 1 {
		t.Errorf("UnsupportedTypes = %v", r.UnsupportedTypes)
	}
	if len(r.ManualSteps) != 4 {
		t.Errorf("ManualSteps = %d, want 4", len(r.ManualSteps))
	}
	// SMTP supported with notes -> warning expected
	if len(r.Warnings) == 0 {
		t.Error("expected at least one warning for SMTP note")
	}
}

func TestGenerateManualStep_ByType(t *testing.T) {
	cases := []struct {
		checkType    string
		descContains string
		actContains  string
	}{
		{"dns", "DNS checks are not directly supported", "DNS-over-HTTPS"},
		{"udp", "UDP checks are not supported", "TCP alternative"},
		{"transaction", "Transaction/browser checks", "Playwright/Selenium"},
		{"weirdo", "is not supported", "Manual review"},
	}
	r := NewReporter()
	for _, tt := range cases {
		t.Run(tt.checkType, func(t *testing.T) {
			step := r.generateManualStep(
				pingdom.Check{ID: 1, Name: "x", Type: tt.checkType},
				converter.ConversionResult{},
			)
			if !strings.Contains(step.Description, tt.descContains) {
				t.Errorf("Description = %q, want substring %q", step.Description, tt.descContains)
			}
			if !strings.Contains(step.Action, tt.actContains) {
				t.Errorf("Action = %q, want substring %q", step.Action, tt.actContains)
			}
		})
	}
}

func TestGenerateJSONReport_RoundTrip(t *testing.T) {
	checks, results := sampleInputs()
	rep := NewReporter()
	report := rep.GenerateReport(checks, results)

	out, err := rep.GenerateJSONReport(report)
	if err != nil {
		t.Fatalf("GenerateJSONReport error: %v", err)
	}

	var got MigrationReport
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.TotalChecks != report.TotalChecks {
		t.Errorf("TotalChecks roundtrip: got %d, want %d", got.TotalChecks, report.TotalChecks)
	}
	if len(got.ManualSteps) != len(report.ManualSteps) {
		t.Errorf("ManualSteps roundtrip: got %d, want %d", len(got.ManualSteps), len(report.ManualSteps))
	}
}

func TestGenerateTextReport_Sections(t *testing.T) {
	checks, results := sampleInputs()
	rep := NewReporter()
	report := rep.GenerateReport(checks, results)

	out := rep.GenerateTextReport(report)
	for _, want := range []string{
		"Pingdom to Hyperping Migration Report",
		"Total Checks:",
		"Checks by Type",
		"Unsupported Check Types",
		"Warnings",
		"Manual Steps Required",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("text report missing %q", want)
		}
	}
}

func TestGenerateTextReport_NoWarningsNoSteps(t *testing.T) {
	checks := []pingdom.Check{
		{ID: 1, Name: "Plain HTTP", Type: "http", Hostname: "h"},
	}
	results := []converter.ConversionResult{
		converter.NewCheckConverter().Convert(checks[0]),
	}
	rep := NewReporter()
	report := rep.GenerateReport(checks, results)

	out := rep.GenerateTextReport(report)
	if strings.Contains(out, "Warnings\n--------") {
		t.Error("did not expect Warnings section")
	}
	if strings.Contains(out, "Manual Steps Required") {
		t.Error("did not expect Manual Steps section")
	}
}

func TestGenerateManualStepsMarkdown_Empty(t *testing.T) {
	out := NewReporter().GenerateManualStepsMarkdown(&MigrationReport{})
	if !strings.Contains(out, "No manual steps required") {
		t.Errorf("expected empty-state message, got:\n%s", out)
	}
}

func TestGenerateManualStepsMarkdown_WithSteps(t *testing.T) {
	checks, results := sampleInputs()
	rep := NewReporter()
	report := rep.GenerateReport(checks, results)

	out := rep.GenerateManualStepsMarkdown(report)
	for _, want := range []string{
		"# Manual Migration Steps",
		"## 1.",
		"**Type:**",
		"**Issue:**",
		"**Action Required:**",
		"## Additional Resources",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("markdown missing %q", want)
		}
	}
}
