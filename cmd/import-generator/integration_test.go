// Copyright (c) 2026 Develeap
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/develeap/terraform-provider-hyperping/internal/client"
)

// TestIntegration_ValidationMode tests the full validation workflow
func TestIntegration_ValidationMode(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_valid123", Name: "Monitor 1"},
			{UUID: "mon_valid456", Name: "Monitor 2"},
		},
		healthchecks: []client.Healthcheck{
			{UUID: "hc_valid123", Name: "Healthcheck 1"},
		},
	}

	gen := &Generator{
		client:    mock,
		resources: []string{"monitors", "healthchecks"},
	}

	result := gen.Validate(context.Background())

	if !result.IsValid() {
		t.Error("Expected validation to pass")
	}

	var buf bytes.Buffer
	result.Print(&buf)

	output := buf.String()
	if !strings.Contains(output, "✓ Monitors: 2 valid ID(s)") {
		t.Errorf("Expected monitors success message, got: %s", output)
	}

	if !strings.Contains(output, "✓ Healthchecks: 1 valid ID(s)") {
		t.Errorf("Expected healthchecks success message, got: %s", output)
	}
}

// TestIntegration_ProgressMode tests progress reporting during fetch
func TestIntegration_ProgressMode(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Monitor"},
		},
		healthchecks: []client.Healthcheck{
			{UUID: "hc_123", Name: "Healthcheck"},
		},
	}

	gen := &Generator{
		client:       mock,
		resources:    []string{"monitors", "healthchecks"},
		showProgress: true,
	}

	// Generate and capture any progress output through stderr
	output, err := gen.Generate(context.Background(), "import")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify output contains expected imports
	if !strings.Contains(output, "terraform import hyperping_monitor.monitor") {
		t.Error("Expected monitor import in output")
	}

	if !strings.Contains(output, "terraform import hyperping_healthcheck.healthcheck") {
		t.Error("Expected healthcheck import in output")
	}
}

// TestIntegration_ScriptFormat tests script generation end-to-end
func TestIntegration_ScriptFormat(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_123", Name: "API Monitor"},
			{UUID: "mon_456", Name: "Web Monitor"},
		},
	}

	gen := &Generator{
		client:    mock,
		resources: []string{"monitors"},
		prefix:    "prod_",
	}

	script, err := gen.Generate(context.Background(), "script")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify script structure
	requiredElements := []string{
		"#!/bin/bash",
		"set -e",
		"import_resource",
		"hyperping_monitor.prod_api_monitor",
		"mon_123",
		"hyperping_monitor.prod_web_monitor",
		"mon_456",
		"Import Summary",
		"TOTAL",
		"SUCCESS",
		"FAILED",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(script, elem) {
			t.Errorf("Script missing required element: %q", elem)
		}
	}
}

// TestIntegration_ErrorRecovery tests continue-on-error behavior
func TestIntegration_ErrorRecovery(t *testing.T) {
	mock := &mockClient{
		monitorsErr: errors.New("monitors API error"),
		healthchecks: []client.Healthcheck{
			{UUID: "hc_123", Name: "Healthcheck"},
		},
		statusPages: []client.StatusPage{
			{UUID: "sp_123", Name: "Status Page"},
		},
	}

	gen := &Generator{
		client:          mock,
		resources:       []string{"monitors", "healthchecks", "statuspages"},
		continueOnError: true,
	}

	output, err := gen.Generate(context.Background(), "import")
	if err != nil {
		t.Fatalf("Generate should not fail with continueOnError: %v", err)
	}

	// Should not contain monitors (errored)
	if strings.Contains(output, "hyperping_monitor") {
		t.Error("Output should not contain monitor imports")
	}

	// Should contain healthchecks and status pages (succeeded)
	if !strings.Contains(output, "hyperping_healthcheck.healthcheck") {
		t.Error("Output should contain healthcheck import")
	}

	if !strings.Contains(output, "hyperping_statuspage.status_page") {
		t.Error("Output should contain status page import")
	}
}

// TestIntegration_CombinedModes tests multiple modes working together
func TestIntegration_CombinedModes(t *testing.T) {
	mock := &mockClient{
		monitors: []client.Monitor{
			{UUID: "mon_123", Name: "Monitor"},
		},
		healthchecksErr: errors.New("healthchecks error"),
	}

	gen := &Generator{
		client:          mock,
		resources:       []string{"monitors", "healthchecks"},
		continueOnError: true,
		showProgress:    true,
	}

	output, err := gen.Generate(context.Background(), "script")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should generate script with only monitors
	if !strings.Contains(output, "#!/bin/bash") {
		t.Error("Expected bash script output")
	}

	if !strings.Contains(output, "hyperping_monitor.monitor") {
		t.Error("Expected monitor in script")
	}

	if strings.Contains(output, "hyperping_healthcheck") {
		t.Error("Should not contain healthcheck (errored)")
	}
}

// TestIntegration_AllFormatsWithAllResources tests all formats with all resource types
func TestIntegration_AllFormatsWithAllResources(t *testing.T) {
	mock := &mockClient{
		monitors:     []client.Monitor{{UUID: "mon_1", Name: "Mon"}},
		healthchecks: []client.Healthcheck{{UUID: "hc_1", Name: "HC"}},
		statusPages:  []client.StatusPage{{UUID: "sp_1", Name: "SP"}},
		incidents:    []client.Incident{{UUID: "inc_1", Title: client.LocalizedText{En: "Inc"}}},
		maintenance:  []client.Maintenance{{UUID: "maint_1", Title: client.LocalizedText{En: "Maint"}, Name: "Maint"}},
		outages:      []client.Outage{{UUID: "outage_1", Monitor: client.MonitorReference{Name: "OutMon"}}},
	}

	gen := &Generator{
		client:    mock,
		resources: []string{"monitors", "healthchecks", "statuspages", "incidents", "maintenance", "outages"},
	}

	formats := []string{"import", "hcl", "both", "script"}
	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			output, err := gen.Generate(context.Background(), format)
			if err != nil {
				t.Fatalf("Generate failed for format %s: %v", format, err)
			}

			if output == "" {
				t.Errorf("Empty output for format %s", format)
			}

			// Basic sanity checks
			switch format {
			case "import":
				if !strings.Contains(output, "terraform import") {
					t.Error("Import format should contain 'terraform import'")
				}
			case "hcl":
				if !strings.Contains(output, "resource \"hyperping_") {
					t.Error("HCL format should contain resource blocks")
				}
			case "both":
				if !strings.Contains(output, "terraform import") || !strings.Contains(output, "resource \"hyperping_") {
					t.Error("Both format should contain import commands and HCL")
				}
			case "script":
				if !strings.Contains(output, "#!/bin/bash") {
					t.Error("Script format should contain shebang")
				}
			}
		})
	}
}
