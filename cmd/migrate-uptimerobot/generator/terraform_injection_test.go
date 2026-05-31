// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
)

// TestGenerateTerraform_HCLTemplateInjection verifies that attacker-controlled
// string fields in the source monitor cannot inject HCL template-interpolation
// sequences (${...} and %{...}) into the generated Terraform configuration.
//
// A malicious UptimeRobot monitor name such as `${file("/etc/passwd")}` must be
// emitted with the leading dollar/percent doubled so Terraform treats it as a
// literal rather than evaluating it during plan/apply.
func TestGenerateTerraform_HCLTemplateInjection(t *testing.T) {
	result := &converter.ConversionResult{
		Monitors: []converter.HyperpingMonitor{
			{
				ResourceName:    "evil",
				Name:            `${file("/etc/passwd")}`,
				URL:             `https://example.com/${file("/etc/hosts")}`,
				Protocol:        "http",
				HTTPMethod:      "GET",
				CheckFrequency:  60,
				RequiredKeyword: `%{for x in y}`,
				FollowRedirects: true,
			},
		},
		Healthchecks: []converter.HyperpingHealthcheck{
			{
				ResourceName:     "evil_hc",
				Name:             `${env.SECRET}`,
				PeriodValue:      5,
				PeriodType:       "minutes",
				GracePeriodValue: 1,
				GracePeriodType:  "minutes",
			},
		},
	}

	out := GenerateTerraform(result)

	// Raw template sequences must NOT appear unescaped anywhere in the output.
	forbidden := []string{
		`${file(`,
		`${env.SECRET}`,
		`%{for x in y}`,
	}
	for _, f := range forbidden {
		if strings.Contains(out, f) {
			t.Errorf("generated HCL contains unescaped template sequence %q\n---\n%s\n---", f, out)
		}
	}

	// The escaped forms MUST appear.
	expected := []string{
		`$${file(`,
		`$${env.SECRET}`,
		`%%{for x in y}`,
	}
	for _, e := range expected {
		if !strings.Contains(out, e) {
			t.Errorf("generated HCL missing escaped sequence %q\n---\n%s\n---", e, out)
		}
	}
}
