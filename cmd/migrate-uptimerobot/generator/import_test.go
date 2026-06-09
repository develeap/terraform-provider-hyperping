// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package generator

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/cmd/migrate-uptimerobot/converter"
)

func TestGenerateImportScript_Golden(t *testing.T) {
	got := GenerateImportScript(goldenResult())
	goldenAssert(t, "import.sh.golden", got)
}

func TestGenerateImportScript_NoResources(t *testing.T) {
	got := GenerateImportScript(&converter.ConversionResult{})
	for _, want := range []string{
		"#!/bin/bash",
		"set -e",
		"Total resources to import: 0",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("empty script missing %q", want)
		}
	}
	if strings.Contains(got, "terraform import") {
		t.Errorf("empty result should not produce import commands, got:\n%s", got)
	}
}

func TestGenerateImportScript_OnlyHealthchecks(t *testing.T) {
	r := &converter.ConversionResult{
		Healthchecks: []converter.HyperpingHealthcheck{
			{ResourceName: "hb", Name: "H", OriginalID: 1, PeriodValue: 1, PeriodType: "hours"},
		},
	}
	got := GenerateImportScript(r)
	if !strings.Contains(got, "terraform import 'hyperping_healthcheck.hb'") {
		t.Errorf("expected healthcheck import line, got:\n%s", got)
	}
	if strings.Contains(got, "hyperping_monitor.") {
		t.Error("monitor import lines should not appear when no monitors are present")
	}
}
