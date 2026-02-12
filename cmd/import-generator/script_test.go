// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

func TestGenerateScript_EmptyData(t *testing.T) {
	gen := &Generator{
		prefix: "",
	}

	data := &ResourceData{}
	script := gen.generateScript(data)

	// Check for shebang
	if !strings.HasPrefix(script, "#!/bin/bash") {
		t.Error("Script should start with shebang")
	}

	// Check for error handling
	if !strings.Contains(script, "set -e") {
		t.Error("Script should have 'set -e' for error handling")
	}

	if !strings.Contains(script, "set -u") {
		t.Error("Script should have 'set -u' for undefined variable checking")
	}

	// Check for terraform check
	if !strings.Contains(script, "command -v terraform") {
		t.Error("Script should check for terraform command")
	}

	// Check for import function
	if !strings.Contains(script, "import_resource()") {
		t.Error("Script should define import_resource function")
	}

	// Check for summary
	if !strings.Contains(script, "Import Summary") {
		t.Error("Script should have import summary")
	}
}

func TestGenerateScript_Monitors(t *testing.T) {
	gen := &Generator{
		prefix: "prod_",
	}

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_abc123", Name: "API Monitor"},
			{UUID: "mon_def456", Name: "Web Monitor"},
		},
	}

	script := gen.generateScript(data)

	// Check for monitor imports
	if !strings.Contains(script, "import_resource \"hyperping_monitor.prod_api_monitor\" \"mon_abc123\"") {
		t.Error("Script should contain API monitor import with prefix")
	}

	if !strings.Contains(script, "import_resource \"hyperping_monitor.prod_web_monitor\" \"mon_def456\"") {
		t.Error("Script should contain Web monitor import with prefix")
	}

	// Check for monitors section header
	if !strings.Contains(script, "# Monitors") {
		t.Error("Script should have Monitors section header")
	}
}

func TestGenerateScript_AllResourceTypes(t *testing.T) {
	gen := &Generator{
		prefix: "",
	}

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Monitor"},
		},
		Healthchecks: []client.Healthcheck{
			{UUID: "hc_123", Name: "Healthcheck"},
		},
		StatusPages: []client.StatusPage{
			{UUID: "sp_123", Name: "Status Page"},
		},
		Incidents: []client.Incident{
			{UUID: "inc_123", Title: client.LocalizedText{En: "Incident"}},
		},
		Maintenance: []client.Maintenance{
			{UUID: "maint_123", Title: client.LocalizedText{En: "Maintenance"}, Name: "Maintenance"},
		},
		Outages: []client.Outage{
			{UUID: "outage_123", Monitor: client.MonitorReference{Name: "Outage Monitor"}},
		},
	}

	script := gen.generateScript(data)

	// Check for all resource type headers
	expectedHeaders := []string{
		"# Monitors",
		"# Healthchecks",
		"# Status Pages",
		"# Incidents",
		"# Maintenance Windows",
		"# Outages",
	}

	for _, header := range expectedHeaders {
		if !strings.Contains(script, header) {
			t.Errorf("Script should contain %q header", header)
		}
	}

	// Check for all import commands
	expectedImports := []string{
		"hyperping_monitor.monitor",
		"hyperping_healthcheck.healthcheck",
		"hyperping_statuspage.status_page",
		"hyperping_incident.incident",
		"hyperping_maintenance.maintenance",
		"hyperping_outage.outage_monitor",
	}

	for _, importCmd := range expectedImports {
		if !strings.Contains(script, importCmd) {
			t.Errorf("Script should contain import for %q", importCmd)
		}
	}
}

func TestGenerateScript_CountersAndSummary(t *testing.T) {
	gen := &Generator{
		prefix: "",
	}

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Monitor"},
		},
	}

	script := gen.generateScript(data)

	// Check for counter initialization
	expectedCounters := []string{
		"TOTAL=0",
		"SUCCESS=0",
		"FAILED=0",
	}

	for _, counter := range expectedCounters {
		if !strings.Contains(script, counter) {
			t.Errorf("Script should initialize counter: %s", counter)
		}
	}

	// Check for summary output
	expectedSummary := []string{
		"echo \"Total:   $TOTAL\"",
		"echo \"Success: $SUCCESS\"",
		"echo \"Failed:  $FAILED\"",
	}

	for _, summary := range expectedSummary {
		if !strings.Contains(script, summary) {
			t.Errorf("Script should contain summary line: %s", summary)
		}
	}

	// Check for success/failure handling
	if !strings.Contains(script, "if [ $FAILED -gt 0 ]; then") {
		t.Error("Script should handle failure case")
	}

	if !strings.Contains(script, "All imports completed successfully!") {
		t.Error("Script should have success message")
	}
}

func TestGenerateScript_ErrorHandling(t *testing.T) {
	gen := &Generator{
		prefix: "",
	}

	data := &ResourceData{
		Monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Monitor"},
		},
	}

	script := gen.generateScript(data)

	// Check for function error handling
	if !strings.Contains(script, "if terraform import") {
		t.Error("Script should check terraform import result")
	}

	if !strings.Contains(script, "SUCCESS=$((SUCCESS + 1))") {
		t.Error("Script should increment SUCCESS counter on success")
	}

	if !strings.Contains(script, "FAILED=$((FAILED + 1))") {
		t.Error("Script should increment FAILED counter on failure")
	}

	if !strings.Contains(script, "echo \"  ✓ Success\"") {
		t.Error("Script should output success indicator")
	}

	if !strings.Contains(script, "echo \"  ✗ Failed\"") {
		t.Error("Script should output failure indicator")
	}
}

func TestGenerateScript_UsageInstructions(t *testing.T) {
	gen := &Generator{
		prefix: "",
	}

	data := &ResourceData{}
	script := gen.generateScript(data)

	// Check for usage instructions in comments
	expectedInstructions := []string{
		"# Usage:",
		"#   chmod +x import.sh",
		"#   ./import.sh",
		"# Requirements:",
		"#   - Terraform must be installed",
	}

	for _, instruction := range expectedInstructions {
		if !strings.Contains(script, instruction) {
			t.Errorf("Script should contain instruction: %s", instruction)
		}
	}
}
