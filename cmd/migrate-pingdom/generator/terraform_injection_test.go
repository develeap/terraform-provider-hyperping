// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	hyperping "github.com/develeap/hyperping-go"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

// TestGenerateHCL_TemplateInjection verifies that attacker-controlled string
// fields in a Pingdom check cannot inject HCL template-interpolation sequences
// (${...} and %{...}) into the generated Terraform configuration.
//
// A malicious Pingdom check whose name contains `${file("/etc/passwd")}` must
// be emitted with the leading dollar/percent doubled so Terraform treats it
// as literal text rather than evaluating it during plan/apply. This locks in
// the use of migrate.QuoteHCL inside generator/terraform.go; replacing those
// calls with fmt's %q verb (which does not escape template sigils) would
// regress and this test would catch it.
func TestGenerateHCL_TemplateInjection(t *testing.T) {
	checks := []pingdom.Check{
		{
			ID:   42,
			Name: `${file("/etc/passwd")}`,
			Type: "http",
			URL:  `https://example.com/${file("/etc/hosts")}`,
		},
	}

	requiredKeyword := `%{for x in y}`
	requestBody := `${env.SECRET}`
	followRedirects := false
	results := []converter.ConversionResult{
		{
			Supported: true,
			Monitor: &hyperping.CreateMonitorRequest{
				Name:            `${file("/etc/passwd")}`,
				URL:             `https://example.com/${file("/etc/hosts")}`,
				Protocol:        "http",
				HTTPMethod:      "POST",
				CheckFrequency:  60,
				Regions:         []string{`virginia-${env.LEAK}`},
				FollowRedirects: &followRedirects,
				RequiredKeyword: &requiredKeyword,
				RequestBody:     &requestBody,
				RequestHeaders: []hyperping.RequestHeader{
					{Name: `X-${env.SECRET}`, Value: `%{for x in y}`},
				},
			},
		},
	}

	g := NewTerraformGenerator("")
	out := g.GenerateHCL(checks, results)

	// Escaped forms MUST appear.
	expected := []string{
		`$${file(`,
		`$${env.LEAK}`,
		`$${env.SECRET}`,
		`%%{for x in y}`,
	}
	for _, e := range expected {
		if !strings.Contains(out, e) {
			t.Errorf("generated HCL missing escaped sequence %q\n---\n%s\n---", e, out)
		}
	}

	// No unescaped template starts inside attribute string literals may remain.
	if hasUnescapedTemplateInQuotes(out, "${", '$') {
		t.Errorf("generated HCL contains unescaped ${ template start in a quoted string\n---\n%s\n---", out)
	}
	if hasUnescapedTemplateInQuotes(out, "%{", '%') {
		t.Errorf("generated HCL contains unescaped %%{ template start in a quoted string\n---\n%s\n---", out)
	}
}

// hasUnescapedTemplateInQuotes reports true when needle ("${" or "%{") appears
// inside a double-quoted HCL literal in s without being preceded by the
// matching escape byte. We restrict the search to double-quoted spans because
// the generator legitimately emits `# Original Name: ${...}` comments with the
// raw user-supplied name (comments do not undergo HCL template evaluation).
func hasUnescapedTemplateInQuotes(s, needle string, escape byte) bool {
	inQuote := false
	prevByteEscape := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Track double-quote spans, but only toggle on quotes that are not
		// themselves escaped by a backslash (so `\"` inside the literal does
		// not flip the inQuote state).
		if c == '"' && prevByteEscape != '\\' {
			inQuote = !inQuote
		}
		if inQuote && i+len(needle) <= len(s) && s[i:i+len(needle)] == needle {
			if i == 0 || s[i-1] != escape {
				return true
			}
		}
		prevByteEscape = c
	}
	return false
}
