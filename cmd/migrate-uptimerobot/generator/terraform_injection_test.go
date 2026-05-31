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

	// The escaped forms MUST appear: ${...} must become $${...} and %{...}
	// must become %%{...} so HCL treats them as literal text.
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

	// No unescaped template start ("${" or "%{") may appear. We detect this by
	// scanning for the literal two-byte starts that lack the leading sigil
	// double. An unescaped ${...} appears as "${" not preceded by "$"; an
	// unescaped %{...} appears as "%{" not preceded by "%".
	if hasUnescapedTemplate(out, "${", '$') {
		t.Errorf("generated HCL contains unescaped ${ template start\n---\n%s\n---", out)
	}
	if hasUnescapedTemplate(out, "%{", '%') {
		t.Errorf("generated HCL contains unescaped %%{ template start\n---\n%s\n---", out)
	}
}

// hasUnescapedTemplate returns true when needle ("${" or "%{") appears in s
// without the leading escape byte (a duplicate of the sigil) directly in front
// of it. This is the same check Terraform's template lexer effectively does.
func hasUnescapedTemplate(s, needle string, escape byte) bool {
	from := 0
	for {
		idx := strings.Index(s[from:], needle)
		if idx < 0 {
			return false
		}
		abs := from + idx
		if abs == 0 || s[abs-1] != escape {
			return true
		}
		from = abs + len(needle)
	}
}
