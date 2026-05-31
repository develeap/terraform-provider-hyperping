// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"strings"
	"testing"

	hyperping "github.com/develeap/hyperping-go"
)

// TestGenerateHCL_TemplateInjection verifies that attacker-controlled string
// fields in a Hyperping resource cannot inject HCL template-interpolation
// sequences (${...} and %{...}) into the generated Terraform configuration.
//
// If the API ever returns a monitor name like `${file("/etc/passwd")}` (or if
// an operator has crafted such a name via the UI), the generated .tf must
// neutralize the sigil so Terraform does not evaluate it at plan time. This
// test locks in the use of migrate.QuoteHCL in hcl.go; swapping in fmt's %q
// verb would silently regress and this test would catch it.
func TestGenerateHCL_TemplateInjection(t *testing.T) {
	g := &Generator{}
	keyword := `%{for x in y}`
	policy := hyperping.EscalationPolicyRef{UUID: "esc_safe_123"}
	data := &ResourceData{
		Monitors: []hyperping.Monitor{
			{
				UUID:             "mon_evil",
				Name:             `${file("/etc/passwd")}`,
				URL:              `https://example.com/${file("/etc/hosts")}`,
				Protocol:         "http",
				HTTPMethod:       "POST",
				CheckFrequency:   30,
				Regions:          []string{`virginia-${env.LEAK}`},
				FollowRedirects:  true,
				RequiredKeyword:  &keyword,
				EscalationPolicy: &policy,
				RequestHeaders: []hyperping.RequestHeader{
					{Name: `X-${env.SECRET}`, Value: `%{for x in y}`},
				},
				RequestBody: `${env.SECRET}`,
			},
		},
		Healthchecks: []hyperping.Healthcheck{
			{
				UUID: "hc_evil",
				Name: `${env.SECRET}`,
			},
		},
	}

	var sb strings.Builder
	g.generateHCL(&sb, data)
	out := sb.String()

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

	if hasUnescapedTemplateInQuotedHCL(out, "${", '$') {
		t.Errorf("generated HCL contains unescaped ${ template start in a quoted string\n---\n%s\n---", out)
	}
	if hasUnescapedTemplateInQuotedHCL(out, "%{", '%') {
		t.Errorf("generated HCL contains unescaped %%{ template start in a quoted string\n---\n%s\n---", out)
	}
}

// TestGenerateScript_TemplateInjectionInHCLPath is a defensive check that the
// HCL output (via generateHCL) keeps escaped sigils for malicious monitor
// names. The shell script path uses %q for UUIDs only (validated separately);
// names flow through terraformName() which strips non-identifier characters,
// so they cannot leak template sigils into the script.
func TestGenerateScript_NameSanitization(t *testing.T) {
	g := &Generator{}
	data := &ResourceData{
		Monitors: []hyperping.Monitor{
			{
				UUID: "mon_safe_123",
				Name: `${file("/etc/passwd")}`,
			},
		},
	}
	out := g.generateScript(data)
	// terraformName must have stripped the template sigil from the resource
	// address; the UUID is plain alphanumeric, so the script must not
	// contain any unescaped template sequences.
	if strings.Contains(out, "${file(") {
		t.Errorf("script leaked unescaped template sigil from monitor name:\n%s", out)
	}
}

// hasUnescapedTemplateInQuotedHCL reports true when needle ("${" or "%{")
// appears inside a double-quoted HCL literal in s without being preceded by
// the matching escape byte. Comments (`# ...`) are not scanned because they
// do not undergo HCL template evaluation.
func hasUnescapedTemplateInQuotedHCL(s, needle string, escape byte) bool {
	inQuote := false
	prevByteEscape := byte(0)
	for i := 0; i < len(s); i++ {
		c := s[i]
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
