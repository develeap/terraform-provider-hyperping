// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"strings"
	"testing"

	hyperping "github.com/develeap/hyperping-go"
)

// TestGenerateScript_RejectsMaliciousUUID is defense-in-depth: the Hyperping
// API is the source of truth for resource UUIDs, but if a future regression
// (or partner-API forwarding) ever surfaces an attacker-influenced value, the
// generated shell script must refuse to interpolate it. The previous
// implementation emitted UUIDs through fmt's %q verb, which does NOT escape
// bash metacharacters ($, `, ;), so a UUID-shaped string like
// `'; rm -rf $HOME; '` would survive into the script and execute when run.
func TestGenerateScript_RejectsMaliciousUUID(t *testing.T) {
	g := &Generator{}
	data := &ResourceData{
		Monitors: []hyperping.Monitor{
			{UUID: `$(rm -rf $HOME)`, Name: "evil_one"},
			{UUID: "`whoami`", Name: "evil_two"},
			{UUID: "'; rm -rf $HOME; '", Name: "evil_three"},
			{UUID: "mon_safe_123", Name: "good"},
		},
	}
	out := g.generateScript(data)

	// Malicious payloads must NOT survive into the script.
	forbidden := []string{
		"$(rm",
		"`whoami`",
		"'; rm",
	}
	for _, f := range forbidden {
		if strings.Contains(out, f) {
			t.Errorf("script contains forbidden substring %q\n---\n%s\n---", f, out)
		}
	}

	// The safe UUID must round-trip unchanged.
	if !strings.Contains(out, "mon_safe_123") {
		t.Errorf("safe UUID was rejected; script:\n%s", out)
	}
}

// TestGenerateImports_RejectsMaliciousUUID exercises the same defense for the
// terraform-import-command output path (Generator.generateImports). HCL would
// reject these strings on parse, but operators sometimes pipe the output into
// other tooling, so defense in depth applies.
func TestGenerateImports_RejectsMaliciousUUID(t *testing.T) {
	g := &Generator{}
	data := &ResourceData{
		Healthchecks: []hyperping.Healthcheck{
			{UUID: "$(rm -rf /)", Name: "evil"},
			{UUID: "hc_safe_123", Name: "good"},
		},
	}
	var sb strings.Builder
	g.generateImports(&sb, data)
	out := sb.String()

	if strings.Contains(out, "$(rm") {
		t.Errorf("imports output contains malicious substitution\n---\n%s\n---", out)
	}
	if !strings.Contains(out, "hc_safe_123") {
		t.Errorf("safe UUID was rejected; output:\n%s", out)
	}
}
