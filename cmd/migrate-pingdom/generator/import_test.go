// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/converter"
	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-pingdom/pingdom"
)

func makeChecks() ([]pingdom.Check, []converter.ConversionResult) {
	checks := []pingdom.Check{
		{
			ID:       1,
			Name:     "API",
			Type:     "http",
			Hostname: "api.example.com",
			URL:      "/health",
			Tags:     []pingdom.Tag{{Name: "production"}, {Name: "api"}},
		},
		{
			ID:       2,
			Name:     "DNS",
			Type:     "dns",
			Hostname: "example.com",
		},
		{
			ID:       3,
			Name:     "Web",
			Type:     "http",
			Hostname: "site.example.com",
			Tags:     []pingdom.Tag{{Name: "production"}, {Name: "web"}},
		},
	}
	conv := converter.NewCheckConverter()
	results := make([]converter.ConversionResult, len(checks))
	for i, c := range checks {
		results[i] = conv.Convert(c)
	}
	return checks, results
}

// TestGenerateImportScript_Golden snapshots the full script for a representative
// input mix (one created, one unsupported, one supported-but-not-created). The
// script is user-facing and shells get sensitive to subtle shape regressions
// (echo placement, semicolons, comment vs command), so we pin the entire output.
func TestGenerateImportScript_Golden(t *testing.T) {
	checks, results := makeChecks()
	created := map[int]string{
		1: "mon_aaaa",
		// check 2 is unsupported, check 3 not yet created
	}
	got := NewImportGenerator("").GenerateImportScript(checks, results, created)
	goldenAssert(t, "import.sh.golden", got)
}

func TestGenerateImportScript_Prefix(t *testing.T) {
	checks, results := makeChecks()
	created := map[int]string{1: "mon_aaaa"}
	out := NewImportGenerator("pd_").GenerateImportScript(checks, results, created)
	if !strings.Contains(out, "terraform import hyperping_monitor.pd_") {
		t.Errorf("expected prefix in import target:\n%s", out)
	}
}

func TestGenerateImportCommands_Shape(t *testing.T) {
	checks, results := makeChecks()
	created := map[int]string{1: "mon_aaaa"}
	out := NewImportGenerator("").GenerateImportCommands(checks, results, created)

	if strings.HasPrefix(out, "#!/bin/bash") {
		t.Errorf("import commands form should not contain shebang:\n%s", out)
	}
	if !strings.Contains(out, "terraform import hyperping_monitor.") {
		t.Errorf("expected import command:\n%s", out)
	}
	if strings.Contains(out, "set -e") {
		t.Errorf("set -e should not be in commands form:\n%s", out)
	}
	if strings.Contains(out, "Pingdom Check 3") {
		t.Errorf("uncreated check 3 should not appear:\n%s", out)
	}
	if strings.Contains(out, "Pingdom Check 2") {
		t.Errorf("unsupported check 2 should not appear:\n%s", out)
	}
}

func TestGenerateImportScript_NoResources(t *testing.T) {
	out := NewImportGenerator("").GenerateImportScript(nil, nil, nil)
	if !strings.Contains(out, "Imported 0 resources") {
		t.Errorf("expected zero count for empty input:\n%s", out)
	}
}
